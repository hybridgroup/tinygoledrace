package main

import (
	"fmt"
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
	racers map[string]*game.Racer

	broker *mqtt.Adaptor
	sound  *audio.Adaptor

	bgSound chan bool
)

func gameAvailable() {
	if status != game.Looking {
		// game already going
		return
	}

	fmt.Println("race available")
	status = game.Available
	broker.Publish(game.TopicRaceAvailable, []byte{})

	// give players 15 seconds to join
	gobot.After(15*time.Second, func() {
		gameStarting()
	})
}

func gameStarting() {
	fmt.Println("race starting")
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
	fmt.Println("race started")
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
	racers[racerID] = &game.Racer{}
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
		fmt.Println("invalid racing topic")
		return
	}
	racerID := el[2]

	r, err := strconv.Atoi(string(msg.Payload()))
	if err != nil {
		fmt.Println(err)
		return
	}

	racers[racerID].Pos += r
	if racers[racerID].Pos > game.TrackLength {
		racers[racerID].Pos -= game.TrackLength
		racers[racerID].Laps++

		// check for winner
		if racers[racerID].Laps > game.Laps {
			fmt.Println("race over")
			status = game.Over
			broker.Publish(game.TopicRaceOver, []byte(racerID))

			gobot.After(1*time.Second, func() {
				fmt.Println("winner is", racerID)
				status = game.Winner
				broker.Publish(game.TopicRaceWinner, []byte(racerID))

				gobot.After(10*time.Second, func() {
					status = game.Looking
					gameAvailable()
				})
			})
			return
		}

		// send new pos
		topic := strings.Replace(game.TopicRacerPosition, "+", racerID, 1)
		result := strconv.Itoa(racers[racerID].Pos) + "," +
			strconv.Itoa(racers[racerID].Laps)
		broker.Publish(topic, []byte(result))
	}
}

func stopSounds() {
	bgSound <- false
}

func main() {
	host := os.Args[1]

	broker = mqtt.NewAdaptor(host, "hub")
	sound = audio.NewAdaptor()

	racers = make(map[string]*game.Racer)

	work := func() {
		broker.On(game.TopicRacerJoin, func(msg mqtt.Message) {
			fmt.Println("racer joined")
			racerJoin(msg)
		})

		broker.On(game.TopicRacerRacing, func(msg mqtt.Message) {
			fmt.Println("racing data received")
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
