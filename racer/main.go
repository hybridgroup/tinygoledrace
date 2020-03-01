package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/touch/resistive"

	"tinygo.org/x/drivers/ili9341"
)

const (
	BLACK = iota
	PLAYER1
	PLAYER2
	PLAYER3
	PLAYER4
	BACKGROUND
	WHITE
	STEPL
	STEPR
)

var display = ili9341.NewParallel(
	machine.LCD_DATA0,
	machine.TFT_WR,
	machine.TFT_DC,
	machine.TFT_CS,
	machine.TFT_RESET,
	machine.TFT_RD,
)

// set this to the player to want to use
var player = 1

var (
	rawspeed    float32
	speed       int16
	oldSpeed    int16
	position    int16
	oldPosition int16
	laps        int16
	oldLaps     int16
)

func main() {
	time.Sleep(3 * time.Second)
	machine.InitADC()
	resistiveTouch.Configure(&resistive.FourWireConfig{
		YP: machine.TOUCH_YD, // y+
		YM: machine.TOUCH_YU, // y-
		XP: machine.TOUCH_XR, // x+
		XM: machine.TOUCH_XL, // x-
	})

	machine.TFT_BACKLIGHT.Configure(machine.PinConfig{machine.PinOutput})

	display.Configure(ili9341.Config{})
	display.SetRotation(ili9341.Rotation270)

	display.FillScreen(colors[PLAYER2])
	machine.TFT_BACKLIGHT.High()

	configureWifi(player)

	handleDisplay()
}
