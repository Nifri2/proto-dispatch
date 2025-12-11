package cmd

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

// Each MCU takes care of 1 mouth and 1 eye WS2812 strip
// There will be a sperate worker that will only display Insignia animations

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
			// msg := fmt.Sprintf("Rx: Addr=%x Cmd=%x Eye=%x Mouth=%x\n", addrByte, cmdByte, eyeIdxByte, mouthIdxByte)
			// fmt.Print(msg)

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
// - Implement some sort of Queue for animation updates? The animations abruptly change now and it doesnt look great

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

	// Logic from old.logic, but using GP2 and GP3 to avoid UART conflict
	ledPin1 := machine.GP2
	ledPin1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip1 := ws2812.New(ledPin1)

	ledPin2 := machine.GP3
	ledPin2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip2 := ws2812.New(ledPin2)

	var currentEyeAnim *Animation
	var currentMouthAnim *Animation
	var queuedEyeAnim *Animation = nil
	var queuedMouthAnim *Animation = nil

	// Initial setup for currentEyeAnim
	if eyeIdleAnim == nil {
		if len(LoadedAnimations) > 0 {
			currentEyeAnim = LoadedAnimations[0]
		} else {
			currentEyeAnim = &Animation{FrameCount: 1, Frames: [][]byte{{}}} // Dummy
		}
	} else {
		currentEyeAnim = eyeIdleAnim
	}

	// Initial setup for currentMouthAnim
	if mouthAnim == nil {
		currentMouthAnim = &Animation{FrameCount: 1, Frames: [][]byte{{}}}
	} else {
		currentMouthAnim = mouthAnim
	}

	var eyeFrameCounter int64
	var mouthFrameCounter int64

	for {
		// Check for new animation command
		select {
		case update := <-animChan:
			if update.Eye != nil {
				queuedEyeAnim = update.Eye
			}
			if update.Mouth != nil {
				queuedMouthAnim = update.Mouth
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

		// Check for queued animations and transition
		if queuedEyeAnim != nil && currentEyeAnim != nil && eyeFrameCounter > 0 && (currentEyeAnim.FrameCount > 0 && eyeFrameCounter%int64(currentEyeAnim.FrameCount) == 0) {
			if queuedEyeAnim != currentEyeAnim { // Only switch if new animation is different
				fmt.Printf("Transitioning Eye to: %s\n", queuedEyeAnim.Name)
				currentEyeAnim = queuedEyeAnim
				eyeFrameCounter = 0
			}
			queuedEyeAnim = nil // Clear the queue
		}

		if queuedMouthAnim != nil && currentMouthAnim != nil && mouthFrameCounter > 0 && (currentMouthAnim.FrameCount > 0 && mouthFrameCounter%int64(currentMouthAnim.FrameCount) == 0) {
			if queuedMouthAnim != currentMouthAnim { // Only switch if new animation is different
				fmt.Printf("Transitioning Mouth to: %s\n", queuedMouthAnim.Name)
				currentMouthAnim = queuedMouthAnim
				mouthFrameCounter = 0
			}
			queuedMouthAnim = nil // Clear the queue
		}

		// Yield to allow other goroutines (like UART) to run if needed
		time.Sleep(10 * time.Millisecond)
	}
}
