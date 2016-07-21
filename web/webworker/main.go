package main

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/cubewhisper"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/weakai/rnn"
)

const SearchBlankCutoff = 1e-4

var Network rnn.SeqFunc
var SampleRate int
var Samples []float64

func main() {
	js.Global.Set("onmessage", js.MakeFunc(messageHandler))
}

func messageHandler(this *js.Object, dataArg []*js.Object) interface{} {
	if len(dataArg) != 1 {
		panic("expected one argument")
	}
	data := dataArg[0].Get("data")
	command := data.Index(0).String()

	switch command {
	case "init":
		initCommand(data.Index(1).Interface().([]byte))
	case "start":
		SampleRate = data.Index(1).Int()
		Samples = nil
	case "samples":
		list := data.Index(1)
		for i := 0; i < list.Length(); i++ {
			sample := list.Index(i).Float()
			Samples = append(Samples, sample)
		}
	case "end":
		classifySamples()
	}

	return nil
}

func initCommand(rnnData []byte) {
	rnn, err := rnn.DeserializeBidirectional(rnnData)
	if err != nil {
		panic(err)
	}
	Network = rnn
}

func classifySamples() {
	emitLoading("Processing audio")

	sample := cubewhisper.SeqForAudioSamples(Samples, SampleRate)

	inSeq := make([]autofunc.Result, len(sample))
	for i, x := range sample {
		inSeq[i] = &autofunc.Variable{Vector: x}
	}

	emitLoading("Running neural network")

	res := Network.BatchSeqs([][]autofunc.Result{inSeq})

	classification := ctc.PrefixSearch(res.OutputSeqs()[0], SearchBlankCutoff)
	labels := make([]cubewhisper.Label, len(classification))
	for i, c := range classification {
		labels[i] = cubewhisper.Label(c)
	}

	emitMoves(cubewhisper.LabelsToMoveString(labels), fmt.Sprintf("%v", labels))
}

func emitLoading(status string) {
	js.Global.Call("postMessage", map[string]string{"status": status})
}

func emitMoves(moves, raw string) {
	js.Global.Call("postMessage", map[string]string{"moves": moves, "raw": raw})
}
