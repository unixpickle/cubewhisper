package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/unixpickle/cubewhisper"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
)

const (
	FeatureCount  = 26
	HiddenSize    = 200
	BatchSize     = 20
	CostBatchSize = 20

	MaxSubBatch    = 10
	MaxConcurrency = 2

	CrossRatio = 0.3

	WeightStddev  = 0.05
	InputDropout  = 0.9
	HiddenDropout = 0.5
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

	gradienter := &sgd.Adam{
		Gradienter: &ctc.RGradienter{
			Learner:        seqFunc,
			SeqFunc:        seqFunc,
			MaxConcurrency: MaxConcurrency,
			MaxSubBatch:    MaxSubBatch,
		},
	}

	var epoch int
	toggleDropout(seqFunc, true)
	sgd.SGDInteractive(gradienter, training, stepSize, BatchSize, func() bool {
		toggleDropout(seqFunc, false)
		cost := ctc.TotalCost(seqFunc, training, CostBatchSize, 0)
		crossCost := ctc.TotalCost(seqFunc, validation, CostBatchSize, 0)
		toggleDropout(seqFunc, true)
		log.Printf("Epoch %d: cost=%e cross=%e", epoch, cost, crossCost)
		epoch++
		return true
	})
	toggleDropout(seqFunc, false)

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
		&neuralnet.DropoutLayer{
			KeepProbability: HiddenDropout,
			Training:        false,
		},
		&neuralnet.DenseLayer{
			InputCount:  HiddenSize * 2,
			OutputCount: len(cubewhisper.Labels) + 1,
		},
		&neuralnet.LogSoftmaxLayer{},
	}
	outputNet.Randomize()

	inputNet := neuralnet.Network{
		&neuralnet.VecRescaleLayer{
			Biases: means,
			Scales: stddevs,
		},
		&neuralnet.DropoutLayer{
			KeepProbability: InputDropout,
			Training:        false,
		},
	}
	netBlock := rnn.NewNetworkBlock(inputNet, 0)
	forwardBlock := rnn.StackedBlock{
		netBlock,
		rnn.NewGRU(FeatureCount, HiddenSize),
	}
	backwardBlock := rnn.StackedBlock{
		netBlock,
		rnn.NewGRU(FeatureCount, HiddenSize),
	}
	for _, block := range []rnn.StackedBlock{forwardBlock, backwardBlock} {
		for i, param := range block.Parameters() {
			if i%2 == 0 {
				for i := range param.Vector {
					param.Vector[i] = rand.NormFloat64() * WeightStddev
				}
			}
		}
	}
	return &rnn.Bidirectional{
		Forward:  &rnn.RNNSeqFunc{Block: forwardBlock},
		Backward: &rnn.RNNSeqFunc{Block: backwardBlock},
		Output:   &rnn.NetworkSeqFunc{Network: outputNet},
	}
}

func toggleDropout(bd *rnn.Bidirectional, dropout bool) {
	output := bd.Output.(*rnn.NetworkSeqFunc).Network[0].(*neuralnet.DropoutLayer)
	output.Training = dropout
	inNet := bd.Forward.(*rnn.RNNSeqFunc).Block.(rnn.StackedBlock)[0].(*rnn.NetworkBlock)
	inDropout := inNet.Network()[1].(*neuralnet.DropoutLayer)
	inDropout.Training = dropout
}
