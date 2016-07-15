package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/unixpickle/sgd"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
)

const (
	FeatureCount  = 26
	HiddenSize    = 100
	BatchSize     = 100
	CostBatchSize = 20
)

func Train(rnnFile, sampleDir string, stepSize float64) {
	var seqFunc *rnn.Bidirectional
	rnnData, err := ioutil.ReadFile(rnnFile)
	if err == nil {
		log.Println("Loaded network from file.")
		seqFunc, err = rnn.DeserializeBidirectional(rnnData)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to deserialize network:", err)
			os.Exit(1)
		}
	} else {
		log.Println("Created network.")
		seqFunc = createNetwork()
	}

	log.Println("Loading samples...")
	samples, err := ReadSamples(sampleDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read samples:", err)
		os.Exit(1)
	}

	log.Println("Training with", samples.Len(), "samples...")

	gradienter := &ctc.RGradienter{
		Learner:        seqFunc,
		SeqFunc:        seqFunc,
		MaxConcurrency: 2,
	}
	var epoch int
	sgd.SGDInteractive(gradienter, samples, stepSize, BatchSize, func() bool {
		cost := ctc.TotalCost(seqFunc, samples, CostBatchSize, 0)
		log.Printf("Epoch %d: cost=%e", epoch, cost)
		epoch++
		return true
	})

	data, err := seqFunc.Serialize()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to serialize:", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(rnnFile, data, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save:", err)
		os.Exit(1)
	}
}

func createNetwork() *rnn.Bidirectional {
	outputNet := neuralnet.Network{
		&neuralnet.DenseLayer{
			InputCount:  HiddenSize * 2,
			OutputCount: len(Labels) + 1,
		},
		&neuralnet.LogSoftmaxLayer{},
	}
	outputNet.Randomize()
	return &rnn.Bidirectional{
		Forward:  &rnn.RNNSeqFunc{Block: rnn.NewLSTM(FeatureCount, HiddenSize)},
		Backward: &rnn.RNNSeqFunc{Block: rnn.NewLSTM(FeatureCount, HiddenSize)},
		Output:   &rnn.NetworkSeqFunc{Network: outputNet},
	}
}
