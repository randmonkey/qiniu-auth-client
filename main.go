package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	var targetURL string
	var body string
	var contentType string
	var accessKey string
	var secretKey string
	var method string
	var verbose bool
	flag.StringVar(&method, "X", "GET", "http request method")
	flag.StringVar(&targetURL, "u", "", "URL")
	flag.StringVar(&body, "d", "", "request body")
	flag.StringVar(&contentType, "t", "application/json", "content type")
	flag.StringVar(&accessKey, "ak", "", "QINIU access key")
	flag.StringVar(&secretKey, "sk", "", "QINIU secret key")
	flag.BoolVar(&verbose, "v", false, "show verbose (HTTP headers)")
	flag.Parse()

	client := &http.Client{}
	client.Transport = &qiniuMacTransport{
		AccessKey:    accessKey,
		SecretKey:    []byte(secretKey),
		RoundTripper: http.DefaultTransport,
		showHeaders:  verbose,
	}

	reqURL, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("invalid URL %s, parse URL failed: %v", targetURL, err)
	}
	if reqURL.Scheme == "" {
		reqURL.Scheme = "http"
	}
	req, err := http.NewRequest(method, reqURL.String(), bytes.NewReader([]byte(body)))
	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}
	if verbose {
		fmt.Printf("> %s %s\n", req.Method, req.URL.String())
	}
	if body != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println()
		log.Fatalf("failed to get response: %v", err)
	}
	if verbose {
		fmt.Println()
		fmt.Printf("< %s %s %s\n", resp.Request.Method, resp.Request.URL.String(), resp.Status)
		for key, value := range resp.Header {
			fmt.Printf("< %s=%s\n", key, strings.Join(value, ";"))
		}
	}
	respBobyBuf, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(respBobyBuf))

}
