package main

import (
	//"encoding/json"
	//"fmt"
	"flag"
	"github.com/beatrichartz/martini-sockets"
	"github.com/cpucycle/astrotime"
	"github.com/go-martini/martini"
	"github.com/jonaz/gosmhi"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	flag.Parse()
	initOauth()

	clients = newClients()

	m := martini.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Get("/test/:id", testSendWs)
	m.Get("/getsmhi", getSmhi)
	m.Get("/websocket", sockets.JSON(Message{}), websocketRoute)

	//OAUTH2
	m.Get("/oauthsetup", handleSetupOauth)
	m.Get("/oauthredirect", handleOauthRedirect)

	m.Get("/cal", getEvents)

	initPeriodicalPush()

	m.Run()
}

func initPeriodicalPush() {
	//this runs doPerdoPeriodicalStuff() every 15 minutes!
	ticker := time.NewTicker(15 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				doPeriodicalStuff()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func doPeriodicalStuff() {
	clients.messageOtherClients(&Message{"temp", getTemp()})
	clients.messageOtherClients(&Message{"sunset", getSun("set")})
	clients.messageOtherClients(&Message{"sunrise", getSun("rise")})
	clients.messageOtherClients(&Message{"weather", getSmhi()})
	clients.messageOtherClients(&Message{"calendarEvents", getEvents(6)})
}
func testSendWs(p martini.Params) {
	clients.messageOtherClients(&Message{p["id"], "Left this chat"})
}

//Download temp from temperatur.nu.
func getTemp() string { // {{{
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
} // }}}

//sunset/runrise
func getSun(p string) string { // {{{

	var t time.Time
	switch p {
	case "set":
		t = astrotime.NextSunset(time.Now(), float64(56.87697), float64(-14.80918))
		break
	case "rise":
		t = astrotime.NextSunrise(time.Now(), float64(56.87697), float64(-14.80918))
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
} // }}}

//Weather from SMHI
type day struct { // {{{
	Max     float64 `json:"max"`
	Min     float64 `json:"min"`
	Day     string  `json:"day"`
	Weather int     `json:"weather"`
	Cloud   int     `json:"cloud"`
}

type response struct {
	Days    []day   `json:"days"`
	WindMax float64 `json:"windmax"`
	WindMin float64 `json:"windmin"`
	Weather int     `json:"weather"`
	Cloud   int     `json:"cloud"`
}

func getSmhi() response { // {{{
	//we will get weather for the next 6 days including today.
	days := make([]day, 6)
	today := time.Now()
	resp := response{days, 0, 0, 0, 0}
	smhi := gosmhi.New()
	smhiResponse := smhi.GetByLatLong("56.8769", "14.8092")

	resp.WindMin, _ = smhiResponse.GetMinWindByDate(today)
	resp.WindMax, _ = smhiResponse.GetMaxWindByDate(today)
	resp.Weather = smhiResponse.GetPrecipitationByHour(today)
	resp.Cloud = smhiResponse.GetTotalCloudCoverageByHour(today)

	for key, _ := range days {
		days[key].Day = today.Weekday().String()
		days[key].Max, _ = smhiResponse.GetMaxTempByDate(today)
		days[key].Min, _ = smhiResponse.GetMinTempByDate(today)
		days[key].Weather = smhiResponse.GetPrecipitationByDate(today)
		days[key].Cloud = smhiResponse.GetTotalCloudCoverageByDate(today)
		today = today.Add(24 * time.Hour)
	}
	return resp
} // }}}
// }}}

//WEBSOCKETS:

var clients *Clients

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
type Clients struct {
	sync.Mutex
	clients []*Client
}
type Client struct {
	Name       string
	in         <-chan *Message
	out        chan<- *Message
	done       <-chan bool
	err        <-chan error
	disconnect chan<- int
}

// Add a client to a room
func (r *Clients) appendClient(client *Client) {
	r.Lock()
	r.clients = append(r.clients, client)
	for _, c := range r.clients {
		if c != client {
			c.out <- &Message{"status", "Joined this chat"}
		}
	}
	r.Unlock()
}

// Message all the other clients
func (r *Clients) messageOtherClients(msg *Message) {
	r.Lock()
	for _, c := range r.clients {
		c.out <- msg
	}
	defer r.Unlock()
}

// Remove a client
func (r *Clients) removeClient(client *Client) {
	r.Lock()
	defer r.Unlock()

	for index, c := range r.clients {
		if c == client {
			r.clients = append(r.clients[:index], r.clients[(index+1):]...)
		} else {
			c.out <- &Message{"status", "Left this chat"}
		}
	}
}

func newClients() *Clients {
	return &Clients{sync.Mutex{}, make([]*Client, 0)}
}
func websocketRoute(params martini.Params, receiver <-chan *Message, sender chan<- *Message, done <-chan bool, disconnect chan<- int, err <-chan error) (int, string) {
	client := &Client{params["clientname"], receiver, sender, done, err, disconnect}
	clients.appendClient(client)

	doPeriodicalStuff()

	// A single select can be used to do all the messaging
	for {
		select {
		case <-client.err:
			// Don't try to do this:
			// client.out <- &Message{"system", "system", "There has been an error with your connection"}
			// The socket connection is already long gone.
			// Use the error for statistics etc
		case msg := <-client.in:
			//TODO handle request from websocket frontend here.
			clients.messageOtherClients(msg)
		case <-client.done:
			clients.removeClient(client)
			return 200, "OK"
		}
	}
}
