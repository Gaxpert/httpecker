package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	app       = kingpin.New("httpecker", "Checks response for http/https urls")
	use_http  = app.Flag("http-only", "Use http only").Default("false").Bool()
	use_https = app.Flag("https-only", "Use https only").Default("false").Bool()
	use_both  = app.Flag("both", "Use http and https ").Default("false").Bool()
	filename  = app.Flag("filename", "File with urls").Short('f').Default("").String()
	threads   = app.Flag("threads", "Number of threads").Short('t').Default("5").Int()
	verbosity = app.Flag("verb", "Enable debugging").Short('v').Default("false").Bool()
)

//For string replacements since we will be using them alot
var http_to_https = strings.NewReplacer("http://", "https://")
var https_to_http = strings.NewReplacer("https://", "http://")

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if *verbosity {
		log.SetLevel(log.TraceLevel)
	}
	//If no flags are set we use http and https
	if *use_both {
		*use_http = true
		*use_https = true
	}
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		Dial:                (&net.Dialer{Timeout: 0, KeepAlive: 0}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}}
	log.Debug("http: ", *use_http)
	log.Debug("https: ", *use_https)

	//Accept file from stdin or from a file
	var s *bufio.Scanner
	if *filename == "" {
		s = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(*filename)
		if err != nil {
			log.Fatal("Failed opening file: %s", err)
		}
		s = bufio.NewScanner(file)
	}

	chan_urls := make(chan string)
	go func() {
		for s.Scan() {
			// run(s.Text(), client, *use_http, *use_https)
			chan_urls <- s.Text()
		}
		close(chan_urls)
	}()

	wg := &sync.WaitGroup{}
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go run(chan_urls, client, *use_http, *use_https, wg)
	}
	wg.Wait()

}

func run(chan_urls chan string, client *http.Client, http bool, https bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range chan_urls {

		log.Debug("Original url: ", url)

		url_original := url
		var status_http int
		var status_https int

		//Check http prefix. If not warn and add prefix
		if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
			log.Warn("Missing \"http://\" ...  adding")
			url = "http://" + url
		}

		//Run according to args
		if http && https { // Run for http and https with apropiate conversions in url

			//Modify protocol to http if needed and query
			if !strings.HasPrefix("http://", url) {
				url = https_to_http.Replace(url)
			}
			log.Debug("Query url http: ", url)

			status_http = check_status(url, client)

			//Modify protocol to https if needed and query
			if !strings.HasPrefix("https://", url) {
				url = http_to_https.Replace(url)
			}

			log.Debug("Query url https: ", url)

			status_https = check_status(url, client)
			fmt.Printf("%d | %d | %s \n", status_http, status_https, url_original)

		} else if http { //Run for --http-only

			//Modify protocol to http if needed and query
			if !strings.HasPrefix("https://", url) {
				url = https_to_http.Replace(url)
			}
			log.Debug("Query url http: ", url)

			status_http = check_status(url, client)
			fmt.Printf("%d | %s \n", status_http, url_original)

		} else if https { //Run for --https-only

			//Modify protocol to https if needed and query
			if !strings.HasPrefix("http://", url) {
				url = http_to_https.Replace(url)
			}
			log.Debug("Query url https: ", url)

			status_https = check_status(url, client)
			fmt.Printf("%d | %s \n", status_https, url_original)

		} else { //Default case, don't convert urls

			log.Debug("Query url default: ", url)
			status_http = check_status(url, client)
			fmt.Printf("%d | %s \n", status_http, url_original)
		}
	}
}

func check_status(url string, client *http.Client) int {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Connection", "close")
	if err != nil {
		return 999
	}
	resp, err := client.Do(req)
	if err != nil {
		return 999
	}
	defer resp.Body.Close()

	return resp.StatusCode
}
