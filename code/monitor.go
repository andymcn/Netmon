package main

import "encoding/json"
import "fmt"
import "os/exec"
import "time"


// Monitor object.

// CreateMonitor - Create a monitor object and start it running.
func CreateMonitor(config *configDef, verbose bool) *Monitor {
    var p Monitor
    p.verbose = verbose
    p.powerState = make([]bool, LedCount)
    p.pingState = make([]bool, LedCount)
    p.leds = config.Leds
    p.powerIP = config.PowerIP
    p.powerDelaySec = config.PowerDelaySec
    p.pingDelaySec = config.PingDelaySec
    p.powerChannel = make(chan []bool, 10)
    p.pingChannel = make(chan pingInfo, 100)

    return &p
}


// Run - Run this monitor.
// Goroutine, never returns.
func (this *Monitor) Run() {
    // First turn on all LEDs for 1 second as a test.
    for i := 0; i < len(this.leds); i++ {
        SetLed(i, LedYellow)
    }

    time.Sleep(time.Second)

    for i := 0; i < len(this.leds); i++ {
        SetLed(i, LedOff)
    }

    // Monitor power.
    go this.monitorPower()

    // Monitor servers for defined LEDs.
    for i, led := range this.leds {
        if led.Name != "" {
            // This LED is defined.
            go this.monitorServer(i)
        }
    }

    // Now process all the resulting info.
    this.collate()
}


// Internals.

// Monitor - Monitor class state.
type Monitor struct {
    // State we're determined.
    powerState []bool
    pingState []bool

    // Configuration.
    verbose bool
    leds []ledDef
    powerDelaySec time.Duration
    pingDelaySec time.Duration
    powerIP string

    // Channels for reporting determined state ready to be collated.
    powerChannel chan []bool
    pingChannel chan pingInfo
}

// pingInfo - Ping status information for a single machine.
type pingInfo struct {
    led int
    pingable bool
}


// monitorPower - Monitor remote power control and report via the power channel.
// Go routine, never returns.
func (this *Monitor) monitorPower() {
    for true {
        // Get the current remote power status.
        powerMap := getPower(this.powerIP)

        // Extract an LED indexed array from the map we've got.
        // It may be that not all servers are on remote power. For any servers that aren't, we want
        // a red LED if that server is unpingable. We can trivially get that behaviour by treating
        // any server not on remote power as if it were on remote power and turned on. Therefore,
        // we default all remote power channels to be on, unless we're told they're not.
        status := make([]bool, LedCount)
        for i := 0; i < LedCount; i++ {
            status[i] = true
        }

        for name, state := range powerMap {
            if !state {
                // Server is remote powered off, mark its remote power channel as such.
                // Note that there might not be an LED defined for this server, that's fine, we'll
                // just ignore it.
                for i, led := range this.leds {
                    if led.Name == name {
                        // This is the led for this remote power channel.
                        status[i] = false
                        break
                    }
                }
            }
        }

        // Report our status.
        this.powerChannel <- status

        // Wait a bit, then do it again.
        time.Sleep(this.powerDelaySec * time.Second)
    }
}


// getPower - Get the remote power status of all machines.
func getPower(powerIP string) map[string]bool {
    dest := fmt.Sprintf("power@%s", powerIP)

    out, err := exec.Command("ssh", dest, "power", "status", "-j").Output()
    if err != nil {
        fmt.Printf("Failure to get power status: %v\n", err)
        return nil
    }

    var status powerRawInfo
    err = json.Unmarshal(out, &status)
    if err != nil {
        fmt.Printf("Failure parsing power status: %v\n", err)
        fmt.Printf("%s\n", string(out))
        return nil
    }

    return status.Power
}


// powerRawInfo - Format of remote power status raw information.
type powerRawInfo struct {
    Power map[string]bool `json:"power"`
}


// monitorServer - Monitor whether the specified machine is pingable.
// Goroutine, never returns.
func (this *Monitor) monitorServer(led int) {
    for true {
        // See if the machine is pingable.
        var status pingInfo
        status.led = led
        status.pingable = ping(this.leds[led].IP)

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
        select {
        case power := <-this.powerChannel:
            // Update all LEDs.
            this.powerState = power
            for i := 0; i < LedCount; i++ {
                name := this.leds[i].Name
                if name != "" {
                    colour := collateLed(power[i], this.pingState[i], name, this.verbose)
                    SetLed(i, colour)
                }
            }

        case status := <-this.pingChannel:
            // Update the single corresponding LED.
            ledIndex := status.led
            pingable := status.pingable
            this.pingState[ledIndex] = pingable
            name := this.leds[ledIndex].Name
            colour := collateLed(this.powerState[ledIndex], pingable, name, this.verbose)
            SetLed(ledIndex, colour)
        }
    }
}


// collateLed - Determine the colour for an LED based on the given determined information.
func collateLed(power bool, pingable bool, name string, verbose bool) int {
    if pingable {
        // Server is responding.
        if verbose { fmt.Printf("Server %s responding\n", name) }
        return LedGreen
    }

    if power {
        // Server is powered on but not responding.
        if verbose { fmt.Printf("Server %s powered on\n", name) }
        return LedRed
    }

    // Machine is powered off.
    if verbose { fmt.Printf("Server %s not found\n", name) }
    return LedOff
}

