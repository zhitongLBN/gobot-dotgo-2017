package main

import (
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/api"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/firmata"
)

var button *gpio.GroveButtonDriver
var blue *gpio.GroveLedDriver
var green *gpio.GroveLedDriver
var red *gpio.GroveLedDriver
var buzzer *gpio.GroveBuzzerDriver
var touch *gpio.GroveTouchDriver
var rotary *aio.GroveRotaryDriver
var sensor *aio.GroveTemperatureSensorDriver
var sound *aio.GroveSoundSensorDriver
var light *aio.GroveLightSensorDriver
var screen *i2c.GroveLcdDriver

func DetectSound(level int) {
	if level >= 400 {
		Message("Sound detected")
		TurnOff()
		Blue()
		time.Sleep(1 * time.Second)
		Reset()
	}
}

func DetectLight(level int) {
	if level >= 700 {
		Message("Light detected")
		TurnOff()
		Blue()
		time.Sleep(1 * time.Second)
		Reset()
	}
}

func CheckFireAlarm() {
	temp := sensor.Temperature()
	msg := fmt.Sprintf("Temp: %v", temp)
	Message(msg)
	if temp >= 40 {
		TurnOff()
		Red()
		buzzer.Tone(gpio.F4, gpio.Half)
	}
}

func Doorbell() {
	TurnOff()
	Blue()
	buzzer.Tone(gpio.C4, gpio.Quarter)
	time.Sleep(1 * time.Second)
	Reset()
}

func TurnOff() {
	red.Off()
	blue.Off()
	green.Off()
}

func Reset() {
	TurnOff()
	Message("Airlock ready.")
	Green()
}

func Message(text string) {
	fmt.Println(text)
	screen.Clear()
	screen.Home()
	screen.Write(text)
}

func Blue() {
	blue.On()
	screen.SetRGB(0, 0, 255) // blue
}

func Red() {
	red.On()
	screen.SetRGB(255, 0, 0) // red
}

func Green() {
	green.On()
	screen.SetRGB(0, 255, 0) // green
}

func main() {
	master := gobot.NewMaster()

	a := api.NewAPI(master)
	a.Start()

	board := firmata.NewAdaptor(os.Args[1])

	// digital
	button = gpio.NewGroveButtonDriver(board, "2")
	blue = gpio.NewGroveLedDriver(board, "3")
	green = gpio.NewGroveLedDriver(board, "4")
	red = gpio.NewGroveLedDriver(board, "5")
	buzzer = gpio.NewGroveBuzzerDriver(board, "6")
	touch = gpio.NewGroveTouchDriver(board, "8")

	// analog
	rotary = aio.NewGroveRotaryDriver(board, "0")
	sensor = aio.NewGroveTemperatureSensorDriver(board, "1")
	sound = aio.NewGroveSoundSensorDriver(board, "2")
	light = aio.NewGroveLightSensorDriver(board, "3")

	// i2c
	screen = i2c.NewGroveLcdDriver(board)

	work := func() {
		Reset()

		button.On(gpio.ButtonPush, func(data interface{}) {
			TurnOff()
			Message("On!")
			Blue()
		})

		button.On(gpio.ButtonRelease, func(data interface{}) {
			Reset()
		})

		touch.On(gpio.ButtonPush, func(data interface{}) {
			Doorbell()
		})

		rotary.On(aio.Data, func(data interface{}) {
			b := uint8(
				gobot.ToScale(gobot.FromScale(float64(data.(int)), 0, 4096), 0, 255),
			)
			blue.Brightness(b)
		})

		sound.On(aio.Data, func(data interface{}) {
			DetectSound(data.(int))
		})

		light.On(aio.Data, func(data interface{}) {
			DetectLight(data.(int))
		})

		gobot.Every(1*time.Second, func() {
			CheckFireAlarm()
		})
	}

	robot := gobot.NewRobot("airlock",
		[]gobot.Connection{board},
		[]gobot.Device{button, blue, green, red, buzzer, touch, rotary, sensor, sound, light, screen},
		work,
	)

	master.AddRobot(robot)

	master.Start()
}
