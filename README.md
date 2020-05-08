# Netmon - Hardware server up-status monitor.
![Netmon](https://github.com/andymcn/Netmon)

A hardware device which displays the powered and/or up status of a list of servers.

For each server there are two LEDs, one red and one green.

* Both LEDs off mean the server is remote powered off.
* The red LED on means the server is remote powered on (or is not controlled by remote power), but we cannot talk to it.
* The green LED on means we can talk to the server (via ping).

The hardware is a model 3B Raspberry Pi, with a simple circuit to boost the current to drive the LEDs. Since there are far more LEDs needed than GPIOs available on a Pi, we use a simple constant time multiplex grid to drive the LEDs.

The software is a simple Go program. Goroutines are used to check the remote power state and server connectivity in parallel, plus an extra Goroutine to drive the LED grid.
