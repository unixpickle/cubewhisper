package cubewhisper

import (
	"math/rand"
	"time"

	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/speechrecog/mfcc"
	"github.com/unixpickle/wav"
)

const (
	audioWindowTime    = time.Millisecond * 20
	audioWindowOverlap = time.Millisecond * 10

	noiseAmount = 1e-5
)

// ReadAudioFile reads an audio file and converts it
// into a sequence of MFCC vectors.
func ReadAudioFile(file string) ([]linalg.Vector, error) {
	wavFile, err := wav.ReadSoundFile(file)
	if err != nil {
		return nil, err
	}

	var audioData []float64
	for i, x := range wavFile.Samples() {
		if i%wavFile.Channels() == 0 {
			// Add random noise to avoid zero-power chunks
			// of signal which cause -Infs in the MFCCs.
			sample := float64(x) + rand.NormFloat64()*noiseAmount
			audioData = append(audioData, sample)
		}
	}

	mfccSource := mfcc.MFCC(&mfcc.SliceSource{Slice: audioData}, wavFile.SampleRate(),
		&mfcc.Options{Window: audioWindowTime, Overlap: audioWindowOverlap})
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
