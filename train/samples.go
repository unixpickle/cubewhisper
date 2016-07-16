package main

import (
	"path/filepath"

	"github.com/unixpickle/cubewhisper"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/speechrecog/speechdata"
)

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
		label := cubewhisper.LabelsForMoveString(sample.Label)
		wavPath := filepath.Join(index.DirPath, sample.File)
		sampleSeq, err := cubewhisper.ReadAudioFile(wavPath)
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
