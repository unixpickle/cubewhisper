package main

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/cubewhisper"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/weakai/rnn"
)

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
	sample := cubewhisper.SeqForAudioSamples(Samples, SampleRate)

	inSeq := make([]autofunc.Result, len(sample))
	for i, x := range sample {
		inSeq[i] = &autofunc.Variable{Vector: x}
	}

	res := Network.BatchSeqs([][]autofunc.Result{inSeq})

	classification := ctc.BestPath(res.OutputSeqs()[0])
	labels := make([]cubewhisper.Label, len(classification))
	for i, c := range classification {
		labels[i] = cubewhisper.Label(c)
	}

	emitMoves("Algorithms: " + cubewhisper.LabelsToMoveString(labels) +
		" (Raw labels: " + fmt.Sprintf("%v", labels) + ")")
}

func emitMoves(moves string) {
	js.Global.Call("postMessage", moves)
}
