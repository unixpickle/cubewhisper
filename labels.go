package cubewhisper

import "strings"

// A Label is a unit of speech for narrated Rubik's cube
// algorithms.
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

// Labels contains all of the available speech labels
// in the order they should appear in from a classifier.
var Labels = []Label{Wide, Prime, Squared, EMove, MMove, SMove, RMove, UMove, FMove, BMove,
	DMove, LMove, XMove, YMove, ZMove}

// LabelsForMoveString creates an array of labels for an
// algorithm in Rubik's cube notation.
func LabelsForMoveString(algorithm string) []Label {
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

// LabelsToMoveString turns a sequence of Labels into
// a string in Rubik's cube notation.
func LabelsToMoveString(l []Label) string {
	var parts []string
	var wide bool
	for _, label := range l {
		if label == Wide {
			wide = true
			continue
		}
		switch label {
		case EMove, MMove, SMove, XMove, YMove, ZMove:
			moveMap := map[Label]string{EMove: "E", MMove: "M", SMove: "S",
				XMove: "x", YMove: "y", ZMove: "z"}
			parts = append(parts, moveMap[label])
		case Prime, Squared:
			suffix := "'"
			if label == Squared {
				suffix = "2"
			}
			if len(parts) > 0 {
				p := parts[len(parts)-1]
				parts[len(parts)-1] = p + suffix
			}
		default:
			letters := map[Label]string{RMove: "R", LMove: "L", DMove: "D",
				FMove: "F", UMove: "U", BMove: "B"}
			letter := letters[label]
			if wide {
				letter = strings.ToLower(letter)
			}
			parts = append(parts, letter)
		}
		wide = false
	}
	return strings.Join(parts, " ")
}
