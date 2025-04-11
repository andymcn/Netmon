package main

import "fmt"
import "os/exec"
import "time"


// Monitor object.

// CreateMonitor - Create a monitor object and start it running.
func CreateMonitor(config *configDef, verbose bool) *Monitor {
    var p Monitor
    p.ledCount = len(config.Leds)
    p.verbose = verbose
    p.offRedSec = config.OffRedSec * time.Second
    p.pingDelaySec = config.PingDelaySec
    p.pingChannel = make(chan pingInfo, 100)

    p.machines = make(map[int]*machineState)
    for i, led := range config.Leds {
        if led.Name != "" {
            var m machineState
            m.name = led.Name
            m.ledNo = i
            m.ip = led.IP
            m.offError = led.OffError
            p.machines[i] = &m
        }
    }

    return &p
}


// Run - Run this monitor.
// Goroutine, never returns.
func (this *Monitor) Run() {
    // First turn on all LEDs for 1 second as a test.
    for i := 0; i < this.ledCount; i++ {
        SetLed(i, LedYellow)
    }

    time.Sleep(time.Second)

    for i := 0; i < this.ledCount; i++ {
        SetLed(i, LedOff)
    }

    // Monitor servers for defined LEDs.
    for i, led := range this.machines {
        go this.monitorServer(i, led.ip)
    }

    // Now process all the resulting info.
    this.collate()
}


// Internals.

// Monitor - Monitor class state.
type Monitor struct {
    // State of machines.
    machines map[int]*machineState  // Indexed by led number.

    // Configuration.
    ledCount int
    verbose bool
    offRedSec time.Duration
    pingDelaySec time.Duration

    // Channel for reporting determined state ready to be collated.
    pingChannel chan pingInfo
}

// The state we have for one machine.
type machineState struct {
    name string
    ledNo int
    ip string
    pingable bool
    offError bool
    offTime time.Time
}

// pingInfo - Ping status information for a single machine.
type pingInfo struct {
    led int
    pingable bool
}


// monitorServer - Monitor whether the specified machine is pingable.
// Goroutine, never returns.
func (this *Monitor) monitorServer(led int, ip string) {
    for true {
        // See if the machine is pingable.
        var status pingInfo
        status.led = led
        status.pingable = ping(ip)

        // Report our status.
        this.pingChannel <- status

        // Wait a bit, then do it again.
        time.Sleep(this.pingDelaySec * time.Second)
    }
}


// ping - Ping the given IP address to check if we can see it.
func ping(ip string) bool {
    err := exec.Command("ping", "-c", "1", "-i", "1", ip).Run()
    return err == nil
}


// collate - Collate together all of our determined information.
// Goroutine, never returns.
func (this *Monitor) collate() {
    for true {
        status := <-this.pingChannel

        // Update the single corresponding LED.
        ledIndex := status.led
        pingable := status.pingable

        machine := this.machines[ledIndex]
        colour := machine.collate(pingable, this.offRedSec, this.verbose)
        SetLed(ledIndex, colour)
    }
}


// collate - Collate information for this machine.
// Returns the colour this machine's LED should be.
func (this *machineState) collate(pingable bool, offRedSec time.Duration, verbose bool) (colour int) {
    if pingable {
        // Machine is responding.
        if verbose { fmt.Printf("Server %s responding\n", this.name) }
        this.pingable = true
        return LedGreen
    }

    // Machine is not responding.
    if this.pingable {
        // We just became unpingable, set off timestamp.
        if verbose { fmt.Printf("Server %s turned off\n", this.name) }
        this.offTime = time.Now()
        this.pingable = false
    }

    if this.offError {
        // Machine not responding is an error.
        if verbose { fmt.Printf("Server %s is off, which is an error\n", this.name) }
        return LedRed
    }

    offForSec := time.Now().Sub(this.offTime)
    if offForSec < offRedSec {
        // Machine was turned off recently.
        if verbose { fmt.Printf("Server %s recently turned off\n", this.name) }
        return LedRed
    }

    // Machine has been off for a while.
    if verbose { fmt.Printf("Server %s not found\n", this.name) }
    return LedOff
}

