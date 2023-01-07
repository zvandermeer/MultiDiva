package connectionManager

import (
	"strings"

	"github.com/ovandermeer/MultiDiva/internal/scoreManager"
)

type handshake struct {
	instruction   string
	clientVersion string
	username      string
}

type Note struct {
	Instruction string
	Score       int
}

type FinalScore struct {
	Instruction string
	TotalScore  int
	Combo       int
	Cool        int
	Fine        int
	Safe        int
	Sad         int
	Worst       int
	Completion  float32
	PV          int
	Difficulty  scoreManager.Difficulty
	Grade       scoreManager.FinalGrade
}

func encode(input []string) (output string) {
	for i := 0; i < len(input); i++ {
		output = output + input[i] + ","
	}
	return
}

func decode(input string) []string {
	return strings.Split(input, ",")
}
