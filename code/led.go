package main

import "fmt"
import "os"
import "os/exec"
import "time"


// External API.

// InitLed - Initialise the LED display.
func InitLed(verbose bool) {
    var display LedDisplay
    display.verbose = verbose
    display.commandLink = openCommandLink()
    display.commandChannel = make(chan *ledCommand, 100)

    _display = &display
    go _display.run()
}


// TestLeds - Light each LED in turn, to test them all, forever.
func TestLeds() {
    for true {
//        sendCommand(_display.commandLink, 9, LedRed, false)
//        sendCommand(_display.commandLink, 12, LedRed, false)
//        SetLed(8, LedRed)
//        SetLed(12, LedRed)
//        time.Sleep(2 * time.Second)

//*
        for led := 0; led < LedCount; led++ {
            SetLed(led, LedRed)
            time.Sleep(time.Second / 2)
            SetLed(led, LedGreen)
            time.Sleep(time.Second / 2)
            SetLed(led, LedOff)
//            time.Sleep(time.Second)
        }
//*/
/*
        for led := 0; led < LedCount; led++ {
            display.SetLed(led, LedRed)
            time.Sleep(time.Second)
            display.SetLed(led, LedGreen)
            time.Sleep(time.Second)
            display.SetLed(led, LedOff)
        }
*/
    }
}


// TestAllLeds - Turn on all LEDs, forever.
func TestAllLeds() {
    for led := 0; led < LedCount; led++ {
        SetLed(led, LedYellow)
//        time.Sleep(time.Second)
    }

    // This is the main program thread. We need to keep this going until the above commands
    // have all been sent.
    SetLed(-1, LedOff)  // Shut down.
    select{}
}


// SetLed - Set the specified LED to the given colour.
func SetLed(ledIndex int, colour int) {
    var cmd ledCommand
    cmd.led = ledIndex
    cmd.colour = colour

    _display.commandChannel<- &cmd
}


// LED colours.
const (
    LedOff = 0
    LedRed = 1
    LedGreen = 2
    LedYellow = 3
)


// Internals.

// LedDisplay - LED display object.
type LedDisplay struct {
    verbose bool
    commandChannel chan *ledCommand
    commandLink *os.File
}

// We want the LED display to be a singleton.
var _display *LedDisplay

const (
    LedCount = 40
)


// run - Run the LED display.
// Never returns - should be called as a Goroutine.
func (this *LedDisplay) run() {
    for {
        select {
        case command := <-this.commandChannel:
            if command.led < 0 {
                // This is a shutdown command.
                os.Exit(0)
            }

            this.execCommand(command)
            // TODO: ms delay?
//            time.Sleep(time.Millisecond)
        }
    }
}


// execCommand -  Execute the given LED command.
func (this *LedDisplay) execCommand(command *ledCommand) {
    sendCommand(this.commandLink, command.led, command.colour, this.verbose)
}


// ledCommand - Command to alter a single colour LED.
type ledCommand struct {
    led int  // <0 => close down.
    colour int
}


// openCommandLink - Open our command link to the LED display.
func openCommandLink() *os.File {
    args := []string{"-F", "/dev/serial0", "9600", "cs8", "-onlcr", "-opost"}

    out, err := exec.Command("stty", args...).Output()
    if err != nil {
        fmt.Printf("Error setting option, %v, %v\n", out, err)
        return nil
    }

    f, err := os.OpenFile("/dev/serial0", os.O_RDWR | 0x100 | 0x101000, 0666)
    if err != nil {
        fmt.Printf("Error opening: %v\n", err)
        return nil
    }


//    d := []byte{0x0A, 0x0A, 0x0A, 0x0A}
//    f.Write(d)

    return f
}


// sendCommand - Send an LED command over the given link.
func sendCommand(link *os.File, led int, colour int, verbose bool) {
    cmd := led | (colour << 6)
    d := []byte{byte(cmd)}

    if verbose {
        fmt.Printf("Led %d to %s (%02x)\n", led, _colourNames[colour], d)
    }

    _, err := link.Write(d)
//    fmt.Printf("%d, %v\n", n, err)

    if err != nil {
        fmt.Printf("Erorr sending command %02X, %v\n", cmd, err)
    }
}

var _colourNames []string = []string{"off", "red", "green", "yellow"}

