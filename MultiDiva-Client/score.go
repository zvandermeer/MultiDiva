package main

import "C"
import (
	"fmt"
	"strconv"
	"unsafe"
)

type DivaScore struct {
	TotalScore     uint32
	Unknown1       uint32
	Unknown2       uint32
	Unknown3       uint32
	Unknown4       uint32
	Unknown5       uint32
	Unknown6       uint32
	Unknown7       uint32
	Unknown8       uint32
	Combo          uint32
	PreAdjustCool  uint32
	PreAdjustFine  uint32
	PreAdjustSafe  uint32
	PreAdjustSad   uint32
	PreAdjustWorst uint32
	Cool           uint32
	Fine           uint32
	Safe           uint32
	Sad            uint32
	Worst          uint32
}

// Due to how holds are added to the score by Diva, tracking "Bad", "Wrong Bad" and "Wrong Safe" grades isn't possible through this method
type NoteGrade int64

const (
	Cool NoteGrade = iota
	Good
	Safe
	Cool_Wrong
	Good_Wrong
)

type FinalGrade int64

const (
	Failed FinalGrade = iota
	Cheap
	Standard
	Great
	Excellent
	Perfect
)

type Difficulty int64

const (
	Easy Difficulty = iota
	Normal
	Hard
	Extreme
	ExExtreme
)

type NoteData struct {
	FullScore   int32
	SlicedScore []int32
	Combo       int32
	Grade       NoteGrade
}


const (
	MainScoreAddress            = uintptr(0x00000001412EF568)
	WorstCounterAddress         = uintptr(0x00000001416E2D40)
	CompletionAddress           = uintptr(0x00000001412EF634)
	PVIdAddress                 = uintptr(0x00000001412C2340)
	SongLimitPatchPVDiffAddress = uintptr(0x00000001423157AC) // SongLimitPatch 1.02
	VanillaPVDiffAddress        = uintptr(0x00000001412B634C) // Non-SongLimitPatch 1.02
	PVGradeAddress              = uintptr(0x00000001416E2D00)
	PVTitleAddress              = uintptr(0x00000001412EF228)
	ScoreAddress                = uintptr(0x1412EF56C)
	ComboAddress                = uintptr(0x1412EF578)
)

var oldScore int
var lastCombo int = -1

func GetFrameScore() {
	score := *((*int)(unsafe.Pointer(ScoreAddress)))
	combo := *((*int)(unsafe.Pointer(ComboAddress)))

	score = score - 4294967296

	if score != oldScore {

		scoreChange := score - oldScore

		// For every 10 combo, you get a 50 point bonus, so subtract that. Bonus also maxes out at 50 combo
		if combo >= 50 {
			scoreChange = scoreChange - 250
		} else {
			scoreChange = scoreChange - ((combo / 10) * 50)
		}

		oldScore = score

		updateScore := false

		var noteGrade NoteGrade

		switch scoreChange {
		// Add to account for possible health bonus, or combo being added at the same time. single combo adds 10, double 20, so on. Multi notes are also
		// accounted for. A double press at cool is worth 1000 points, but there might be a hold bonus or health bonus on top of that. Triple press
		// at cool is worth 1500, etc.
		case 500, 510, 520, 530, 540, 550, 1000, 1010, 1020, 1030, 1040, 1050, 1500, 1510, 1520, 1530, 1540, 1550, 2000, 2010, 2020, 2030, 2040, 2050:
			if lastCombo < combo {
				noteGrade = Cool
				updateScore = true
			}
		case 300, 310, 320, 330, 340, 350, 600, 610, 620, 630, 640, 650, 900, 910, 920, 930, 940, 950, 1200, 1210, 1220, 1230, 1240, 1250:
			if lastCombo < combo {
				noteGrade = Good
				updateScore = true
			}
		case 100, 110, 120, 130, 140, 200, 210, 220, 230, 240, 400, 410, 420, 430, 440:
			if combo == 0 {
				noteGrade = Safe
				updateScore = true
			}
		// Everything below this point causes you to lose health, so a health bonus isn't possible. Also skipped over some values that conflict
		// with other scores. Opted to go for the score that's more likely to happen
		case 250, 260, 270, 280, 290, 750, 760, 770, 780, 790:
			if combo == 0 {
				noteGrade = Cool_Wrong
				updateScore = true
			}
		case 150, 160, 170, 180, 190, 450, 460, 470, 480, 490:
			if combo == 0 {
				noteGrade = Good_Wrong
				updateScore = true
			}
		}

		lastCombo = combo

		if updateScore {
			scoreString := strconv.Itoa(score)
			divalog.Log("Score: " + scoreString, 2)
			divalog.Log("Change in score: " + strconv.Itoa(scoreChange), 2)
			divalog.Log("Note Grade: %s", 2)
			divalog.Log(noteGrade, 2)

			myData := map[string]string{
				"Instruction": "note",
				"Score":       strconv.Itoa(score),
				"Combo":       strconv.Itoa(combo),
				"Grade":       fmt.Sprintf("%d", int(noteGrade)),
			}

			myClient.sendJsonMessage(myData)
		}
	}
}

