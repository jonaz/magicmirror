package main

import (
	"github.com/go-martini/martini"
	"io/ioutil"
	"net/http"
	"strings"
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
func getTemp() string {
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
}
