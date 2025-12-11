package cmd

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

// Each MCU takes care of 1 mouth and 1 eye WS2812 strip
// There will be a sperate worker that will only display Insignia animations

type animUpdate struct {
	Eye   *Animation
	Mouth *Animation
}

func RunWorker(config Settings, uart *machine.UART, led machine.Pin) {
	// Listen for commands from Dispatcher
	fmt.Println("Starting Worker Loop")

	animChan := make(chan animUpdate, 1)

	// Start animation routine in background
	go displayAnimation(animChan)

	for {
		if uart.Buffered() >= 4 {
			addrByte, _ := uart.ReadByte()
			cmdByte, _ := uart.ReadByte()
			eyeIdxByte, _ := uart.ReadByte()
			mouthIdxByte, _ := uart.ReadByte()

			// Only react if packet is for this worker
			if Address(addrByte) == config.Address {
				cmd := Command(cmdByte)
				// fmt.Printf("Rx: Cmd=%d Eye=%d Mouth=%d\n", cmd, eyeIdxByte, mouthIdxByte)

				switch cmd {
				case Cmd_LedOn:
					led.High()
				case Cmd_LedOff:
					led.Low()
				case Cmd_NoOp:
					// No operation
				case Cmd_DisplayAnim:
					eyeIdx := int(eyeIdxByte)
					mouthIdx := int(mouthIdxByte)

					var update animUpdate

					if eyeIdx >= 0 && eyeIdx < len(LoadedAnimations) {
						update.Eye = LoadedAnimations[eyeIdx]
					}

					if mouthIdx >= 0 && mouthIdx < len(LoadedAnimations) {
						update.Mouth = LoadedAnimations[mouthIdx]
					}

					// Send update if we have at least one valid animation
					if update.Eye != nil || update.Mouth != nil {
						animChan <- update
					}
				}
			}
		}
		// Small sleep to yield if loop is tight, though Buffered() check is fast
		time.Sleep(time.Millisecond)
	}
}

// TODOS:
// - Add Functionality to display 2 animations (eye + mouth), need to modify command protocol
// 	- - Currently only 1 animation channel is supported
//  - - Since 1 MCU handles both eye and mouth at a stabkle 47hz we use 2 strips at 2 different pins
// - Implement all Commands (LED on/off, NoOp, etc) Esp. Cmd_DisplayAnim

func displayAnimation(animChan chan animUpdate) {
	// Wait for the board to stabilize
	time.Sleep(2 * time.Second)

	// Defaults if loading failed (or empty)
	// We assume LoadedAnimations is populated by main before calling RunWorker

	// Helper to find animation by name
	findAnim := func(name string) *Animation {
		for _, a := range LoadedAnimations {
			if a.Name == name {
				return a
			}
		}
		return nil
	}

	eyeIdleAnim := findAnim("eye_idle")
	mouthAnim := findAnim("mouth_idle")
	// nifriAnim := findAnim("nifri")
	// spinnyAnim := findAnim("spinnylambda")

	// Logic from old.logic, but using GP2 and GP3 to avoid UART conflict
	ledPin1 := machine.GP2
	ledPin1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip1 := ws2812.New(ledPin1)

	ledPin2 := machine.GP3
	ledPin2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip2 := ws2812.New(ledPin2)

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
		case update := <-animChan:
			if update.Eye != nil && update.Eye != currentEyeAnim {
				fmt.Printf("Switching Eye to: %s\n", update.Eye.Name)
				currentEyeAnim = update.Eye
				eyeFrameCounter = 0
			}
			if update.Mouth != nil && update.Mouth != currentMouthAnim {
				fmt.Printf("Switching Mouth to: %s\n", update.Mouth.Name)
				currentMouthAnim = update.Mouth
				mouthFrameCounter = 0
			}
		default:
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
