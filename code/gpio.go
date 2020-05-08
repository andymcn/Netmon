package main

import "fmt"
import "os"


// Functions to control GPIOs via the /sys/ interface.

// GpioExport - Claim the specified GPIO for use.
// Each exported pin must be unexported again before any other process may use it.
func GpioExport(pin int) {
//    fd, err := os.Open("/sys/class/gpio/export")
    fd, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0)

    if err != nil {
        fmt.Printf("Failure to open export file for %d, %v\n", pin, err)
        return
    }

    _, err = fmt.Fprintf(fd, "%d", pin)

    if err != nil {
        fmt.Printf("Failure to export %d, %v\n", pin, err)
    }

    fd.Close()
}


// GpioUnexport - Release the specified GPIO so other processes can use it.
func GpioUnexport(pin int) {
    fd, err := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, 0)

    if err != nil {
        fmt.Printf("Failure to open unexport file for %d, %v\n", pin, err)
        return
    }

    _, err = fmt.Fprintf(fd, "%d", pin)

    if err != nil {
        fmt.Printf("Failure to unexport %d, %v\n", pin, err)
    }

    fd.Close()
}


// GpioDirIn - Set the direction of the specified GPIO.
func GpioDirIn(pin int, in bool) {
    sysFileName := fmt.Sprintf("/sys/class/gpio/gpio%d/direction", pin)
    fd, err := os.OpenFile(sysFileName, os.O_WRONLY, 0)

    if err != nil {
        fmt.Printf("Failure to set direction %d->%v, %v\n", pin, in, err)
        return
    }

    dir := "out"
    if in { dir = "in" }
    fmt.Fprintf(fd, dir)
    fd.Close()
}


// GpioWrite - Set the specified GPIO to the specified value.
func GpioWrite(pin int, on bool) {
    sysFileName := fmt.Sprintf("/sys/class/gpio/gpio%d/value", pin)
    fd, err := os.OpenFile(sysFileName, os.O_WRONLY, 0)

    if err != nil {
        fmt.Printf("Failure to set value %d->%v, %v\n", pin, on, err)
        return
    }

    value := "0"
    if on { value = "1" }
    fmt.Fprintf(fd, value)
    fd.Close()
}

