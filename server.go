package main

import (
	"github.com/go-martini/martini"
	"io/ioutil"
	"net/http"
)

func main() {
	m := martini.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Get("/api/temp", getTemp)
	m.Run()
}

//Download temp from temperatur.nu.
func getTemp() []byte {
	resp, err := http.Get("http://www.temperatur.nu/termo/soder/termo.txt")
	defer resp.Body.Close()

	if err != nil {
		return []byte("error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body
}
