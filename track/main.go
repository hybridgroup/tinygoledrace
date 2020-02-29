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

	cl       mqtt.Client
	ledstrip *apa102.Device
	leds     [game.TrackLength]color.RGBA
	ledIndex uint8
)

func main() {
	time.Sleep(3000 * time.Millisecond)

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	//go handleDisplay()

	rand.Seed(time.Now().UnixNano())

	// Configure SPI0 for 500K, Mode 0
	spi0.Configure(machine.SPIConfig{
		Frequency: 500000,
		Mode:      0,
	})

	a := apa102.New(spi0)
	ledstrip = &a
	//leds = make([]color.RGBA, game.TrackLength)

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
	println("Connected")

	// subscribe
	setupSubs()

	go heartbeat()

	handleLED()
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
		case game.Start:
			// draw racers
			for i := range leds {
				switch {
				case (i == racer1.Pos) && (i == racer2.Pos):
					leds[i] = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
				case i == racer1.Pos:
					leds[i] = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
				case i == racer2.Pos:
					leds[i] = color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}
				default:
					leds[i] = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}
				}
			}
		case game.Over:
			// excite visual
		}

		ledstrip.WriteColors(leds[:])
		time.Sleep(100 * time.Millisecond)
	}
}

func clearTrack() {
	for i := range leds {
		leds[i] = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}
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

		time.Sleep(500 * time.Millisecond)
	}
}

func heartbeat() {
	for {
		//if status == game.Looking || status == game.Available {
		if token := cl.Publish(game.TopicTrackAvailable, 0, false, []byte("")); token.Wait() && token.Error() != nil {
			println("heartbeat:", token.Error().Error())
		}
		//}
		time.Sleep(time.Second * 5)
	}
}

func setupSubs() {
	if token := cl.Subscribe(game.TopicRaceAvailable, 0, handleRaceAvailable); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRaceStart, 0, handleRaceStart); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRaceOver, 0, handleRaceOver); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	if token := cl.Subscribe(game.TopicRacerPosition, 0, handleRacing); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}
}

func handleRaceAvailable(client mqtt.Client, msg mqtt.Message) {
	status = game.Available
}

func handleRaceStart(client mqtt.Client, msg mqtt.Message) {
	status = game.Start
}

func handleRaceOver(client mqtt.Client, msg mqtt.Message) {
	status = game.Over
}

func handleRacing(client mqtt.Client, msg mqtt.Message) {
	// use msg.Topic() to determine which racer aka el[2]
	el := strings.Split(msg.Topic(), "/")
	if len(el) < 4 {
		println("topic too short")
		return
	}

	data := strings.Split(string(msg.Payload()), ",")
	if len(data) != 2 {
		println("data too short")
		return
	}

	pos, err := strconv.Atoi(data[0])
	if err != nil {
		println(err)
		return
	}

	if pos >= game.TrackLength {
		pos = game.TrackLength
	}

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
		println("fail:", msg)
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
