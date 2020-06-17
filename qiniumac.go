package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
)

const qiniuHeaderPrefix = "X-Qiniu-"

func includeBody(req *http.Request, ctType string) bool {
	return req.ContentLength != 0 && req.Body != nil && ctType != "" && ctType != "application/octet-stream"
}

// ---------------------------------------------------------------------------------------

type sortByHeaderKey []string

func (p sortByHeaderKey) Len() int           { return len(p) }
func (p sortByHeaderKey) Less(i, j int) bool { return p[i] < p[j] }
func (p sortByHeaderKey) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func signQiniuHeaderValues(header http.Header, w io.Writer) {
	var keys []string
	for key := range header {
		if len(key) > len(qiniuHeaderPrefix) && key[:len(qiniuHeaderPrefix)] == qiniuHeaderPrefix {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return
	}

	if len(keys) > 1 {
		sort.Sort(sortByHeaderKey(keys))
	}
	for _, key := range keys {
		io.WriteString(w, "\n"+key+": "+header.Get(key))
	}
}

// SignRequest calculate
func SignRequest(sk []byte, req *http.Request) ([]byte, error) {

	h := hmac.New(sha1.New, sk)

	u := req.URL
	data := req.Method + " " + u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data+"\nHost: "+req.Host)

	ctType := req.Header.Get("Content-Type")
	if ctType != "" {
		io.WriteString(h, "\nContent-Type: "+ctType)
	}

	signQiniuHeaderValues(req.Header, h)

	io.WriteString(h, "\n\n")

	if includeBody(req, ctType) {
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		req.Body.Close() //  must close
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		h.Write(bodyBytes)
	}

	return h.Sum(nil), nil
}
