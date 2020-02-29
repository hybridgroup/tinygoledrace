// TinyGo track
package main

import (
	"image/color"
	"machine"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/drivers/apa102"
	"tinygo.org/x/drivers/net/mqtt"
	"tinygo.org/x/tinyfont"

	// comes from "tinygo.org/x/tinyfont/freemono"
	freemono "./fonts"
	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/drivers/wifinina"

	"../connect"
	"../game"
)

var (
	status = game.Looking
	racer1 = game.Racer{}
	racer2 = game.Racer{}
)

// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	// these are the default pins for the Arduino Nano33 IoT.
	spi0 = machine.SPI0
	spi1 = machine.NINA_SPI

	// this is the ESP chip that has the WIFININA firmware flashed on it
	adaptor = &wifinina.Device{
		SPI:   spi1,
		CS:    machine.NINA_CS,
		ACK:   machine.NINA_ACK,
		GPIO0: machine.NINA_GPIO0,
		RESET: machine.NINA_RESETN,
	}

	console = machine.UART0

	cl      mqtt.Client
	topicTx = "tinygorace/track/ready"

	ledstrip *apa102.Device
	leds     []color.RGBA
	ledIndex uint8
)

func main() {
	time.Sleep(3000 * time.Millisecond)

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	go handleDisplay()

	rand.Seed(time.Now().UnixNano())

	// Configure SPI0 for 5K, Mode 0
	spi0.Configure(machine.SPIConfig{
		Mode: 0,
	})

	a := apa102.New(spi0)
	ledstrip = &a
	leds = make([]color.RGBA, game.TrackLength)

	// Configure SPI1 for 8Mhz, Mode 0, MSB First
	spi1.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		MOSI:      machine.NINA_MOSI,
		MISO:      machine.NINA_MISO,
		SCK:       machine.NINA_SCK,
	})

	// Init esp8266/esp32
	adaptor.Configure()

	connectToAP()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(connect.Broker).SetClientID("track-" + randomString(10))

	println("Connecting to MQTT broker at", connect.Broker)
	cl = mqtt.NewClient(opts)
	if token := cl.Connect(); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	// subscribe
	setupSubs()

	go handleLED()

	for {
		if status == game.Looking {
			if token := cl.Publish(game.TopicTrackAvailable, 0, false, []byte{}); token.Wait() && token.Error() != nil {
				println(token.Error().Error())
			}
		}

		time.Sleep(time.Millisecond * 1000)
	}
}

func handleLED() {
	for {
		switch status {
		case game.Looking, game.Available:
			for i := range leds {
				leds[i] = getRainbowRGB(uint8((i*256)/game.TrackLength) + ledIndex)
			}
			ledIndex++
		case game.Ready:
			clearTrack()
		case game.Starting:
			// excite visual
		case game.Countdown, game.Start, game.Over:
			clearTrack()

			// draw racers
			leds[racer1.Pos] = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0x77}
			leds[racer2.Pos] = color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0x77}
		case game.Winner:
			// excite visual
		}

		ledstrip.WriteColors(leds)
		time.Sleep(100 * time.Millisecond)
	}
}

func clearTrack() {
	// clear the track
	for i := range leds {
		leds[i] = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x77}
	}
}

func handleDisplay() {
	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: ssd1306.Address_128_32,
		Width:   128,
		Height:  64,
	})

	display.ClearDisplay()

	black := color.RGBA{1, 1, 1, 255}

	for {
		display.ClearBuffer()

		r1 := strconv.Itoa(int(racer1.Pos))
		r2 := strconv.Itoa(int(racer2.Pos))
		msg := []byte("r1: " + r1)
		tinyfont.WriteLine(&display, &freemono.Bold9pt7b, 10, 20, msg, black)

		msg2 := []byte("r2: " + r2)
		tinyfont.WriteLine(&display, &freemono.Bold9pt7b, 10, 40, msg2, black)

		display.Display()

		time.Sleep(100 * time.Millisecond)
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

	if token := cl.Subscribe(game.TopicRacerPosition, 0, handleRacing); token.Wait() && token.Error() != nil {
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

func handleRacing(client mqtt.Client, msg mqtt.Message) {
	// use msg.Topic() to determine which racer aka el[2]
	el := strings.Split(msg.Topic(), "/")
	if len(el) < 4 {
		// something wrong
		return
	}

	data := strings.Split(string(msg.Payload()), ",")
	if len(data) != 2 {
		// something wrong
		println("data too short")
		return
	}

	pos, _ := strconv.Atoi(data[0])
	//laps, _ := strconv.Atoi(data[1])

	switch el[2] {
	case "1":
		racer1.Pos = pos
	case "2":
		racer2.Pos = pos
	}
}

// connect to access point
func connectToAP() {
	time.Sleep(2 * time.Second)
	println("Connecting to " + connect.SSID)
	adaptor.SetPassphrase(connect.SSID, connect.PASS)
	for st, _ := adaptor.GetConnectionStatus(); st != wifinina.StatusConnected; {
		println("Connection status: " + st.String())
		time.Sleep(1 * time.Second)
		st, _ = adaptor.GetConnectionStatus()
	}
	println("Connected.")
	time.Sleep(2 * time.Second)
	ip, _, _, err := adaptor.GetIP()
	for ; err != nil; ip, _, _, err = adaptor.GetIP() {
		println(err.Error())
		time.Sleep(1 * time.Second)
	}
	println(ip.String())
}

// Returns an int >= min, < max
func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate a random string of A-Z chars with len = l
func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}

func getRainbowRGB(i uint8) color.RGBA {
	if i < 85 {
		return color.RGBA{i * 3, 255 - i*3, 0, 255}
	} else if i < 170 {
		i -= 85
		return color.RGBA{255 - i*3, 0, i * 3, 255}
	}
	i -= 170
	return color.RGBA{0, i * 3, 255 - i*3, 255}
}
