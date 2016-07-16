package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
)

const (
	FeatureCount  = 26
	HiddenSize    = 100
	BatchSize     = 20
	CostBatchSize = 20

	MaxSubBatch    = 10
	MaxConcurrency = 2

	CrossRatio = 0.3
)

func Train(rnnFile, sampleDir string, stepSize float64) {
	log.Println("Loading samples...")
	samples, err := ReadSamples(sampleDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read samples:", err)
		os.Exit(1)
	}

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
		seqFunc = createNetwork(samples)
	}

	crossLen := int(CrossRatio * float64(samples.Len()))
	log.Println("Using", samples.Len()-crossLen, "training and",
		crossLen, "validation samples...")

	sgd.ShuffleSampleSet(samples)
	validation := samples.Subset(0, crossLen)
	training := samples.Subset(crossLen, samples.Len())

	gradienter := &ctc.RGradienter{
		Learner:        seqFunc,
		SeqFunc:        seqFunc,
		MaxConcurrency: MaxConcurrency,
		MaxSubBatch:    MaxSubBatch,
	}
	var epoch int
	sgd.SGDInteractive(gradienter, training, stepSize, BatchSize, func() bool {
		cost := ctc.TotalCost(seqFunc, training, CostBatchSize, 0)
		crossCost := ctc.TotalCost(seqFunc, validation, CostBatchSize, 0)
		log.Printf("Epoch %d: cost=%e cross=%e", epoch, cost, crossCost)
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

func createNetwork(samples sgd.SampleSet) *rnn.Bidirectional {
	means := make(linalg.Vector, FeatureCount)
	var count float64

	for i := 0; i < samples.Len(); i++ {
		inputSeq := samples.GetSample(i).(ctc.Sample).Input
		for _, vec := range inputSeq {
			means.Add(vec)
			count++
		}
	}
	means.Scale(-1 / count)

	stddevs := make(linalg.Vector, FeatureCount)
	for i := 0; i < samples.Len(); i++ {
		inputSeq := samples.GetSample(i).(ctc.Sample).Input
		for _, vec := range inputSeq {
			for j, v := range vec {
				stddevs[j] += math.Pow(v+means[j], 2)
			}
		}
	}
	stddevs.Scale(1 / count)
	for i, x := range stddevs {
		stddevs[i] = 1 / math.Sqrt(x)
	}

	outputNet := neuralnet.Network{
		&neuralnet.DenseLayer{
			InputCount:  HiddenSize * 2,
			OutputCount: len(Labels) + 1,
		},
		&neuralnet.LogSoftmaxLayer{},
	}
	outputNet.Randomize()

	inputNet := neuralnet.Network{
		&neuralnet.VecRescaleLayer{
			Biases: means,
			Scales: stddevs,
		},
	}
	netBlock := rnn.NewNetworkBlock(inputNet, 0)
	forwardBlock := rnn.StackedBlock{
		netBlock,
		rnn.NewLSTM(FeatureCount, HiddenSize),
	}
	backwardBlock := rnn.StackedBlock{
		netBlock,
		rnn.NewLSTM(FeatureCount, HiddenSize),
	}
	return &rnn.Bidirectional{
		Forward:  &rnn.RNNSeqFunc{Block: forwardBlock},
		Backward: &rnn.RNNSeqFunc{Block: backwardBlock},
		Output:   &rnn.NetworkSeqFunc{Network: outputNet},
	}
}
