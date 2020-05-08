package main

import "time"


const (
    RowCount = 2
    ColumnCount = 4
    LedCount = RowCount * ColumnCount
)


// GPIO definitions.
var _redRowGpios [RowCount]int = [RowCount]int{14, 18}
var _greenRowGpios [RowCount]int = [RowCount]int{15, 23}
var _colGpios [ColumnCount]int = [ColumnCount]int{24, 25, 8, 7}


// TestLeds - Light each LED in turn, to test them all, forever.
func TestLeds(display *LedDisplay) {
    for true {
        for led := 0; led < LedCount; led++ {
            display.Update(led, LedRed)
            time.Sleep(time.Second)
            display.Update(led, LedGreen)
            time.Sleep(time.Second)
            display.Update(led, LedOff)
        }
    }
}


// LED display object.

// CreateDisplay - Create an LED display object.
func CreateDisplay() *LedDisplay {
    var display LedDisplay
//    display.redState = [][]bool{{false}, {false}}
//    display.greenState = [][]bool{{true, true}, {false, true}}
    display.column = 0
    display.updateChannel = make(chan updateInfo, 10)

    go display.run()
    return &display
}


// Update - Send a machine state update to this display.
func (this *LedDisplay) Update(machine int, colour int) {
    update := updateInfo{machine: machine, colour: colour}
    this.updateChannel<- update
}


// LED colours.
const (
    LedOff = iota
    LedRed
    LedYellow
    LedGreen
)


// LedDisplay - LED display object.
type LedDisplay struct {
    redState [RowCount][ColumnCount]bool
    greenState [RowCount][ColumnCount]bool
    column int  // Column we're currently displaying.
    updateChannel chan updateInfo
}


// run - Run the LED display.
// This should be called as a Goroutine.
func (this *LedDisplay) run() {
    // Use a ticker to time the LED refreshes.
    ticker := time.NewTicker(5 * time.Millisecond)

    for {
        select {
        case update := <-this.updateChannel:
            this.setMachineColour(update.machine, update.colour)

        case <-ticker.C:
            // Time to display the next column of LEDs.
            // First turn off the old column, then advance to the new.
            ColumnOn(this.column, false)
            this.column++
            if this.column >= ColumnCount { this.column = 0 }

            // Set all rows, then enable the column.
            for row := 0; row < RowCount; row++ {
                RowOn(true, row, this.redState[row][this.column])
                RowOn(false, row, this.greenState[row][this.column])
            }

            ColumnOn(this.column, true)
        }
    }
}


// setMachineColour - Set the specified machine LED to the specified colour.
func (this *LedDisplay) setMachineColour(machine int, colour int) {
    // Convert colour to RG values.
    redState := (colour == LedRed ) || (colour == LedYellow)
    greenState := (colour == LedGreen) || (colour == LedYellow)

    // Map machine to LEDs.
    // TODO: Do this better.
    column := machine % ColumnCount
    row := machine / ColumnCount

    // Set LED states.
    this.redState[row][column] = redState
    this.greenState[row][column] = greenState
}


// updateInfo - Status update for a single machine.
type updateInfo struct {
    machine int
    colour int
}



// Grid level LED control.

// InitPins - Initialise all required GPIOs.
func InitPins() {
    // Rows are active high, so initialise low for off.
    for _, pin := range _redRowGpios {
        GpioExport(pin)
        GpioDirIn(pin, false)
        GpioWrite(pin, false)
    }

    for _, pin := range _greenRowGpios {
        GpioExport(pin)
        GpioDirIn(pin, false)
        GpioWrite(pin, false)
    }

    // Columns are active low, so initialise high for off.
    for _, pin := range _colGpios {
        GpioExport(pin)
        GpioDirIn(pin, false)
        GpioWrite(pin, true)
    }
}


// RowOn - Turn on the specified row state.
func RowOn(red bool, row int, on bool) {
    // Rows are active high, so no need to invert sense.
    gpio := _greenRowGpios[row]
    if red { gpio = _redRowGpios[row] }

    GpioWrite(gpio, on)
}


// ColumnOn - Set the specified column state.
func ColumnOn(col int, on bool) {
    // Columns are active low, so invert sense.
    GpioWrite(_colGpios[col], !on)
}


// FreePins - Release all pins.
func FreePins() {
    for _, pin := range _redRowGpios {
        GpioUnexport(pin)
    }

    for _, pin := range _greenRowGpios {
        GpioUnexport(pin)
    }

    for _, pin := range _colGpios {
        GpioUnexport(pin)
    }
}
