package main

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/speechrecog/mfcc"
	"github.com/unixpickle/speechrecog/speechdata"
	"github.com/unixpickle/wav"
)

const (
	AudioWindowTime    = time.Millisecond * 20
	AudioWindowOverlap = time.Millisecond * 10
)

type Label int

const (
	Wide Label = iota
	Prime
	Squared
	EMove
	MMove
	SMove
	RMove
	UMove
	FMove
	BMove
	DMove
	LMove
	XMove
	YMove
	ZMove
)

var Labels = []Label{Wide, Prime, Squared, EMove, MMove, SMove, RMove, UMove, FMove, BMove,
	DMove, LMove, XMove, YMove, ZMove}

func ReadSamples(dir string) (sgd.SampleSet, error) {
	index, err := speechdata.LoadIndex(dir)
	if err != nil {
		return nil, err
	}

	var samples sgd.SliceSampleSet
	for _, sample := range index.Samples {
		if sample.File == "" {
			continue
		}
		label := labelForMoveString(sample.Label)
		wavPath := filepath.Join(index.DirPath, sample.File)
		sampleSeq, err := readAudioFile(wavPath)
		if err != nil {
			return nil, err
		}
		intLabel := make([]int, len(label))
		for i, x := range label {
			intLabel[i] = int(x)
		}
		samples = append(samples, ctc.Sample{Input: sampleSeq, Label: intLabel})
	}

	return samples, nil
}

func labelForMoveString(algorithm string) []Label {
	var label []Label
	for _, move := range strings.Fields(algorithm) {
		switch move[0] {
		case 'x':
			label = append(label, XMove)
		case 'y':
			label = append(label, YMove)
		case 'z':
			label = append(label, ZMove)
		case 'r', 'u', 'l', 'd', 'f', 'b':
			label = append(label, Wide)
			fallthrough
		case 'R', 'U', 'L', 'D', 'F', 'B':
			mapping := map[byte]Label{
				'R': RMove, 'U': UMove, 'L': LMove, 'D': DMove,
				'F': FMove, 'B': BMove,
			}
			label = append(label, mapping[move[0]])
		case 'E':
			label = append(label, EMove)
		case 'M':
			label = append(label, MMove)
		case 'S':
			label = append(label, SMove)
		}
		for _, c := range move[1:] {
			switch c {
			case '\'':
				label = append(label, Prime)
			case '2':
				label = append(label, Squared)
			}
		}
	}
	return label
}

func readAudioFile(file string) ([]linalg.Vector, error) {
	wavFile, err := wav.ReadSoundFile(file)
	if err != nil {
		return nil, err
	}

	var audioData []float64
	for i, x := range wavFile.Samples() {
		if i%wavFile.Channels() == 0 {
			audioData = append(audioData, float64(x))
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
