package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

type qiniuMacTransport struct {
	AccessKey    string
	SecretKey    []byte
	RoundTripper http.RoundTripper
	showHeaders  bool
}

func (tr *qiniuMacTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if tr.AccessKey != "" {
		sign, err := SignRequest(tr.SecretKey, req)
		if err != nil {
			return nil, err
		}

		auth := "Qiniu " + tr.AccessKey + ":" + base64.URLEncoding.EncodeToString(sign)
		req.Header.Set("Authorization", auth)
	}
	if tr.showHeaders {
		for key, value := range req.Header {
			fmt.Printf("> %s=%s\n", key, strings.Join(value, ";"))
		}
	}

	return tr.RoundTripper.RoundTrip(req)
}
