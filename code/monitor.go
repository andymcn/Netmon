package main

import "encoding/json"
import "fmt"
import "os/exec"
import "time"


// Monitor object.

// CreateMonitor - Create a monitor object and start it running.
func CreateMonitor(config *configDef, display *LedDisplay) *Monitor {
    var p Monitor
    p.powerStatus = map[string]bool{}
    p.pingStatus = map[string]bool{}
    p.leds = config.Leds
    p.display = display
    p.powerChannel = make(chan map[string]bool, 10)

    return &p
}


// Run - Run this monitor.
// Goroutine, never returns.
func (this *Monitor) Run() {

    go this.monitorPower()

    // TODO
    monitorMachines(this.display, this.leds)
}


// Internals.

// Monitor - Monitor class state.
type Monitor struct {
    powerStatus map[string]bool
    pingStatus map[string]bool
    leds []ledDef
    display *LedDisplay

    powerChannel chan map[string]bool  // For reporting power status to collator.
}





// monitorPower - Monitor remote power control and report via the power channel.
// Go routine, never returns.
func (this *Monitor) monitorPower() {
    for true {
        // Try to get current remote power status.
        status := getPower()

        if status == nil {
            // Couldn't get status. Assume all machines are powered on.
            status = map[string]bool{}

            for _, led := range this.leds {
                if led.RemotePower {
                    status[led.Name] = false
                }
            }
        }

        // Report our status.
        this.powerChannel <- status

        // Wait a bit, then do it again.
        time.Sleep(time.Second)
    }
}


// getPower - Get the remote power status of all machines.
func getPower() map[string]bool {
    dest := "power@192.168.1.2"

    out, err := exec.Command("ssh", dest, "power", "status", "-j").Output()
    if err != nil {
        fmt.Printf("Failure to get power status: %v\n", err)
        return nil
    }

    var status powerStatus
    err = json.Unmarshal(out, &status)
    if err != nil {
        fmt.Printf("Failure parsing power status: %v\n", err)
        fmt.Printf("%s\n", string(out))
        return nil
    }

    fmt.Printf("%+v\n", status)
    return status.Power
}


// powerStatus - Format of remote power status raw information.
type powerStatus struct {
    Power map[string]bool `json:"power"`
}


// monitorMachines - Monitor all defined machines, forever.
func monitorMachines(display *LedDisplay, leds []ledDef) {
    for true {
        for i, led := range leds {
            if led.Name != "" {
                ledColour := checkMachine(led.IP, "")
                fmt.Printf("Led %d, %s %s, colour %d\n", i, led.Name, led.IP, ledColour)
                display.Update(i, ledColour)
                time.Sleep(time.Second)
            }
        }
    }
}


// checkMachine - Determine the state of the given machine.
func checkMachine(socIP string, bmcIP string) int {
    if ping(socIP) {
        return LedGreen
    }

    return LedRed//Off
}


// ping - Ping the given IP address to check if we can see it.
func ping(ip string) bool {
    err := exec.Command("ping", "-c", "1", "-i", "1", ip).Run()
    return err == nil

}

