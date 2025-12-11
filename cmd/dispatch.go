package cmd

import (
	"fmt"
	"machine"
	"math/rand"
	"time"
)

const (
	ProjectedFPS = 47
)

// Reworked Protocol to support 2 animation channels (eye + mouth)
// Example: [address, command, animID_eye, animID_mouth]

func RunDispatcher(config Settings, uart *machine.UART, led machine.Pin) {
	fmt.Println("Starting Dispatcher Loop")

	// Channel to serialize UART writes
	// stores [4]byte{address, command, animID_eye, animID_mouth}
	uartChan := make(chan [4]byte, 10)

	// Goroutine to handle UART writes
	go func() {
		for packet := range uartChan {
			uart.Write(packet[:])
			// Small delay between writes to ensure receiver can keep up if needed
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Workers to control
	var workers []Address = []Address{Worker_0, Worker_1}

	r := rand.New(rand.NewSource(time.Now().UnixNano())) // Single random source for the dispatcher

	// Initialize all workers to Idle
	for _, workerAddr := range workers {
		uartChan <- [4]byte{byte(workerAddr), byte(Cmd_DisplayAnim), byte(Anim_EyeIdle), byte(Anim_MouthIdle)}
	}

	for {
		// Calculate a single random sleep duration for all workers
		sleepDuration := time.Duration(2000+r.Intn(4001)) * time.Millisecond
		time.Sleep(sleepDuration)

		// Send Blink command to all workers
		for _, workerAddr := range workers {
			uartChan <- [4]byte{byte(workerAddr), byte(Cmd_DisplayAnim), byte(Anim_EyeBlink), byte(Anim_MouthIdle)}
			fmt.Println("Sent to worker", workerAddr, "to blink")
		}

		// Blink duration (50 frames at ProjectedFPS)
		blinkDuration := time.Duration(50*1000/ProjectedFPS) * time.Millisecond
		time.Sleep(blinkDuration)

		// Send Idle command to all workers
		for _, workerAddr := range workers {
			uartChan <- [4]byte{byte(workerAddr), byte(Cmd_DisplayAnim), byte(Anim_EyeIdle), byte(Anim_MouthIdle)}
		}
	}
	// The `RunDispatcher` function now contains an infinite loop, so `select {}` is not needed.
}
