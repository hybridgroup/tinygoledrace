package main

import (
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/audio"
	"gobot.io/x/gobot/platforms/mqtt"

	"../game"
)

var mqttAdaptor *mqtt.Adaptor
var soundAdaptor *audio.Adaptor

var bgSound chan bool
var gameState string

func gameStarting() {
	soundAdaptor.Sound("./audio/space-race-car.mp3")
}

func gameStart() {
	soundAdaptor.Sound("./audio/space-race-car-2.mp3")
}

func gameOver() {
	soundAdaptor.Sound("./audio/space-race-pit-stop.mp3")

	gobot.After(5*time.Second, func() {
		stopSounds()
	})
}

func stopSounds() {
	bgSound <- false
}

func main() {
	host := os.Args[1]

	mqttAdaptor = mqtt.NewAdaptor(host, "hub")
	soundAdaptor = audio.NewAdaptor()

	work := func() {
		mqttAdaptor.On(game.TopicRaceStarting, func(msg mqtt.Message) {
			gameStarting()
		})

		mqttAdaptor.On(game.TopicRaceStart, func(msg mqtt.Message) {
			gameStart()
		})
	}

	robot := gobot.NewRobot("hubBot",
		[]gobot.Connection{mqttAdaptor, soundAdaptor},
		work,
	)

	robot.Start()
}
