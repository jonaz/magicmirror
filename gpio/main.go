package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

func main() {

	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	btn17, err := embd.NewDigitalPin(17)
	if err != nil {
		panic(err)
	}
	btn18, err := embd.NewDigitalPin(18)
	if err != nil {
		panic(err)
	}

	defer btn17.Close()
	defer btn18.Close()

	if err := btn17.SetDirection(embd.In); err != nil {
		panic(err)
	}
	if err := btn18.SetDirection(embd.In); err != nil {
		panic(err)
	}
	btn17.ActiveLow(false)
	btn18.ActiveLow(false)

	motion := make(chan embd.DigitalPin)
	err = btn17.Watch(embd.EdgeBoth, func(btn embd.DigitalPin) {
		motion <- btn
	})
	if err != nil {
		panic(err)
	}
	err = btn18.Watch(embd.EdgeBoth, func(btn embd.DigitalPin) {
		motion <- btn
	})
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	fmt.Println("listening...")
	dur, _ := time.ParseDuration("-5m")
	onTime := time.Now().Add(dur)
	for {
		select {
		case <-c:
			fmt.Println("bye")
			return
		case btn := <-motion:
			direction := getDirection(btn.N())
			v, _ := btn.Read()
			if v == 0 {
				log.Printf("Motion stopped from %s.\n", direction)
				continue
			}
			log.Printf("Motion detected from %s.\n", direction)
			if time.Since(onTime).Seconds() > 300 {
				turnScreenOn()
				onTime = time.Now()
			}

		}
	}
}

func getDirection(id int) string {

	switch id {
	case 17: //left side
		return "left"
	case 18: //right side
		return "right"
	}

	return "unknown"
}

func turnScreenOn() {
	log.Println("turning screen on")
	out, err := exec.Command("sh", "-c", "XAUTHORITY=/home/jonaz/.Xauthority DISPLAY=:0 xset dpms force on").Output()
	fmt.Println(out)
	fmt.Println(err)
}
