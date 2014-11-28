package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	calendar "code.google.com/p/google-api-go-client/calendar/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	clientId         = flag.String("clientid", "", "OAuth Client ID.  If non-empty, overrides --clientid_file")
	clientIdFile     = flag.String("clientid_file", "clientid.ini", "Name of a file containing just the project's OAuth Client ID from https://code.google.com/apis/console/")
	clientSecret     = flag.String("secret", "", "OAuth Client Secret.  If non-empty, overrides --secret_file")
	clientSecretFile = flag.String("secret_file", "clientsecret.ini", "Name of a file containing just the project's OAuth Client Secret from https://code.google.com/apis/console/")
	clientOptions    *oauth2.Options
	clientToken      *oauth2.Token
	calendarId       = flag.String("calendar", "", "fetch from this calendarId")
)

type CacherRoundTripper struct {
	Transport *oauth2.Transport
}

func getEvents(count int64) []*calendar.Event {

	store := &tokenStore{}
	transport, err := clientOptions.NewTransportFromTokenStore(store)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	client := &http.Client{Transport: transport}

	svc, err := calendar.New(client)
	if err != nil {
		fmt.Println(err)
	}
	//c, err := svc.CalendarList.List().Do()
	startTime := time.Now().Format(time.RFC3339)
	c, err := svc.Events.List(*calendarId).SingleEvents(true).OrderBy("startTime").TimeMin(startTime).MaxResults(count).Do()
	//c, err := svc.Events.List(*calendarId).Do()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return c.Items

}

//TODO sort by time http://stackoverflow.com/questions/23121026/sorting-by-time-time-in-golang

func initOauth() {
	var err error
	clientOptions, err = oauth2.New(
		oauth2.Client(valueOrFileContents(*clientId, *clientIdFile), valueOrFileContents(*clientSecret, *clientSecretFile)),
		oauth2.RedirectURL("http://localhost:3000/oauthredirect"),
		oauth2.Scope(
			calendar.CalendarReadonlyScope,
		),
		google.Endpoint(),
	)
	clientOptions.TokenStore = &tokenStore{}
	if err != nil {
		log.Println("err")
	}
}

func handleSetupOauth(w http.ResponseWriter, r *http.Request) {
	if *clientId == "" || *clientSecret == "" {
		fmt.Println("--clientid and --secret must be provided!")
		return
	}

	url := clientOptions.AuthCodeURL("", "offline", "auto")
	http.Redirect(w, r, url, http.StatusFound)
}

func handleOauthRedirect(w http.ResponseWriter, r *http.Request) (int, string) {
	code := r.FormValue("code")
	if code == "" {
		return 500, "No code!"
	}

	t, err := clientOptions.NewTransportFromCode(code)
	if err != nil {
		log.Fatal(err)
	}
	cacheFile := tokenCacheFile(clientOptions)
	saveToken(cacheFile, t.Token())

	http.Redirect(w, r, "/", http.StatusFound)
	return 200, "<h1>Success</h1>Authorized."
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
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t := new(oauth2.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	clientToken = token
	log.Printf("Saving token to %q token:  %+v ", file, token)
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}

type tokenStore struct {
}

func (t *tokenStore) WriteToken(token *oauth2.Token) {
	cacheFile := tokenCacheFile(clientOptions)
	saveToken(cacheFile, token)
}

func (t *tokenStore) ReadToken() (*oauth2.Token, error) {
	////Fetch token from file if we need to
	if clientToken == nil {
		cacheFile := tokenCacheFile(clientOptions)
		var err error
		clientToken, err = tokenFromFile(cacheFile)
		if err != nil {
			fmt.Println("Cannot find file: %q", cacheFile)
			return nil, fmt.Errorf("Cannot find file: %q", cacheFile)
		}
		log.Printf("Using cached token %+v from %q", clientToken, cacheFile)
		return clientToken, nil
	}
	return clientToken, nil
}
