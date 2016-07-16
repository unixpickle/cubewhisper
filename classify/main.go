package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/speechrecog/mfcc"
	"github.com/unixpickle/wav"
	"github.com/unixpickle/weakai/rnn"
)

const (
	AudioWindowTime    = time.Millisecond * 20
	AudioWindowOverlap = time.Millisecond * 10

	NoiseAmount = 1e-5
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: classify <rnn> <sample.wav>")
		os.Exit(1)
	}

	// TODO: handle errors.
	rnnData, _ := ioutil.ReadFile(os.Args[1])
	seqFunc, _ := rnn.DeserializeBidirectional(rnnData)
	sample, _ := readAudioFile(os.Args[2])

	inSeq := make([]autofunc.Result, len(sample))
	for i, x := range sample {
		inSeq[i] = &autofunc.Variable{Vector: x}
	}
	res := seqFunc.BatchSeqs([][]autofunc.Result{inSeq})

	classification := ctc.BestPath(res.OutputSeqs()[0])
	fmt.Println(classification)
}

func readAudioFile(file string) ([]linalg.Vector, error) {
	wavFile, err := wav.ReadSoundFile(file)
	if err != nil {
		return nil, err
	}

	var audioData []float64
	for i, x := range wavFile.Samples() {
		if i%wavFile.Channels() == 0 {
			sample := float64(x) + rand.NormFloat64()*NoiseAmount
			audioData = append(audioData, sample)
		}
	}

	mfccSource := mfcc.MFCC(&mfcc.SliceSource{Slice: audioData}, wavFile.SampleRate(),
		&mfcc.Options{Window: AudioWindowTime, Overlap: AudioWindowOverlap})
	mfccSource = mfcc.AddVelocities(mfccSource)

	var coeffs []linalg.Vector
	for {
		c, err := mfccSource.NextCoeffs()
		if err == nil {
			coeffs = append(coeffs, c)
		} else {
			break
		}
	}

	return coeffs, nil
}
