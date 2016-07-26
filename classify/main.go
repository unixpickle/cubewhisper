package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/cubewhisper"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/weakai/rnn"
)

const PrefixThreshold = -1e-4

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: classify <rnn> <sample.wav>")
		os.Exit(1)
	}

	rnnData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		die(err)
	}
	seqFunc, err := rnn.DeserializeBidirectional(rnnData)
	if err != nil {
		die(err)
	}
	sample, err := cubewhisper.ReadAudioFile(os.Args[2])
	if err != nil {
		die(err)
	}

	inSeq := make([]autofunc.Result, len(sample))
	for i, x := range sample {
		inSeq[i] = &autofunc.Variable{Vector: x}
	}
	res := seqFunc.BatchSeqs([][]autofunc.Result{inSeq})

	classification := ctc.PrefixSearch(res.OutputSeqs()[0], PrefixThreshold)
	labels := make([]cubewhisper.Label, len(classification))
	for i, c := range classification {
		labels[i] = cubewhisper.Label(c)
	}
	fmt.Println("Raw labels:", labels)
	fmt.Println("Algorithm:", cubewhisper.LabelsToMoveString(labels))
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
