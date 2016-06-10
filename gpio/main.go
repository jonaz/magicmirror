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

	motion := make(chan int)
	err = btn17.Watch(embd.EdgeRising, func(btn embd.DigitalPin) {
		motion <- btn.N()
		//log.Println(btn.N(), v)
	})
	if err != nil {
		panic(err)
	}
	err = btn18.Watch(embd.EdgeRising, func(btn embd.DigitalPin) {
		//v,_ := btn.Read()
		motion <- btn.N()
		//log.Println(btn.N(), v)
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
		case q := <-motion:
			switch q {
			case 17: //left side
				log.Printf("Motion detected from left.\n")
				if time.Since(onTime).Seconds() > 300 {
					turnScreenOn()
					onTime = time.Now()
				}
			case 18: //right side
				log.Printf("Motion detected from right.\n")
				if time.Since(onTime).Seconds() > 300 {
					turnScreenOn()
					onTime = time.Now()
				}
			}
		}
	}
}

func turnScreenOn() {
	log.Println("turning screen on")
	out, err := exec.Command("sh", "-c", "XAUTHORITY=/home/jonaz/.Xauthority DISPLAY=:0 xset dpms force on").Output()
	fmt.Println(out)
	fmt.Println(err)
}
