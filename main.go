package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type httpHeader struct {
	key   string
	value string
}

type httpHeaderSlice []httpHeader

func (s *httpHeaderSlice) String() string {
	ret := ""
	for i, kv := range *s {
		ret = ret + kv.key + ":" + kv.value
		if i < len(*s)-1 {
			ret = ret + ","
		}
	}
	return ret
}

func (s *httpHeaderSlice) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid header should be in `key:value` format")
	}
	if len(parts[0]) == 0 {
		return fmt.Errorf("invalid header, key is empty")
	}
	*s = append(*s, httpHeader{
		key:   parts[0],
		value: parts[1],
	})
	return nil
}

func main() {
	var targetURL string
	var body string
	var bodyFile string
	var contentType string
	var accessKey string
	var secretKey string
	var method string
	var verbose bool
	var headers httpHeaderSlice
	flag.StringVar(&method, "X", "GET", "http request method")
	flag.Var(&headers, "H", "HTTP headers")
	flag.StringVar(&targetURL, "u", "", "URL")
	flag.StringVar(&body, "d", "", "request body, cannot be used with -f")
	flag.StringVar(&bodyFile, "f", "", "filename to load body from, cannot be used with -d")
	flag.StringVar(&contentType, "t", "application/json", "content type")
	flag.StringVar(&accessKey, "ak", "", "QINIU access key")
	flag.StringVar(&secretKey, "sk", "", "QINIU secret key")
	flag.BoolVar(&verbose, "v", false, "show verbose (HTTP headers)")
	flag.Parse()

	if body != "" && bodyFile != "" {
		log.Fatalf("cannot specify both body content and body file")
	}

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
	var req *http.Request
	if body != "" || bodyFile == "" {
		req, err = http.NewRequest(method, reqURL.String(), bytes.NewReader([]byte(body)))
	} else {
		file, errOpen := os.Open(bodyFile)
		if errOpen != nil {
			log.Fatalf("failed to open file %s, error %v", bodyFile, err)
		}
		req, err = http.NewRequest(method, reqURL.String(), file)
	}

	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}
	if verbose {
		fmt.Printf("> %s %s\n", req.Method, req.URL.String())
	}
	if body != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if len(headers) > 0 {
		for _, h := range headers {
			req.Header.Set(h.key, h.value)
		}
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
