// Command rate runs a speech RNN on a directory of
// speech samples and prints out the error for each
// of the samples.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/cubewhisper"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/speechrecog/speechdata"
	"github.com/unixpickle/weakai/rnn"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: rate <rnn> <sample dir>")
		os.Exit(1)
	}

	rnnData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		die("Read RNN", err)
	}
	seqFunc, err := rnn.DeserializeBidirectional(rnnData)
	if err != nil {
		die("Deserialize RNN", err)
	}

	index, err := speechdata.LoadIndex(os.Args[2])
	if err != nil {
		die("Load speech index", err)
	}

	log.Println("Crunching numbers...")

	var res results
	for _, sample := range index.Samples {
		if sample.File == "" {
			continue
		}
		label := cubewhisper.LabelsForMoveString(sample.Label)
		wavPath := filepath.Join(index.DirPath, sample.File)
		sampleSeq, err := cubewhisper.ReadAudioFile(wavPath)
		if err != nil {
			die("Load sample audio", err)
		}
		intLabel := make([]int, len(label))
		for i, x := range label {
			intLabel[i] = int(x)
		}
		output := evalSample(seqFunc, sampleSeq)
		likelihood := ctc.LogLikelihood(output, intLabel).Output()[0]
		res.Likelihoods = append(res.Likelihoods, likelihood)
		res.SampleIDs = append(res.SampleIDs, sample.ID)
	}
	sort.Sort(&res)

	for i, id := range res.SampleIDs {
		likelihood := res.Likelihoods[i]
		fmt.Printf("%d. %s - %e\n", i, id, likelihood)
	}
}

func die(action string, err error) {
	fmt.Fprintln(os.Stderr, action+": "+err.Error())
	os.Exit(1)
}

func evalSample(net rnn.SeqFunc, seq []linalg.Vector) []autofunc.Result {
	inVars := make([]autofunc.Result, len(seq))
	for i, x := range seq {
		inVars[i] = &autofunc.Variable{Vector: x}
	}
	input := [][]autofunc.Result{inVars}
	output := net.BatchSeqs(input)
	rawOut := output.OutputSeqs()[0]
	res := make([]autofunc.Result, len(rawOut))
	for i, x := range rawOut {
		res[i] = &autofunc.Variable{Vector: x}
	}
	return res
}

type results struct {
	Likelihoods []float64
	SampleIDs   []string
}

func (r *results) Len() int {
	return len(r.Likelihoods)
}

func (r *results) Swap(i, j int) {
	r.Likelihoods[i], r.Likelihoods[j] = r.Likelihoods[j], r.Likelihoods[i]
	r.SampleIDs[i], r.SampleIDs[j] = r.SampleIDs[j], r.SampleIDs[i]
}

func (r *results) Less(i, j int) bool {
	return r.Likelihoods[i] < r.Likelihoods[j]
}
