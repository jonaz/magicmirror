package main

import (
	"github.com/beatrichartz/martini-sockets"
	"github.com/go-martini/martini"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

func main() {

	clients = newClients()

	m := martini.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Get("/api/temp", getTemp)
	m.Get("/websocket", sockets.JSON(Message{}), websocketRoute)
	m.Run()
}

//Download temp from temperatur.nu.
func getTemp() string { // {{{
	resp, err := http.Get("http://www.temperatur.nu/termo/soder/termo.txt")
	defer resp.Body.Close()

	if err != nil {
		return "error"
	}
	body, err := ioutil.ReadAll(resp.Body)

	temp := strings.Split(string(body), ",")
	temp2 := strings.Split(temp[2], ": ")
	temp3 := strings.Split(temp2[1], "&")

	return "{\"temp\":" + temp3[0] + "}"
} // }}}

//WEBSOCKETS:

var clients *Clients

type Message struct {
	Typ  string `json:"typ"`
	From string `json:"from"`
	Text string `json:"text"`
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
			c.out <- &Message{"status", client.Name, "Joined this chat"}
		}
	}
	r.Unlock()
}

// Message all the other clients in the same room
func (r *Clients) messageOtherClients(client *Client, msg *Message) {
	r.Lock()
	msg.From = client.Name

	for _, c := range r.clients {
		if c != client {
			c.out <- msg
		}
	}
	defer r.Unlock()
}

// Remove a client from a room
func (r *Clients) removeClient(client *Client) {
	r.Lock()
	defer r.Unlock()

	for index, c := range r.clients {
		if c == client {
			r.clients = append(r.clients[:index], r.clients[(index+1):]...)
		} else {
			c.out <- &Message{"status", client.Name, "Left this chat"}
		}
	}
}

func newClients() *Clients {
	return &Clients{sync.Mutex{}, make([]*Client, 0)}
}
func websocketRoute(params martini.Params, receiver <-chan *Message, sender chan<- *Message, done <-chan bool, disconnect chan<- int, err <-chan error) (int, string) {
	client := &Client{params["clientname"], receiver, sender, done, err, disconnect}
	clients.appendClient(client)

	// A single select can be used to do all the messaging
	for {
		select {
		case <-client.err:
			// Don't try to do this:
			// client.out <- &Message{"system", "system", "There has been an error with your connection"}
			// The socket connection is already long gone.
			// Use the error for statistics etc
		case msg := <-client.in:
			clients.messageOtherClients(client, msg)
		case <-client.done:
			clients.removeClient(client)
			return 200, "OK"
		}
	}
}
