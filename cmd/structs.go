package cmd

type Role int

const (
	Dispatcher Role = 0x00 + iota
	Worker
)

type Address int

const (
	Dispatch Address = 0x00 + iota
	Worker_0
	Worker_1
	Worker_2
	Worker_3
)

type Command int

const (
	Cmd_NoOp Command = 0x00 + iota
	Cmd_LedOn
	Cmd_LedOff
)

type Settings struct {
	Role    Role
	Address Address
}

type Animation struct {
	Frames     [][]byte // slice of frames, each frame is []byte
	FrameCount int
	Name       string
}

var LoadedAnimations []*Animation

const (
	EyeFrameWidth  = 16
	EyeFrameHeight = 16

	MouthFrameWidth  = 32
	MouthFrameHeight = 16
)
