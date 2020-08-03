/*
BAD2BEEF's POST Handler
https://github.com/bad2beef/Toolbox
Receive HTTP POSTs (or whatevers), write to disk. No more PHP + FastCGI + Nginx for this simple task!
*/

package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type POSTLog struct {
	URI           string      `json:"uri"`
	RemoteAddress string      `json:"remote_address"`
	Headers       http.Header `json:"headers"`
	Body          string      `json:"body"`
}

func httpRedirectFromForm(httpWriter http.ResponseWriter, httpRequest *http.Request) {
	for key, value := range httpRequest.Form {
		switch strings.ToLower(key) {
		case "redirect":
		case "redir":
		case "uri":
		case "url":
		default:
			continue
		}

		log.Printf("%s REDIRECT %s", httpRequest.RemoteAddr, value[0])
		http.Redirect(httpWriter, httpRequest, value[0], http.StatusFound)
	}
}

func httpHandlerPOST(httpWriter http.ResponseWriter, httpRequest *http.Request) {
	body, err := ioutil.ReadAll(httpRequest.Body)
	if err != nil {
		log.Printf("Unable to read body. %s", err)
		return
	}

	err = httpRequest.ParseForm()
	if err != nil {
		log.Printf("%s Could not parse form data. %s", httpRequest.RemoteAddr, err)
	} else {
		defer httpRedirectFromForm(httpWriter, httpRequest)
	}

	hash := sha256.New()
	_, _ = hash.Write([]byte(httpRequest.RequestURI))
	_, _ = hash.Write(body)
	hashSum := hash.Sum(nil)

	filePath := path.Join("data", fmt.Sprintf("%0x", hashSum[0]), fmt.Sprintf("%0x", hashSum[1]))
	fileName := path.Join(filePath, fmt.Sprintf("%0x", hashSum))

	if _, err := os.Stat(fileName); err == nil { // Don't re-write same payload
		return
	}

	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		log.Printf("Unable to create data directory. %s", err)
		return
	}

	var POST POSTLog
	POST.URI = httpRequest.RequestURI
	POST.RemoteAddress = httpRequest.RemoteAddr
	POST.Headers = httpRequest.Header
	POST.Body = string(body)

	json, err := json.MarshalIndent(POST, "", "  ")
	if err != nil {
		log.Printf("Unable to marshal JSON. %s", err)
		return
	}

	err = ioutil.WriteFile(fileName, json, 0600)
	if err != nil {
		log.Printf("Unable to write file. %s", err)
		return
	}

	log.Printf("%s %d %0x", httpRequest.RemoteAddr, len(body), hashSum)
}

func main() {
	listener := flag.String("listen", ":8080", "HTTP listener address")
	route := flag.String("route", "/", "URL Route")
	certificate := flag.String("cert", "", "Certificate chain")
	key := flag.String("key", "", "Private key for certificate")
	flag.Parse()

	log.Printf("Listening on %s%s", *listener, *route)
	http.HandleFunc(*route, httpHandlerPOST)

	if len(*certificate) > 0 && len(*key) > 0 {
		if err := http.ListenAndServeTLS(*listener, *certificate, *key, nil); err != nil {
			log.Fatalln(err)
		}
	} else {
		if err := http.ListenAndServe(*listener, nil); err != nil {
			log.Fatalln(err)
		}
	}
}
