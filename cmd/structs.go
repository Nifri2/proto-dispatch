package cmd

type Role int

const (
	Dispatcher Role = 0x00 + iota
	Worker
)

type Address int

const (
	Dispatch Address = 0x00 + iota
	Worker_0         // 0x01
	Worker_1         // 0x02
	Worker_2         // 0x03
	Worker_3         // 0x04
)

type Command int

const (
	Cmd_NoOp        Command = 0x00 + iota
	Cmd_LedOn               // 0x01
	Cmd_LedOff              // 0x02
	Cmd_DisplayAnim         // 0x03
)

type AnimationID int

const (
	Anim_EyeIdle      AnimationID = 0x00 + iota
	Anim_EyeBlink                 // 0x01
	Anim_MouthIdle                // 0x02
	Anim_Nifri                    // 0x03
	Anim_SpinnyLambda             // 0x04
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

type animUpdate struct {
	Eye   *Animation
	Mouth *Animation
}
