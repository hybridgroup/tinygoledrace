# TinyGoLEDRace

This is basically a version of https://openledrace.net/ written in Go using TinyGo, Gobot, with an MQTT server.

The racer controllers will be foam mats that have a force sensitive resistor inside them. Each player will run or jump in place on top of their mat to make their racer move.

## How it works

![arch](./images/arch-diagram.png)

## Hardware

### Hub

Raspberry Pi 3 Model 1, with WiFi/Bluetooth
Some kind of speakers
Hub runs the MQTT server, and also the Gobot program with the game sounds/logic.

### Racer

PyPortal
1 connected force sensitive resistors (FSR).
Display

### Track

Arduino Nano33 IoT
Strip of APA102 lights
Button to start the game?

## MQTT protocol

### Hub

`tinygorace/hub/{hubid}/available`

Published by the hub when it is available for a race

`tinygorace/race/{raceid}/starting`

Published by the hub when there is a race getting ready to start

`tinygorace/race/{raceid}/racer/{racerid}/joined`

Published by the hub for each racer when it is ready to join a race

`tinygorace/race/{raceid}/countdown`

data: {count}
published by the hub when there is a race counting down to start

`tinygorace/race/{raceid}/start`

published by the hub when the race starts

`tinygorace/race/{raceid}/end`

published by the hub when the race ends

`tinygorace/race/{raceid}/racer/{racerid}/winner`

published by hub when the race ends to signify the race winner

### Racer

`tinygorace/racer/{racerid}/ready`

published by the racer when it is online but not racing

`tinygorace/racer/{racerid}/join`

data: {raceid}
published by the racer when it is ready to join a race

`tinygorace/race/{raceid}/racer/{racerid}/racing`

data: {about the racer's movement}
published by the racer when it is racing in a race

Subscribes to:

`tinygorace/hub/{hubid}/available`

`tinygorace/race/{raceid}/starting`

`tinygorace/race/{raceid}/countdown`

`tinygorace/race/{raceid}/start`

`tinygorace/race/{raceid}/racing`

`tinygorace/race/{raceid}/end`

### Track

`tinygorace/track/{trackid}/ready`

published by the track when it is ready for a race

Subscribes to:

`tinygorace/hub/{hubid}/available`

`tinygorace/race/{raceid}/starting`

`tinygorace/race/{raceid}/countdown`

`tinygorace/race/{raceid}/start`

`tinygorace/race/{raceid}/end`

and

`tinygorace/race/{raceid}/racer/{racerid}/racing`