func GetFinalScore() {
	myScore := *((*DivaScore)(unsafe.Pointer(MainScoreAddress)))
	worstCount := *((*uint32)(unsafe.Pointer(WorstCounterAddress)))
	completePercent := *((*float32)(unsafe.Pointer(CompletionAddress)))
	PVId := *((*uint32)(unsafe.Pointer(PVIdAddress)))
	var PVDiff int64
	if cfg.SongLimitPatch {
		PVDiff = int64(*((*uint32)(unsafe.Pointer(SongLimitPatchPVDiffAddress))))
	} else {
		PVDiff = int64(*((*uint32)(unsafe.Pointer(VanillaPVDiffAddress))))
	}
	PVGrade := int64(*((*uint32)(unsafe.Pointer(PVGradeAddress))))

	divalog.Log("Total score: ", 2)
	divalog.Log(myScore.TotalScore, 2)
	divalog.Log("Combo: ", 2)
	divalog.Log(myScore.Combo, 2)
	divalog.Log("Cool: ", 2)
	divalog.Log(myScore.Cool, 2)
	divalog.Log("Fine: ", 2)
	divalog.Log(myScore.Fine, 2)
	divalog.Log("Safe: ", 2)
	divalog.Log(myScore.Safe, 2)
	divalog.Log("Sad: ", 2)
	divalog.Log(myScore.Sad, 2)
	divalog.Log("Worst: ", 2)
	divalog.Log(worstCount, 2)
	divalog.Log("Completion: ", 2)
	divalog.Log(completePercent, 2)
	divalog.Log("PV: ", 2)
	divalog.Log(PVId, 2)
	divalog.Log("Difficulty: ", 2)
	divalog.Log(PVDiff, 2)
	divalog.Log("Grade: ", 2)
	divalog.Log(PVGrade, 2)

	myData := map[string]string{
		"Instruction": "finalScore",
		"TotalScore":  strconv.FormatUint(uint64(myScore.TotalScore), 10),
		"Combo":       strconv.FormatUint(uint64(myScore.Combo), 10),
		"Cool":        strconv.FormatUint(uint64(myScore.Cool), 10),
		"Fine":        strconv.FormatUint(uint64(myScore.Fine), 10),
		"Safe":        strconv.FormatUint(uint64(myScore.Safe), 10),
		"Sad":         strconv.FormatUint(uint64(myScore.Sad), 10),
		"Worst":       strconv.FormatUint(uint64(worstCount), 10),
		"Completion":  fmt.Sprintf("%f", completePercent),
		"PV":          strconv.FormatUint(uint64(PVId), 10),
		"Difficulty":  strconv.FormatUint(uint64(PVDiff), 10),
		"Grade":       strconv.FormatUint(uint64(PVGrade), 10),
	}

	myClient.sendJsonMessage(myData)
}

// split integer into slice of single digits
func splitInt(n int) []int {
	var slc []int
	for n > 0 {
		slc = append(slc, n%10)
		n = n / 10
	}
	return slc
}
