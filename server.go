package main

import (
	//"encoding/json"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-martini/martini"
	"github.com/jonaz/astrotime"
	"github.com/jonaz/gosmhi"
	"github.com/olahol/melody"
)

var Shutdown = make(chan bool)
var (
	port = flag.String("port", "8080", "server port")
)

func main() {
	flag.Parse()
	initOauth()

	//clients = newClients()

	m := martini.Classic()
	mel := melody.New()

	mel.HandleConnect(func(s *melody.Session) {
		doPeriodicalStuff(mel)
	})

	gracefulShutdown := NewGracefulShutdown(10 * time.Second)
	gracefulShutdown.RunOnShutDown(mel.Close)
	m.Use(gracefulShutdown.Handler)

	//limiter := &ConnectionLimit{limit: 2}
	//m.Use(limiter.Handler)

	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Get("/control/:action", func(p martini.Params) {
		if val, err := json.Marshal(&Message{p["action"], nil}); err != nil {
			mel.Broadcast(val)
		}
		//clients.messageOtherClients(&Message{p["action"], nil})
	})
	m.Get("/getsmhi", getSmhi)
	m.Get("/api/sun", func() string {
		return getSun("rise") + getSun("set")
	})
	//m.Get("/websocket", sockets.JSON(Message{}), websocketRoute)
	m.Get("/websocket", func(w http.ResponseWriter, r *http.Request) {
		doPeriodicalStuff(mel)
		mel.HandleRequest(w, r)
	})

	//OAUTH2
	m.Get("/oauthsetup", handleSetupOauth)
	m.Get("/oauthredirect", handleOauthRedirect)

	m.Get("/cal", func() interface{} {
		return getEvents(6)
	})

	initPeriodicalPush(mel)

	go func() {
		m.RunOnAddr(":" + *port)

	}()

	err := gracefulShutdown.WaitForSignal(syscall.SIGTERM, syscall.SIGINT)
	if err != nil {
		log.Println(err)
	}

}

// do periodical stuff and push over websocket to all.
func initPeriodicalPush(mel *melody.Melody) {
	//this runs doPerdoPeriodicalStuff() every 15 minutes!
	ticker := time.NewTicker(15 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				doPeriodicalStuff(mel)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
func doPeriodicalStuff(mel *melody.Melody) {
	if val, err := json.Marshal(&Message{"temp", getTemp()}); err == nil {
		mel.Broadcast(val)
	}
	if val, err := json.Marshal(&Message{"sunset", getSun("set")}); err == nil {
		mel.Broadcast(val)
	}
	if val, err := json.Marshal(&Message{"sunrise", getSun("rise")}); err == nil {
		mel.Broadcast(val)
	}
	if val, err := json.Marshal(&Message{"weather", getSmhi()}); err == nil {
		mel.Broadcast(val)
	}
	//clients.messageOtherClients(&Message{"temp", getTemp()})
	//clients.messageOtherClients(&Message{"sunset", getSun("set")})
	//clients.messageOtherClients(&Message{"sunrise", getSun("rise")})
	//clients.messageOtherClients(&Message{"weather", getSmhi()})
	//clients.messageOtherClients(&Message{"calendarEvents", getEvents(6)})
}

//Download temp from temperatur.nu.
func getTemp() string {
	resp, err := http.Get("http://www.temperatur.nu/termo/soder/termo.txt")

	if err != nil {
		return "error"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	temp := strings.Split(string(body), ",")
	temp2 := strings.Split(temp[2], ": ")
	temp3 := strings.Split(temp2[1], "&")

	return temp3[0]
}

//sunset/runrise
func getSun(p string) string {

	var t time.Time
	switch p {
	case "set":
		t = astrotime.NextSunset(time.Now(), float64(56.878333), float64(14.809167))
		break
	case "rise":
		t = astrotime.NextSunrise(time.Now(), float64(56.878333), float64(14.809167))
		break
	}

	padHour := ""
	padMinute := ""
	if t.Hour() < 10 {
		padHour = "0"
	}
	if t.Minute() < 10 {
		padMinute = "0"
	}
	ti := padHour + strconv.Itoa(t.Hour()) + ":" + padMinute + strconv.Itoa(t.Minute())
	return ti
	//return "{\"type\":\"sunset\",\"value\":\"" + ti + "\"}"
}

//Weather from SMHI
type day struct {
	Max           float64 `json:"max"`
	Min           float64 `json:"min"`
	Day           string  `json:"day"`
	Weather       int     `json:"weather"`
	Cloud         int     `json:"cloud"`
	Precipitation string  `json:"precipitation"`
}

type response struct {
	Days          []day   `json:"days"`
	WindMax       float64 `json:"windmax"`
	WindMin       float64 `json:"windmin"`
	Weather       int     `json:"weather"`
	Cloud         int     `json:"cloud"`
	Precipitation string  `json:"precipitation"`
}

func getSmhi() response {
	response, err := fetchSmhi()
	if err != nil {
		log.Println(err)
		return response
	}

	return response
}
func fetchSmhi() (response, error) {
	//we will get weather for the next 6 days including today.
	days := make([]day, 6)
	today := time.Now()
	resp := response{Days: days}
	smhi := gosmhi.New()
	smhiResponse, err := smhi.GetByLatLong("56.8769", "14.8092")
	if err != nil {
		return resp, err
	}

	resp.WindMin, _ = smhiResponse.GetMinWindByDate(today)
	resp.WindMax, _ = smhiResponse.GetMaxWindByDate(today)
	resp.Weather = smhiResponse.GetPrecipitationByHour(today)
	resp.Cloud = smhiResponse.GetTotalCloudCoverageByHour(today)
	resp.Precipitation = fmt.Sprintf("%.1f", smhiResponse.GetMeanPrecipitationByDate(today))

	for key := range days {
		days[key].Day = today.Weekday().String()
		days[key].Max, _ = smhiResponse.GetMaxTempByDate(today)
		days[key].Min, _ = smhiResponse.GetMinTempByDate(today)
		days[key].Weather = smhiResponse.GetPrecipitationByDate(today)
		days[key].Cloud = smhiResponse.GetTotalCloudCoverageByDate(today)
		days[key].Precipitation = fmt.Sprintf("%.1f", smhiResponse.GetMeanPrecipitationByDate(today))
		today = today.Add(24 * time.Hour)
	}
	return resp, nil
}
