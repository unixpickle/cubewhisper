// Command samplegen generates random move sequences
// suited for use with the recorder in
// https://github.com/unixpickle/speechrecog.
package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	Moves = []string{"x", "y", "z", "E", "M", "S", "R", "U", "F", "B", "D", "L",
		"r", "u", "f", "b", "d", "l"}
	Suffixes = []string{"", "'", "2", "2'"}
)

const (
	MinLen = 1
	MaxLen = 30
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: samplegen <count>")
		os.Exit(1)
	}
	count, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid count:", os.Args[1])
		os.Exit(1)
	}
	fmt.Println("[")
	for i := 0; i < count; i++ {
		fmt.Printf("  {\"Label\": \"%s\", \"File\":\"\", \"ID\": \"%s\"},\n",
			generateScramble(MinLen+rand.Intn(1+MaxLen-MinLen)),
			randomID())
	}
	fmt.Println("]")
}

func generateScramble(size int) string {
	var parts []string
	for i := 0; i < size; i++ {
		prefix := Moves[rand.Intn(len(Moves))]
		suffix := Suffixes[rand.Intn(len(Suffixes))]
		parts = append(parts, prefix+suffix)
	}
	return strings.Join(parts, " ")
}

func randomID() string {
	var buf [16]byte
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(rand.Intn(0x100))
	}
	return strings.ToLower(hex.EncodeToString(buf[:]))
}
