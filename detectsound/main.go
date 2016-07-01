package main

import (
	"bufio"
	"time"
	"strings"
	"strconv"
	"log"
	"flag"
	"regexp"
	"os/exec"
)
var(
debug bool
level int
)

func main() {
	flag.BoolVar(&debug, "d", false, "Debug: Print volume level")
	flag.IntVar(&level, "l", 10, "Level threshold")
	flag.Parse()

	for{
		run()
	}

}

func run(){

	//arecord -c 2 -d 0 -f S16_LE -vvv /dev/null
	cmd := exec.Command("arecord", "-c", "2", "-d", "0", "-f", "S16_LE", "-vvv", "/dev/null")
	stdout, err := cmd.StderrPipe()
	if err != nil {
		log.Println(err)
	}

	scanner := bufio.NewScanner(stdout)
	go func() {
		re := regexp.MustCompile(`\d+%`)
		dur, _ := time.ParseDuration("-1m")
		onTime := time.Now().Add(dur)
		for scanner.Scan() {
			t := scanner.Text()
			//fmt.Println(t)
			matches := re.FindAllString(t, -1)
			
			if len(matches) < 1 {
				continue
			}

			i , err := strconv.Atoi(strings.TrimRight(matches[0],"%"))
			if err != nil{
				log.Println(t)
				log.Println(err)
				continue
			}

			if debug {
				log.Println("level:", i)
			}
			if i > level {
				log.Println("SOUND DETECTED", i)
				//Only allow turning on screen every minute
				if  time.Since(onTime).Seconds() > 60 {
					turnScreenOn()
					onTime = time.Now()
				}
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		log.Println(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Println(err)
	}
}

func turnScreenOn() {
	log.Println("turning screen on")
	//TODO use vcgencmd display_power 0 and 1 instead since DPMS does not work on raspberry. Only blanking with will wear out the display anyways
	out, err := exec.Command("sh", "-c", "XAUTHORITY=/home/jonaz/.Xauthority DISPLAY=:0 xset dpms force on").Output()
	if len(out) > 0 {
		log.Println(out)
	}

	if err != nil{
		log.Println(err)
	}
}
