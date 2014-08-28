package main

import (
	calendar "code.google.com/p/google-api-go-client/calendar/v3"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/golang/oauth2"
	"github.com/golang/oauth2/google"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getEvents(count int64) []*calendar.Event {

	client, err := getOAuthClient()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	svc, err := calendar.New(client)
	if err != nil {
		fmt.Println(err)
	}
	//c, err := svc.CalendarList.List().Do()
	c, err := svc.Events.List(*calendarId).SingleEvents(true).OrderBy("startTime").TimeMin("2014-08-28T00:00:00+10:00").MaxResults(count).Do()
	//c, err := svc.Events.List(*calendarId).Do()
	if err != nil {
		fmt.Println(err)
	}
	//var buffer bytes.Buffer
	//for _, val := range c.Items {
	//buffer.WriteString(val.Start.DateTime + " : " + val.Summary + "\n")
	//}
	//return buffer.String()
	return c.Items

	//2014-08-29T13:00:00+02:00
}

//TODO sort by time http://stackoverflow.com/questions/23121026/sorting-by-time-time-in-golang
//TODO fix refresh token with google
//cannot fetch access token without refresh token.

var (
	clientId            = flag.String("clientid", "", "OAuth Client ID.  If non-empty, overrides --clientid_file")
	clientIdFile        = flag.String("clientid_file", "clientid.ini", "Name of a file containing just the project's OAuth Client ID from https://code.google.com/apis/console/")
	clientSecret        = flag.String("secret", "", "OAuth Client Secret.  If non-empty, overrides --secret_file")
	clientSecretFile    = flag.String("secret_file", "clientsecret.ini", "Name of a file containing just the project's OAuth Client Secret from https://code.google.com/apis/console/")
	retriveTokenChannel chan string
	clientOptions       *oauth2.Options
	clientToken         *oauth2.Token
	calendarId          = flag.String("calendar", "", "fetch from this calendarId")
)

func getOAuthClient() (*http.Client, error) {
	var err error
	if clientToken == nil {
		cacheFile := tokenCacheFile(clientOptions)
		clientToken, err = tokenFromFile(cacheFile)
		log.Printf("Using cached token %#v from %q", clientToken, cacheFile)
		if err != nil {
			return nil, fmt.Errorf("Cannot find file: %q", cacheFile)
		}
	}

	config, err := google.NewConfig(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	t := config.NewTransport()
	t.SetToken(clientToken)

	return &http.Client{Transport: t}, nil
}
func initOauth() {
	clientOptions = &oauth2.Options{
		ClientID:     valueOrFileContents(*clientId, *clientIdFile),
		ClientSecret: valueOrFileContents(*clientSecret, *clientSecretFile),
		RedirectURL:  "http://localhost:3000/oauthredirect",
		Scopes:       []string{calendar.CalendarReadonlyScope},
		AccessType:   "offline",
	}
}

func handleSetupOauth(w http.ResponseWriter, r *http.Request) {

	retriveTokenChannel = make(chan string)
	if *clientId == "" || *clientSecret == "" {
		fmt.Println("--clientid and --secret must be provided!")
		return
	}

	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	config, err := google.NewConfig(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	url := config.AuthCodeURL("")
	fmt.Printf("Visit the URL for the auth dialog: %v", url)
	http.Redirect(w, r, url, http.StatusFound)
	go waitForAuthCode(config)
}

func waitForAuthCode(config *oauth2.Config) {
	select {
	case authorizationCode := <-retriveTokenChannel:
		t, err := config.NewTransportWithCode(authorizationCode)
		if err != nil {
			log.Fatal(err)
		}
		cacheFile := tokenCacheFile(clientOptions)
		saveToken(cacheFile, t.Token())
	case <-time.After(time.Second * 30):
		fmt.Print("waiting for code timed out")
		return
	}
}

func handleOauthRedirect(req *http.Request) (int, string) {
	if code := req.FormValue("code"); code != "" {
		retriveTokenChannel <- code
		return 200, "<h1>Success</h1>Authorized."
	}
	return 500, "No code!"
}

func valueOrFileContents(value string, filename string) string {
	if value != "" {
		return value
	}
	slurp, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Error reading %q: %v", filename, err)
	}
	return strings.TrimSpace(string(slurp))
}
func osUserCacheDir() string {
	return filepath.Join(os.Getenv("HOME"), ".cache")
}
func tokenCacheFile(config *oauth2.Options) string {
	hash := fnv.New32a()
	hash.Write([]byte(config.ClientID))
	hash.Write([]byte(config.ClientSecret))
	hash.Write([]byte(config.Scopes[0]))
	fn := fmt.Sprintf("api-tok%v", hash.Sum32())
	return filepath.Join(osUserCacheDir(), url.QueryEscape(fn))
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	//if !*cacheToken {
	//return nil, errors.New("--cachetoken is false")
	//}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := new(oauth2.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}
