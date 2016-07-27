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
	HiddenSize    = 128
	OutHiddenSize = 128
	BatchSize     = 16
	CostBatchSize = 4

	MaxSubBatch    = 8
	MaxConcurrency = 2

	CrossRatio = 0.3

	WeightStddev  = 0.05
	InputNoise    = 0.3
	HiddenDropout = 1
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

	// Always shuffle the samples in the same way.
	rand.Seed(123)
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
	toggleRegularization(seqFunc, true)
	sgd.SGDInteractive(gradienter, training, stepSize, BatchSize, func() bool {
		toggleRegularization(seqFunc, false)
		cost := ctc.TotalCost(seqFunc, training, CostBatchSize, MaxConcurrency)
		crossCost := ctc.TotalCost(seqFunc, validation, CostBatchSize, MaxConcurrency)
		toggleRegularization(seqFunc, true)
		log.Printf("Epoch %d: cost=%e cross=%e", epoch, cost, crossCost)
		epoch++
		return true
	})
	toggleRegularization(seqFunc, false)

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
			OutputCount: OutHiddenSize,
		},
		&neuralnet.HyperbolicTangent{},
		&neuralnet.DenseLayer{
			InputCount:  OutHiddenSize,
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
		&neuralnet.GaussNoiseLayer{
			Stddev:   InputNoise,
			Training: false,
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

func toggleRegularization(bd *rnn.Bidirectional, enabled bool) {
	output := bd.Output.(*rnn.NetworkSeqFunc).Network[0].(*neuralnet.DropoutLayer)
	output.Training = enabled
	inNet := bd.Forward.(*rnn.RNNSeqFunc).Block.(rnn.StackedBlock)[0].(*rnn.NetworkBlock)
	inNoise := inNet.Network()[1].(*neuralnet.GaussNoiseLayer)
	inNoise.Training = enabled
}
