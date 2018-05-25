package main

import (
	"fmt"
	"os/exec"
	"sync/atomic"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gobot.io/x/gobot/platforms/joystick"
)

type pair struct {
	x float64
	y float64
}

var leftX, leftY, rightX, rightY atomic.Value

const offset = 32767.0

func main() {
	drone := tello.NewDriver("8889")
	joystickAdaptor := joystick.NewAdaptor()
	stick := joystick.NewDriver(joystickAdaptor, "xbox360")

	work := func() {
		mplayer := exec.Command("mplayer", "-fps", "30", "-")
		mplayerIn, _ := mplayer.StdinPipe()
		if err := mplayer.Start(); err != nil {
			fmt.Println(err)
			return
		}

		drone.On(tello.ConnectedEvent, func(data interface{}) {
			fmt.Println("Connected")
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})
		})

		drone.On(tello.VideoFrameEvent, func(data interface{}) {
			pkt := data.([]byte)
			if _, err := mplayerIn.Write(pkt); err != nil {
				fmt.Println(err)
			}
		})
	}

	leftX.Store(float64(0.0))
	leftY.Store(float64(0.0))
	rightX.Store(float64(0.0))
	rightY.Store(float64(0.0))

	stick.On(joystick.StartPress, func(data interface{}) {
		drone.TakeOff()
	})

	stick.On(joystick.BackPress, func(data interface{}) {
		drone.Land()
	})

	stick.On(joystick.YPress, func(data interface{}) {
		fmt.Println("FrontFlip")
		drone.FrontFlip()
	})

	stick.On(joystick.APress, func(data interface{}) {
		fmt.Println("BackFlip")
		drone.BackFlip()
	})

	stick.On(joystick.BPress, func(data interface{}) {
		fmt.Println("RightFlip")
		drone.RightFlip()
	})

	stick.On(joystick.XPress, func(data interface{}) {
		fmt.Println("LeftFlip")
		drone.LeftFlip()
	})

	stick.On(joystick.LeftX, func(data interface{}) {
		val := float64(data.(int16))
		leftX.Store(val)
	})

	stick.On(joystick.LeftY, func(data interface{}) {
		val := float64(data.(int16))
		leftY.Store(val)
	})

	stick.On(joystick.RightX, func(data interface{}) {
		val := float64(data.(int16))
		rightX.Store(val)
	})

	stick.On(joystick.RightY, func(data interface{}) {
		val := float64(data.(int16))
		rightY.Store(val)
	})

	gobot.Every(10*time.Millisecond, func() {
		rightStick := getRightStick()

		switch {
		case rightStick.y < -10:
			drone.Forward(tello.ValidatePitch(rightStick.y, offset))
		case rightStick.y > 10:
			drone.Backward(tello.ValidatePitch(rightStick.y, offset))
		default:
			drone.Forward(0)
		}

		switch {
		case rightStick.x > 10:
			drone.Right(tello.ValidatePitch(rightStick.x, offset))
		case rightStick.x < -10:
			drone.Left(tello.ValidatePitch(rightStick.x, offset))
		default:
			drone.Right(0)
		}
	})

	gobot.Every(10*time.Millisecond, func() {
		leftStick := getLeftStick()
		switch {
		case leftStick.y < -10:
			drone.Up(tello.ValidatePitch(leftStick.y, offset))
		case leftStick.y > 10:
			drone.Down(tello.ValidatePitch(leftStick.y, offset))
		default:
			drone.Up(0)
		}

		switch {
		case leftStick.x > 20:
			drone.Clockwise(tello.ValidatePitch(leftStick.x, offset))
		case leftStick.x < -20:
			drone.CounterClockwise(tello.ValidatePitch(leftStick.x, offset))
		default:
			drone.Clockwise(0)
		}
	})

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{joystickAdaptor},
		[]gobot.Device{drone, stick},
		work,
	)

	robot.Start()
}

func getLeftStick() pair {
	s := pair{x: 0, y: 0}
	s.x = leftX.Load().(float64)
	s.y = leftY.Load().(float64)
	return s
}

func getRightStick() pair {
	s := pair{x: 0, y: 0}
	s.x = rightX.Load().(float64)
	s.y = rightY.Load().(float64)
	return s
}
