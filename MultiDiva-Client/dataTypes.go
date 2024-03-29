package main

import (
	"sync"
)

var CurrentConfigVersion int = 1

type ConfigData struct {
	ConfigVersion  int    `yaml:"config_version"`
	Debug          bool   `yaml:"debug"`
	ServerAddress  string `yaml:"server_address"`
	Port           string `yaml:"server_port"`
	SongLimitPatch bool   `yaml:"song_limit_patch_enabled"`
	Username       string `yaml:"username"`
}

func NewConfigData() (config ConfigData) {
	config.ConfigVersion = CurrentConfigVersion
	config.Debug = false
	config.ServerAddress = "localhost"
	config.Port = "9988"
	config.Username = "User"
	return
}

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

type MessageData struct {
	mu   sync.RWMutex
	last []byte
}

func (d *MessageData) Store(data []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.last = data
}

func (d *MessageData) Get() []byte {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.last
}

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
