// Train trains a bidirectional neural network using
// CTC to predict move strings given voice data.
package main

import (
	"fmt"
	"os"
	"strconv"
)

const DefaultStepSize = 1e-3

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "Usage: train <rnn file> <sample dir> [step size]")
		os.Exit(1)
	}
	rnnFile := os.Args[1]
	sampleDir := os.Args[2]
	stepSize := DefaultStepSize

	if len(os.Args) == 4 {
		var err error
		stepSize, err = strconv.ParseFloat(os.Args[3], 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid step size:", os.Args[3])
			os.Exit(1)
		}
	}

	Train(rnnFile, sampleDir, stepSize)
}
