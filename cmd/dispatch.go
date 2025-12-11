package cmd

import (
	"fmt"
	"machine"
	"math/rand"
	"time"
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
	workers := []Address{Worker_0, Worker_1, Worker_2, Worker_3}

	for _, w := range workers {
		go func(workerAddr Address) {
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerAddr)))

			// Init to Idle
			uartChan <- [4]byte{byte(workerAddr), byte(Cmd_DisplayAnim), byte(Anim_EyeIdle), byte(Anim_MouthIdle)}

			for {
				// Random sleep 2-6 seconds
				sleepDuration := time.Duration(2000+r.Intn(4001)) * time.Millisecond
				time.Sleep(sleepDuration)

				// Send Blink
				// fmt.Printf("Blinking Worker %d\n", workerAddr)
				uartChan <- [4]byte{byte(workerAddr), byte(Cmd_DisplayAnim), byte(Anim_EyeBlink), byte(Anim_MouthIdle)}

				// Blink duration (approx)
				time.Sleep(200 * time.Millisecond)

				// Send Idle
				uartChan <- [4]byte{byte(workerAddr), byte(Cmd_DisplayAnim), byte(Anim_EyeIdle), byte(Anim_MouthIdle)}
			}
		}(w)
	}

	// Block forever
	select {}
}
