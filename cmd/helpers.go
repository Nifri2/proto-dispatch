package cmd

import (
	"encoding/binary"
	"fmt"
)

func LoadAnimation(data []byte, width, height int, name string) (*Animation, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}

	bytesPerFrame := width * height * 3

	frameCount := int(binary.LittleEndian.Uint32(data[:4]))
	expectedSize := 4 + frameCount*bytesPerFrame
	if len(data) != expectedSize {
		return nil, fmt.Errorf("invalid data length for %s: expected %d, got %d", name, expectedSize, len(data))
	}

	frames := make([][]byte, frameCount)
	for i := 0; i < frameCount; i++ {
		start := 4 + i*bytesPerFrame
		end := start + bytesPerFrame
		frames[i] = data[start:end]
	}

	return &Animation{
		Frames:     frames,
		FrameCount: frameCount,
		Name:       name,
	}, nil
}

func ParseRole(r string) Role {
	switch r {
	case "worker":
		return Worker
	default:
		return Dispatcher
	}
}

func ParseAddress(a string) Address {
	switch a {
	case "worker-0":
		return Worker_0
	case "worker-1":
		return Worker_1
	case "worker-2":
		return Worker_2
	case "worker-3":
		return Worker_3
	default:
		return Dispatch
	}
}
