package main

import (
	"image/color"
	"time"

	"../game"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

var colors = []color.RGBA{
	color.RGBA{0, 0, 0, 255},       // BLACK
	color.RGBA{255, 0, 0, 255},     // PLAYER 1
	color.RGBA{0, 255, 0, 255},     // PLAYER 2
	color.RGBA{255, 255, 0, 255},   // PLAYER 3
	color.RGBA{0, 0, 255, 255},     // PLAYER 4
	color.RGBA{0, 0, 0, 255},       // BACKGROUND
	color.RGBA{255, 255, 255, 255}, // WHITE
	color.RGBA{102, 255, 51, 255},  // STEPL
	color.RGBA{255, 153, 51, 255},  // STEPR
}

var needlePoint = [91][2]int16{{64, 0}, {63, 1}, {63, 2}, {63, 3}, {63, 4}, {63, 5}, {63, 6}, {63, 7}, {63, 8}, {63, 10}, {63, 11}, {62, 12}, {62, 13}, {62, 14}, {62, 15}, {61, 16}, {61, 17}, {61, 18}, {60, 19}, {60, 20}, {60, 21}, {59, 22}, {59, 23}, {58, 25}, {58, 26}, {58, 27}, {57, 28}, {57, 29}, {56, 30}, {55, 31}, {55, 31}, {54, 32}, {54, 33}, {53, 34}, {53, 35}, {52, 36}, {51, 37}, {51, 38}, {50, 39}, {49, 40}, {49, 41}, {48, 41}, {47, 42}, {46, 43}, {46, 44}, {45, 45}, {44, 46}, {43, 46}, {42, 47}, {41, 48}, {41, 49}, {40, 49}, {39, 50}, {38, 51}, {37, 51}, {36, 52}, {35, 53}, {34, 53}, {33, 54}, {32, 54}, {32, 55}, {31, 55}, {30, 56}, {29, 57}, {28, 57}, {27, 58}, {26, 58}, {25, 58}, {23, 59}, {22, 59}, {21, 60}, {20, 60}, {19, 60}, {18, 61}, {17, 61}, {16, 61}, {15, 62}, {14, 62}, {13, 62}, {12, 62}, {11, 63}, {10, 63}, {8, 63}, {7, 63}, {6, 63}, {5, 63}, {4, 63}, {3, 63}, {2, 63}, {1, 63}, {0, 64}}

func handleDisplay() {
	for {
		switch status {
		case game.Looking:
			displayMessage("Looking for a race", game.Looking)
		case game.Available:
			displayMessage("Found a race", game.Available)
		case game.Starting:
			displayMessage("Race starting", game.Starting)
		case game.Countdown:
			displayMessage("Get ready...", game.Countdown)
		case game.Start:
			race()
		case game.Over:
			displayMessage("Race over.", game.Over)
		case game.Winner:
			winner()
		}
	}
}

func displayMessage(msg string, mode int) {
	display.FillScreen(colors[BACKGROUND])

	for {
		if status != mode {
			return
		}

		// some display thing here
		tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 50, []byte(msg), colors[PLAYER2])
		time.Sleep(100 * time.Millisecond)
	}
}

func winner() {
	for {
		if status != game.Winner {
			return
		}

		// some display thing here
	}
}

func race() {
	resetDisplay()
	// Both progress bar are 0-100 (0 started lap or race, 100 lap or race completed)
	// resetLapBar reset the lap bar for a new lap
	progressLapBar(0)
	progressRaceBar(0)
	// speedGaugeNeedle is 0-250
	speedGaugeNeedle(0, colors[PLAYER1])

	// STEPS are true|false if they are activated or not
	stepL(true)
	stepR(true)

	tapped := false
	lr := false
	rawspeed = 0

	for {
		if status != game.Start {
			return
		}

		// TAP CONTROL
		point := resistiveTouch.ReadTouchPoint()
		if point.Z>>6 > 100 {
			if !tapped {
				tapped = true
				rawspeed += game.Accelleration

				// TAP event
				tap()

				lr = !lr
				if lr {
					stepL(false)
					stepR(true)
				} else {
					stepL(true)
					stepR(false)
				}
			}
		} else {
			tapped = false
		}

		speed = int16(rawspeed)

		// redraw controls here
		if oldSpeed != speed {
			speedGaugeNeedle(speed, colors[BLACK])
			speedGaugeNeedle(speed, colors[player])
			oldSpeed = speed
		}

		if oldPosition != position {
			resetLapBar()
			progressLapBar(int16(position))
			oldPosition = position
		}

		if oldLaps != laps {
			resetRaceBar()
			progressRaceBar(int16(laps))
			oldLaps = laps
		}

		// handle friction
		rawspeed -= rawspeed * game.Friction

		time.Sleep(100 * time.Millisecond)
	}
}

