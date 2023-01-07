package scoreManager

import "C"
import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/ovandermeer/MultiDiva/internal/connectionManager"
)

const (
	MainScoreAddress    = uintptr(0x00000001412EF568)
	WorstCounterAddress = uintptr(0x00000001416E2D40)
	CompletionAddress   = uintptr(0x00000001412EF634)
	PVIdAddress         = uintptr(0x00000001412C2340)
	PVDiffAddress       = uintptr(0x00000001423157AC) // Song limit patch
	PVGradeAddress      = uintptr(0x00000001416E2D00)
	ScoreAddress        = uintptr(0x1412EF56C)
	ComboAddress        = uintptr(0x1412EF578)
)

type FinalGrade int64

const (
	Failed = iota
	Cheap
	Standard
	Great
	Excellent
	Perfect
)

type Difficulty int64

const (
	Easy = iota
	Normal
	Hard
	Extreme
	ExExtreme
)

// Due to how holds are added to the score by Diva, tracking "Bad", "Wrong Bad" and "Wrong Safe" grades isn't possible through this method
type note string

const (
	Cool       note = "Cool"
	Good            = "Good"
	Safe            = "Safe"
	Cool_Wrong      = "Wrong Cool"
	Good_Wrong      = "Wrong Good"
	Bad             = "Bad"
)

var oldScore int
var totalNotes int
var lastTotalNotes int

type ScoreData struct {
	mu          sync.RWMutex
	score       int
	scoreChange int
	noteGrade   note
	totalNotes  int
}

func (d *ScoreData) Store(score int, scoreChange int, noteGrade note, totalNotes int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.score = score
	d.scoreChange = scoreChange
	d.noteGrade = noteGrade
	d.totalNotes = totalNotes
}
func (d *ScoreData) Get() (int, int, note, int) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.score, d.scoreChange, d.noteGrade, d.totalNotes
}

func GetFrameScore(s *ScoreData) {
	score := *((*int)(unsafe.Pointer(ScoreAddress)))
	combo := *((*int)(unsafe.Pointer(ComboAddress)))

	score = score - 4294967296

	if score != oldScore {
		scoreChange := score - oldScore

		//fmt.Println("RAW SCORE CHANGE: " + strconv.Itoa(scoreChange))

		// For every 10 combo, you get a 50 point bonus, so subtract that. Bonus also maxes out at 50 combo
		if combo >= 50 {
			scoreChange = scoreChange - 250
		} else {
			scoreChange = scoreChange - ((combo / 10) * 50)
		}

		oldScore = score

		updateScore := false

		var noteGrade note

		switch scoreChange {
		// Add 10 to account for possible health bonus
		case 500, 510:
			noteGrade = Cool
			updateScore = true
		case 300, 310:
			noteGrade = Good
			updateScore = true
		case 100, 110:
			noteGrade = Safe
			updateScore = true
		// Everything below this point causes you to lose health, so a health bonus isn't possible
		case 250:
			noteGrade = Cool_Wrong
			updateScore = true
		case 150:
			noteGrade = Good_Wrong
			updateScore = true
		}

		if updateScore {
			totalNotes += 1
			s.Store(score, scoreChange, noteGrade, totalNotes)
		}
	}
}

func GetFinalScore() {
	myScore := *((*DivaScore)(unsafe.Pointer(MainScoreAddress)))
	worstCount := *((*int)(unsafe.Pointer(WorstCounterAddress)))
	completePercent := *((*float32)(unsafe.Pointer(CompletionAddress)))
	PVId := *((*int)(unsafe.Pointer(PVIdAddress)))
	PVDiff := Difficulty(*((*int)(unsafe.Pointer(PVDiffAddress))))
	PVGrade := FinalGrade(*((*int)(unsafe.Pointer(PVGradeAddress))))

	fmt.Print("Total score: ")
	fmt.Println(myScore.TotalScore)
	fmt.Print("Combo: ")
	fmt.Println(myScore.Combo)
	fmt.Print("Cool: ")
	fmt.Println(myScore.Cool)
	fmt.Print("Fine: ")
	fmt.Println(myScore.Fine)
	fmt.Print("Safe: ")
	fmt.Println(myScore.Safe)
	fmt.Print("Sad: ")
	fmt.Println(myScore.Sad)
	fmt.Print("Worst: ")
	fmt.Println(worstCount)
	fmt.Print("Completion: ")
	fmt.Println(completePercent)
	fmt.Print("PV: ")
	fmt.Println(PVId)
	fmt.Print("Difficulty: ")
	fmt.Println(PVDiff)
	fmt.Print("Grade: ")
	fmt.Println(PVGrade)

	myData, _ := json.Marshal(connectionManager.FinalScore{
		Instruction: "finalScore",
		TotalScore:  myScore.TotalScore,
		Combo:       myScore.Combo,
		Cool:        myScore.Cool,
		Fine:        myScore.Fine,
		Safe:        myScore.Safe,
		Sad:         myScore.Sad,
		Worst:       worstCount,
		Completion:  completePercent,
		PV:          PVId,
		Difficulty:  PVDiff,
		Grade:       PVGrade})

	connectionManager.SendToServer(myData)
}

func NoteHit(s *ScoreData) {
	time.Sleep(10 * time.Millisecond)
	score, scoreChange, noteGrade, myTotalNotes := s.Get()

	// If the score was successfully found and grabbed by the GetFrameScore() method, then this is skipped.
	// If not, (ie the score change is too low to differentiate between a note being hit and a hold), then
	// grab the current value at the pointer, assume the noteGrade is bad, and continue on
	if myTotalNotes == lastTotalNotes {
		score = *((*int)(unsafe.Pointer(ScoreAddress)))
		scoreChange = 0
		noteGrade = Bad
	}

	fmt.Print("Score: ")
	fmt.Println(score)
	fmt.Print("Change in score: ")
	fmt.Println(scoreChange)
	fmt.Println("Note Grade: " + noteGrade)

	data, _ := json.Marshal(connectionManager.Note{
		Instruction: "note",
		Score:       score,
	})

	connectionManager.SendToServer(data)
}
