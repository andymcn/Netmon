package main

import "fmt"
import "os"
import "syscall"
import "unsafe"


// Functions to control GPIOs via direct memory mapped registers.

// GpioInit - Initialise GPIO control.
func GpioInit() {
    memFd, err := os.OpenFile("/dev/mem", os.O_RDWR, 0)

    if err != nil {
        fmt.Printf("Can't open /dev/mem, %v\n", err)
        os.Exit(2)
    }

    mmap, err := syscall.Mmap(int(memFd.Fd()), gpioBase, blockSize,
        syscall.PROT_READ | syscall.PROT_WRITE, syscall.MAP_SHARED)

    if err != nil {
        fmt.Printf("Couldn't mmap, %v\n", err)
        os.Exit(2)
    }

    _gpioMem = (*[mapSize]int)(unsafe.Pointer(&mmap[0]))

    memFd.Close()
}


// GpioDirIn - Set the direction of the specified GPIO.
func GpioDirIn(pin uint, in bool) {
    if in {
        _gpioMem[pin / 10] &= ^(7 << (3 * (pin % 10)))
    } else {
        // Have to set pin to in before setting it to out.
        _gpioMem[pin / 10] &= ^(7 << (3 * (pin % 10)))
        _gpioMem[pin / 10] |= 1 << (3 * (pin % 10))
    }
}


// GpioWrite - Set the specified GPIO to the given state.
func GpioWrite(pin uint, on bool) {
    if on {
        _gpioMem[7] =  1 << pin
    } else {
        _gpioMem[10] =  1 << pin
    }
}


// Internals.

const bcm2708Peribase = 0x3F000000  // Raspberry Pi 3.
const gpioBase = bcm2708Peribase + 0x200000
const mapSize = 1024
const blockSize = 4 * mapSize

var _gpioMem *[mapSize]int

