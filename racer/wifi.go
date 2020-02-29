package main

import (
	"machine"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/tinydraw"

	"tinygo.org/x/drivers/wifinina"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"

	"tinygo.org/x/drivers/net/mqtt"

	"../connect"
	"../game"
)

var (
	spi = machine.NINA_SPI

	// this is the ESP chip that has the WIFININA firmware flashed on it
	adaptor = &wifinina.Device{
		SPI:   spi,
		CS:    machine.NINA_CS,
		ACK:   machine.NINA_ACK,
		GPIO0: machine.NINA_GPIO0,
		RESET: machine.NINA_RESETN,
	}

	console = machine.UART0

	cl      mqtt.Client
	payload []byte
	enabled bool

	status = game.Looking
)

func updateTrackInfo(client mqtt.Client, msg mqtt.Message) {
	b := msg.Payload()
	if len(b) == 0 {
		println("no data")
		return
	}
	var speed int16
	speed = 0

	speedGaugeNeedle(speed, colors[BLACK])
	speedGaugeNeedle(speed, colors[player])
	oldSpeed = speed

	data := strings.Split(string(b), ",")
	if len(data) != 2 {
		// something wrong
		println("data too short")
		return
	}

	position, _ := strconv.Atoi(data[0])
	laps, _ := strconv.Atoi(data[1])

	resetLapBar()
	progressLapBar(int16(position))
	progressRaceBar(int16(laps))
}

func configureWifi(player int) {
	display.FillScreen(colors[BACKGROUND])

	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		MOSI:      machine.SPI0_MOSI_PIN,
		MISO:      machine.SPI0_MISO_PIN,
		SCK:       machine.SPI0_SCK_PIN,
	})

	// Init esp8266/esp32
	adaptor.Configure()
	connectToAP()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(connect.Broker).SetClientID("tinygo-racer-" + strconv.Itoa(player))

	println("Connecting to MQTT broker at", connect.Broker)
	cl = mqtt.NewClient(opts)
	if token := cl.Connect(); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
		tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 80, []byte(token.Error().Error()), colors[PLAYER1])
	}

	// subscribe
	setupSubs()

	enabled = true

	go heartbeat()

	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 100, []byte("Done."), colors[PLAYER2])
	println("Done.")
}

// connect to access point
func connectToAP() {
	time.Sleep(2 * time.Second)
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 30, []byte("Connecting to '"+connect.SSID+"'"), colors[WHITE])
	println("Connecting to " + connect.SSID)
	adaptor.SetPassphrase(connect.SSID, connect.PASS)
	for st, _ := adaptor.GetConnectionStatus(); st != wifinina.StatusConnected; {
		display.FillRectangle(0, 31, 320, 12, colors[BACKGROUND])
		tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 40, []byte(st.String()), colors[PLAYER1])
		println("Connection status: " + st.String())
		time.Sleep(1000 * time.Millisecond)
		st, _ = adaptor.GetConnectionStatus()
	}
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 50, []byte("Connected :D"), colors[PLAYER2])
	println("Connected.")
	time.Sleep(2 * time.Second)
	ip, _, _, err := adaptor.GetIP()
	for ; err != nil; ip, _, _, err = adaptor.GetIP() {
		display.FillRectangle(0, 51, 320, 12, colors[BACKGROUND])
		tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 60, []byte(err.Error()), colors[PLAYER1])
		println("IP", err.Error())
		time.Sleep(1 * time.Second)
	}
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 70, []byte("IP: "+ip.String()), colors[PLAYER2])
	println(ip.String())
}

func failMessage(msg string) {
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 10, 200, []byte(msg), colors[PLAYER1])
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 10, 220, []byte("0xF1 ErrUnknowHost / 0xF3 ErrConnectionTimeout"), colors[STEPR])
	tinydraw.Rectangle(display, 4, 186, 312, 40, colors[PLAYER1])
	for {
		println(msg)
		time.Sleep(100 * time.Second)
	}
}

func tap() {
	topic := strings.Replace(game.TopicRacerRacing, "+", "1", 1)

	if token := cl.Publish(topic, 0, false, []byte(strconv.Itoa(1))); token.Wait() && token.Error() != nil {
		println(token.Error().Error())
	}
}

func heartbeat() {
	for {
		if status != game.Start {
			topic := strings.Replace(game.TopicRacerAvailable, "+", "1", 1)

			if token := cl.Publish(topic, 0, false, []byte("available")); token.Wait() && token.Error() != nil {
				println(token.Error().Error())
			}
		}

		time.Sleep(time.Millisecond * 1000)
	}
}

func setupSubs() {
	if token := cl.Subscribe(game.TopicRaceAvailable, 0, handleRaceAvailable); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRaceStarting, 0, handleRaceStarting); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRaceCountdown, 0, handleRaceCountdown); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRaceStart, 0, handleRaceStart); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRaceOver, 0, handleRaceOver); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRaceWinner, 0, handleRaceWinner); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRacerPosition, 0, updateTrackInfo); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}
}

func handleRaceAvailable(client mqtt.Client, msg mqtt.Message) {
	status = game.Available
}

func handleRaceStarting(client mqtt.Client, msg mqtt.Message) {
	status = game.Starting
}

func handleRaceCountdown(client mqtt.Client, msg mqtt.Message) {
	status = game.Countdown
}

func handleRaceStart(client mqtt.Client, msg mqtt.Message) {
	status = game.Start
}

func handleRaceOver(client mqtt.Client, msg mqtt.Message) {
	status = game.Over
}

func handleRaceWinner(client mqtt.Client, msg mqtt.Message) {
	status = game.Winner
}
