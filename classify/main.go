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

	classification := ctc.BestPath(res.OutputSeqs()[0])
	fmt.Println(classification)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}