func resetDisplay() {
	display.FillScreen(colors[BACKGROUND])

	// GAUGE
	speedGauge()
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 67, 136, []byte("SPEED"), colors[WHITE])

	// STEP L
	tinydraw.Rectangle(display, 180, 100, 60, 60, colors[WHITE])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 184, 98, []byte("LEFT"), colors[WHITE])

	// STEP R
	tinydraw.Rectangle(display, 244, 100, 60, 60, colors[WHITE])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 248, 98, []byte("RIGHT"), colors[WHITE])

	// LAP PROGRESS BAR
	tinydraw.Rectangle(display, 8, 178, 304, 18, colors[WHITE])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 12, 175, []byte("LAP"), colors[WHITE])
	// RACE PROGRESS BAR
	tinydraw.Rectangle(display, 8, 218, 304, 18, colors[WHITE])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 12, 215, []byte("RACE"), colors[WHITE])
}

func progressLapBar(progress int16) {
	if progress > game.TrackLength {
		progress = game.TrackLength
	}
	if progress < 0 {
		progress = 0
	}
	display.FillRectangle(10, 180, rescale(progress, 0, game.TrackLength, 0, 300), 14, colors[player])
}

func resetLapBar() {
	display.FillRectangle(10, 180, 300, 14, colors[BACKGROUND])
}

func progressRaceBar(progress int16) {
	if progress > game.Laps {
		progress = game.Laps
	}
	if progress < 0 {
		progress = 0
	}
	display.FillRectangle(10, 220, rescale(progress, 0, game.Laps, 0, 300), 14, colors[player])
}

func resetRaceBar() {
	display.FillRectangle(10, 220, 300, 14, colors[BACKGROUND])
}

func speedGauge() {
	tinydraw.FilledCircle(display, 80, 90, 70, colors[WHITE])
	tinydraw.FilledCircle(display, 80, 90, 66, colors[BACKGROUND])
	tinydraw.FilledTriangle(display, 80, 90, 0, 160, 160, 160, colors[BACKGROUND])
}

func speedGaugeNeedle(speed int16, c color.RGBA) {
	speed -= 35
	if speed < 0 {
		speed -= 2 * speed
		tinydraw.Line(display, 80-needlePoint[speed][0], 90+needlePoint[speed][1], 79, 89, c)
		tinydraw.Line(display, 80-needlePoint[speed][0], 90+needlePoint[speed][1], 80, 90, c)
		tinydraw.Line(display, 80-needlePoint[speed][0], 90+needlePoint[speed][1], 81, 91, c)
	} else if speed >= 0 && speed <= 90 {
		tinydraw.Line(display, 80-needlePoint[speed][0], 90-needlePoint[speed][1], 79, 91, c)
		tinydraw.Line(display, 80-needlePoint[speed][0], 90-needlePoint[speed][1], 80, 90, c)
		tinydraw.Line(display, 80-needlePoint[speed][0], 90-needlePoint[speed][1], 81, 89, c)
	} else if speed > 90 && speed <= 180 {
		speed = 180 - speed
		tinydraw.Line(display, 80+needlePoint[speed][0], 90-needlePoint[speed][1], 79, 89, c)
		tinydraw.Line(display, 80+needlePoint[speed][0], 90-needlePoint[speed][1], 80, 90, c)
		tinydraw.Line(display, 80+needlePoint[speed][0], 90-needlePoint[speed][1], 81, 91, c)
	} else {
		if speed > 250 {
			speed = 250
		}
		speed -= 180
		tinydraw.Line(display, 80+needlePoint[speed][0], 90+needlePoint[speed][1], 79, 91, c)
		tinydraw.Line(display, 80+needlePoint[speed][0], 90+needlePoint[speed][1], 80, 90, c)
		tinydraw.Line(display, 80+needlePoint[speed][0], 90+needlePoint[speed][1], 81, 89, c)
	}
}

func stepL(enabled bool) {
	if enabled {
		display.FillRectangle(182, 102, 56, 56, colors[STEPL])
	} else {
		display.FillRectangle(182, 102, 56, 56, colors[BACKGROUND])
	}
}

func stepR(enabled bool) {
	if enabled {
		display.FillRectangle(246, 102, 56, 56, colors[STEPR])
	} else {
		display.FillRectangle(246, 102, 56, 56, colors[BACKGROUND])
	}
}

// Rescale performs a direct linear rescaling of a number from one scale to another.
// TODO: everything with int to avoid extra computational work.
func rescale(input, fromMin, fromMax, toMin, toMax int16) int16 {
	return int16(float32((input-fromMin))*float32((toMax-toMin))/float32((fromMax-fromMin))) + toMin
}
