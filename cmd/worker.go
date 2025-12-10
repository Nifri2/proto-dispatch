package cmd

import (
	"fmt"
	"machine"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

func RunWorker(config Settings, uart *machine.UART, led machine.Pin) {
	// Listen for commands from Dispatcher
	fmt.Println("Starting Worker Loop")

	animChan := make(chan *Animation, 1)

	// Start animation routine in background
	go displayAnimation(animChan)

	for {
		if uart.Buffered() >= 2 {
			addr, _ := uart.ReadByte()
			animIdx, _ := uart.ReadByte()
			fmt.Printf("Received Addr: %d AnimID: %d\n", addr, animIdx)

			// Only react if packet is for this worker
			if Address(addr) == config.Address {
				idx := int(animIdx)
				if idx >= 0 && idx < len(LoadedAnimations) {
					animChan <- LoadedAnimations[idx]
				} else {
					fmt.Printf("Invalid animation index: %d\n", idx)
				}
			}
		}
		// Small sleep to yield if loop is tight, though Buffered() check is fast
		time.Sleep(time.Millisecond)
	}
}

func displayAnimation(animChan chan *Animation) {
	// Wait for the board to stabilize
	// time.Sleep(2 * time.Second)

	// Defaults if loading failed (or empty)
	// We assume LoadedAnimations is populated by main before calling RunWorker
	// But we need to find the specific animations by name or index?
	// The original code used specific variables: eyeBlinkAnim, etc.
	// Since we now have a global LoadedAnimations, we need to pick from it.
	// However, the logic relies on knowing which one is "eye_idle" vs "eye_blink".

	// Helper to find animation by name
	findAnim := func(name string) *Animation {
		for _, a := range LoadedAnimations {
			if a.Name == name {
				return a
			}
		}
		return nil
	}

	eyeBlinkAnim := findAnim("eye_blink")
	eyeIdleAnim := findAnim("eye_idle")
	mouthAnim := findAnim("mouth_idle")
	// nifriAnim := findAnim("nifri")
	// spinnyAnim := findAnim("spinnylambda")

	// Logic from old logic, but using GP2 and GP3 to avoid UART conflict
	ledPin1 := machine.GP2
	ledPin1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip1 := ws2812.New(ledPin1)

	ledPin2 := machine.GP3
	ledPin2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip2 := ws2812.New(ledPin2)

	blinkChannel := make(chan bool, 1)
	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for {
			sleepDuration := time.Duration(5000+r.Intn(2001)) * time.Millisecond
			time.Sleep(sleepDuration)
			select {
			case blinkChannel <- true:
			default:
			}
		}
	}()

	baseAnim := eyeIdleAnim
	if baseAnim == nil {
		// Fallback if not found
		if len(LoadedAnimations) > 0 {
			baseAnim = LoadedAnimations[0]
		} else {
			baseAnim = &Animation{FrameCount: 1, Frames: [][]byte{{}}} // Dummy
		}
	}
	currentEyeAnim := baseAnim

	// Fallback for mouth
	currentMouthAnim := mouthAnim
	if currentMouthAnim == nil {
		currentMouthAnim = &Animation{FrameCount: 1, Frames: [][]byte{{}}}
	}

	var eyeFrameCounter int64
	var mouthFrameCounter int64

	for {
		// Check for new animation command
		select {
		case newAnim := <-animChan:
			fmt.Printf("Switching animation to: %s\n", newAnim.Name)
			baseAnim = newAnim
			currentEyeAnim = baseAnim
			eyeFrameCounter = 0
		case <-blinkChannel:
			// Only blink if we are in the idle state
			if eyeBlinkAnim != nil && baseAnim == eyeIdleAnim && currentEyeAnim != eyeBlinkAnim {
				currentEyeAnim = eyeBlinkAnim
				eyeFrameCounter = 0
			}
		default:
		}

		// If blinking finished, return to base animation
		if currentEyeAnim == eyeBlinkAnim && eyeFrameCounter >= int64(currentEyeAnim.FrameCount) {
			currentEyeAnim = baseAnim
			eyeFrameCounter = 0
		}

		// Safety check for nil animations if load failed
		if currentEyeAnim != nil && len(currentEyeAnim.Frames) > 0 {
			eyeFrame := currentEyeAnim.Frames[eyeFrameCounter%int64(currentEyeAnim.FrameCount)]
			_, err := strip1.Write(eyeFrame)
			if err != nil {
				// println(err.Error()) // Optional: reduce spam
			}
		}

		if currentMouthAnim != nil && len(currentMouthAnim.Frames) > 0 {
			mouthFrame := currentMouthAnim.Frames[mouthFrameCounter%int64(currentMouthAnim.FrameCount)]
			_, err := strip2.Write(mouthFrame)
			if err != nil {
				// println(err.Error())
			}
		}

		// Delay for the WS2812 reset pulse
		time.Sleep(300 * time.Microsecond)

		eyeFrameCounter++
		mouthFrameCounter++

		// Yield to allow other goroutines (like UART) to run if needed
		time.Sleep(10 * time.Millisecond)
	}
}
