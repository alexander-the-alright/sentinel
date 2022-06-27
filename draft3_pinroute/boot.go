// =============================================================================
// Auth: Alex Celani
// File: boot.go
// Revn: 06-23-2022  0.1
// Func: Blink a light when the pi boots (or on login, but who cares)
//
// TODO: create
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
// 06-23-2022: init
//
// =============================================================================

package main


import (
    "github.com/stianeikeland/go-rpio"
    "os"
    "time"
)


/*
Toggles a LED on physical pin 19 (mcu pin 10)
Connect a LED with resistor from pin 19 to ground.
*/


// Use mcu pin 10, corresponds to physical pin 19 on the pi
var pin rpio.Pin


func main() {

    pin = rpio.Pin( 10 )
    // Open and map memory to access gpio, check for errors
    if err := rpio.Open(); err != nil {
        os.Exit(1)
    }

    // Unmap gpio memory when done
    defer rpio.Close()

    // Set pin to output mode
    pin.Output()

    // Toggle pin 20 times
    for i := 0; i < 4; i++ {
        pin.Toggle()
        time.Sleep(time.Second / 10)
    }
}

