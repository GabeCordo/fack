package rpc

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	StandardTimeout = time.Duration(1) * time.Second
	Decimal         = 10
)

type Request struct {
	Function string   `json:"function"`
	Param    []string `json:"param,omitempty"`
	Auth     struct {
		Signature []byte `json:"signature,omitempty"`
		Nonce     int64  `json:"nonce,omitempty"`
	} `json:"auth,omitempty"`
}

func NewRequest(function string) *Request {
	request := new(Request)
	request.Function = function
	return request
}

// interface methods

func (r Request) GetEndpoint() string {
	return r.Function
}

func (r Request) Bytes() []byte {
	byteData, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return byteData
}

func (r Request) GetHash() []byte {
	concatenatedString := r.Function + strconv.FormatInt(r.Auth.Nonce, Decimal)
	bit32ShaBytes := sha256.Sum256([]byte(concatenatedString))

	return bit32ShaBytes[:]
}

func (r Request) GetNonce() int64 {
	return r.Auth.Nonce
}

func (r *Request) SetNonce(nonce int64) {
	r.Auth.Nonce = nonce
}

func (r Request) GetSignature() []byte {
	return r.Auth.Signature
}

func (r *Request) SetSignature(bytes []byte) {
	r.Auth.Signature = bytes
}

// rpc methods

func (r Request) Send(method, url string) (*Response, error) {
	httpClient := http.Client{Timeout: StandardTimeout}

	httpUrl := url + r.Function
	log.Println(httpUrl)
	httpRequest, err := http.NewRequest(method, httpUrl, nil)
	if err != nil {
		return nil, err
	}

	// the server will only accept core provided with the application/json Content-Type
	// header, otherwise the request will be rejected
	httpRequest.Header.Add("Content-Type", "application/json")
	httpRequest.Body = io.NopCloser(bytes.NewReader(r.Bytes()))

	// note -> any auth should be done before this function call
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// the HTTP data will contain and EOF header at the start of the response
	// body that needs to be stripped before we can unmarshal the JSON content
	strBody := string(body)

	// Unmarshal or Decode the JSON to the frontend.
	result := new(Response)
	json.Unmarshal([]byte(strBody), &result)

	return result, nil
}
