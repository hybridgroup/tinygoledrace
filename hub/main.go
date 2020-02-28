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

func gameAvailable() {
	if status != game.Looking {
		// game already going
		return
	}

	status = game.Available
	broker.Publish(game.TopicRaceAvailable, []byte{})

	// give players 15 seconds to join
	gobot.After(15*time.Second, func() {
		gameStarting()
	})
}

func gameStarting() {
	status = game.Starting
	broker.Publish(game.TopicRaceStarting, []byte{})
	sound.Sound("./audio/space-race-car.mp3")

	// give players 15 seconds before start
	// TODO: countdown
	gobot.After(15*time.Second, func() {
		gameStart()
	})
}

func gameStart() {
	status = game.Start
	broker.Publish(game.TopicRaceStart, []byte{})
	sound.Sound("./audio/space-race-car-2.mp3")
}

func gameOver() {
	status = game.Over
	broker.Publish(game.TopicRaceOver, []byte{})
	sound.Sound("./audio/space-race-pit-stop.mp3")

	gobot.After(5*time.Second, func() {
		stopSounds()
	})
}

func racerJoin(msg mqtt.Message) {
	// only let the racer join if current broadcasting available
	if status != game.Available {
		return
	}

	// use msg.Topic() to determine which racer aka el[2]
	el := strings.Split(msg.Topic(), "/")
	if len(el) < 4 {
		// something wrong
		return
	}

	// notify they are in the race
	racerID := el[2]
	topic := strings.Replace(game.TopicRacerJoin, "+", racerID, 1)
	broker.Publish(topic, []byte{})
}

func handleRacing(msg mqtt.Message) {
	// is the race going?
	if status != game.Start {
		return
	}

	// use msg.Topic() to determine which racer aka el[2]
	el := strings.Split(msg.Topic(), "/")
	if len(el) < 4 {
		// something wrong
		return
	}
	racerID := el[2]

	r, _ := strconv.Atoi(string(msg.Payload()))
	switch racerID {
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
		broker.On(game.TopicRacerJoin, func(msg mqtt.Message) {
			racerJoin(msg)
		})

		broker.On(game.TopicRacerRacing, func(msg mqtt.Message) {
			handleRacing(msg)
		})

		// TODO: push button that starts game
		broker.On(game.TopicHubAvailable, func(msg mqtt.Message) {
			gameAvailable()
		})
	}

	robot := gobot.NewRobot("raceBot",
		[]gobot.Connection{broker, sound},
		work,
	)

	robot.Start()
}
