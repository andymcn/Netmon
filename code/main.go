package main

import "encoding/json"
import "flag"
import "fmt"
import "io/ioutil"
import "os"
import "time"


func main() {
    test := flag.Bool("test", false, "Test LEDs")
    flag.BoolVar(test, "t", false, "Test LEDs")
    testAll := flag.Bool("testall", false, "Test all LEDs at once")
    flag.BoolVar(testAll, "a", false, "Test all LEDs at once")
    verbose := flag.Bool("verbose", false, "Verbose")
    flag.BoolVar(verbose, "v", false, "Verbose")
    flag.Parse()

    config := readConfig("/root/Netmon/config.json")

    InitLed(*verbose)

    if *test {
        TestLeds()
    } else if *testAll {
        TestAllLeds()
    } else {
        monitor := CreateMonitor(config, *verbose)

        // We don't need the main thread any more, so let the monitor use it.
        monitor.Run()
    }
}


// readConfig - Read in the specified config file.
func readConfig(filePath string) *configDef {
    fmt.Printf("Reading config file %s\n", filePath)
    fileHandle, err := os.Open(filePath)

    if err != nil {
        fmt.Printf("Failure to read config file, %v\n", err)
        os.Exit(1)
    }

    defer fileHandle.Close()

    src, _ := ioutil.ReadAll(fileHandle)
    var config configDef
    err = json.Unmarshal(src, &config)
    if err != nil {
        fmt.Printf("Failure to parse config file, %v\n", err)
        os.Exit(1)
    }

    if len(config.Leds) != LedCount {
        fmt.Printf("Incorrect number of machines in config, found %d, need %d.\n",
            len(config.Leds), LedCount)
        os.Exit(1)
    }

    return &config
}

// configDef - Config file format definition.
type configDef struct {
    PowerIP string `json:"power_ip"`
    PowerDelaySec time.Duration `json:"power_delay_sec"`
    PingDelaySec time.Duration `json:"ping_delay_sec"`
    Leds []ledDef `json:"leds"`
}

// ledDef - Information for a single LED.
type ledDef struct {
    Name string `json:"name"`
    IP string `json:"ip"`
    RemotePower bool `json:"remote_power"`
}

