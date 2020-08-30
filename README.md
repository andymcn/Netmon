# Netmon - Hardware server up-status monitor
![Netmon](https://github.com/andymcn/Netmon)

A hardware device which displays the powered and/or up status of a list of servers, with dedicated LEDs for each server.

Built as a 1U rack mount box supporting up to 40 servers. The front panel is designed with the [Schaeffer tool](https://www.schaeffer-ag.de/en/) and has space for labels to be added to identify the machine each set of LEDs represents.

For each server there are two LEDs, one red and one green.

* Both LEDs off mean the server is remote powered off.
* The red LED on means the server is remote powered on (or is not controlled by remote power), but we cannot talk to it.
* The green LED on means we can ping the server.

The internal hardware is:

1. A model 3B Raspberry Pi, to check on the servers over ethernet.
   This runs a simple Go program. Goroutines are used to check the remote power state and server connectivity in parallel.
   
2. An Atmel ATmega1284P, to drive the LEDs.
   Since there are far more LEDs than GPIOs on either processor, we use a simple constant time multiplex grid to drive the LEDs. The Pi doesn't give accurate enough timing to give a steady, glitch free display, hence we use the Atmel.
   This runs a simple C program. Interrupts are used to time the LED drive and receive instructions from the Pi.

The Pi sends instructions to the Atmel telling it to turn individual LEDs on and off over a serial link. There is no communication in the other direction.
