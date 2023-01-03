package scoreManager

import "C"
import (
	"fmt"
	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
	"unsafe"
)

const (
	MainScoreAddress    = uintptr(0x00000001412EF568)
	WorstCounterAddress = uintptr(0x00000001416E2D40)
	CompletionAddress   = uintptr(0x00000001412EF634)
	PVIdAddress         = uintptr(0x00000001412C2340)
)

func GetScore(PVDiff int32, PVGrade int32) {
	myScore := *((*dataTypes.DivaScore)(unsafe.Pointer(MainScoreAddress)))
	worstCount := *((*int32)(unsafe.Pointer(WorstCounterAddress)))
	completePercent := *((*float32)(unsafe.Pointer(CompletionAddress)))
	PVId := *((*int32)(unsafe.Pointer(PVIdAddress)))

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
	fmt.Print("Grade: ")
	fmt.Println(PVGrade)
	fmt.Print("Difficultly: ")
	fmt.Println(PVDiff)
}
