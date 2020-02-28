package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/audio"
	"gobot.io/x/gobot/platforms/mqtt"

	"../game"
)

var (
	status = game.Looking
	racer1 = game.Racer{}
	racer2 = game.Racer{}

	broker *mqtt.Adaptor
	sound  *audio.Adaptor

	bgSound chan bool
)

func gameStarting() {
	sound.Sound("./audio/space-race-car.mp3")
}

func gameStart() {
	sound.Sound("./audio/space-race-car-2.mp3")
}

func gameOver() {
	sound.Sound("./audio/space-race-pit-stop.mp3")

	gobot.After(5*time.Second, func() {
		stopSounds()
	})
}

func handleRacing(msg mqtt.Message) {
	// use msg.Topic() to determine which racer aka el[2]
	el := strings.Split(msg.Topic(), "/")
	if len(el) < 4 {
		// something wrong
		return
	}

	r, _ := strconv.Atoi(string(msg.Payload()))
	switch el[2] {
	case "1":
		racer1.Pos += r
		if racer1.Pos > game.TrackLength {
			racer1.Pos -= game.TrackLength
			racer1.Laps++

			// TODO: check for winner

			// send new pos
			topic := strings.Replace(game.TopicRacerPosition, "+", "1", 1)
			broker.Publish(topic, []byte(strconv.Itoa(racer1.Pos)))
		}
	case "2":
		racer2.Pos += r
		if racer1.Pos > game.TrackLength {
			racer1.Pos -= game.TrackLength
			racer1.Laps++

			// TODO: check for winner

			// send new pos
			topic := strings.Replace(game.TopicRacerPosition, "+", "2", 1)
			broker.Publish(topic, []byte(strconv.Itoa(racer2.Pos)))
		}
	}
}

func stopSounds() {
	bgSound <- false
}

func main() {
	host := os.Args[1]

	broker = mqtt.NewAdaptor(host, "hub")
	sound = audio.NewAdaptor()

	work := func() {
		broker.On(game.TopicRaceStarting, func(msg mqtt.Message) {
			gameStarting()
		})

		broker.On(game.TopicRaceStart, func(msg mqtt.Message) {
			gameStart()
		})

		broker.On(game.TopicRacerRacing, func(msg mqtt.Message) {
			handleRacing(msg)
		})

		broker.On(game.TopicRaceOver, func(msg mqtt.Message) {
			gameOver()
		})
	}

	robot := gobot.NewRobot("hubBot",
		[]gobot.Connection{broker, sound},
		work,
	)

	robot.Start()
}
