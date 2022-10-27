package rpc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"github.com/GabeCordo/fack"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	StandardTimeout   = time.Duration(1) * time.Second
	Decimal           = 10
	MissingNonceValue = 0
)

type Request struct {
	Function string   `json:"function"`
	Param    []string `json:"param"`
	Auth     struct {
		Signature []byte `json:"signature"`
		Nonce     int64  `json:"nonce"`
	} `json:"auth"`
}

func NewRequest(function string) *Request {
	request := new(Request)
	request.Function = function
	return request
}

// interface methods

func (r Request) Endpoint() string {
	return r.Function
}

func (r Request) Bytes() []byte {
	byteData, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return byteData
}

func (r Request) Hash() [32]byte {
	concatenatedString := r.Function + strconv.FormatInt(r.Auth.Nonce, Decimal)
	return sha256.Sum256([]byte(concatenatedString))
}

func (r Request) Nonce() int64 {
	return r.Auth.Nonce
}

func (r Request) Signature() []byte {
	return []byte(r.Auth.Signature)
}

func (r *Request) Sign(key *ecdsa.PrivateKey) error {
	// if the nonce has never been created, generate one
	if r.Auth.Nonce == MissingNonceValue {
		r.Auth.Nonce = fack.GenerateNonce() // int64 -> currentTime * random int
	} else {
		// the Node will verify that the nonce is greater than the previous, otherwise
		// we risk allowing a threat actor to re-send the same nonce and signature again
		r.Auth.Nonce++
	}

	hash := r.Hash()
	signature, err := ecdsa.SignASN1(rand.Reader, key, hash[:]) // [32]byte -> []byte
	if err != nil {
		return err
	}
	r.Auth.Signature = signature

	return nil
}

// rpc methods

func (r *Request) Send(method, url string) (*Response, error) {
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
