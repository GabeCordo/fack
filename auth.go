package fack

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"sync"
)

const (
	MissingNonceValue int64 = 0
)

type Auth struct {
	Trusted map[string]*Endpoint `json:"trusted"`
	Mutex   sync.Mutex
}

func NewAuth() *Auth {
	// mutex is initialized implicitly by the struct
	auth := new(Auth)
	auth.Trusted = make(map[string]*Endpoint)
	return auth
}

func (na *Auth) AddTrusted(ip string, ne *Endpoint) bool {
	if ne == nil {
		return false
	}
	na.Mutex.Lock()
	defer na.Mutex.Unlock()

	if _, ok := na.Trusted[ip]; !ok {
		na.Trusted[ip] = ne
		return true
	} else {
		return false
	}
}

func (na *Auth) RemoveTrusted(ip string) error {
	na.Mutex.Lock()
	delete(na.Trusted, ip)
	na.Mutex.Unlock()
	return nil
}

func (na *Auth) IsEndpointAuthorized(sender *Address, request Request, path string, method HTTPMethod) bool {
	validFlag := false // by default, we will assume that the ip doesn't exist in the hash map
	if endpoint, ok := na.Trusted[sender.GetHost()]; ok {
		// 1. does the user have permission to send an HTTP method request to the current path
		// 2. does the message come from a user with the same ECDSA key pair
		validFlag = endpoint.HasPermissionToUseMethod(path, method) && endpoint.ValidateSource(request)
	}
	return validFlag
}

func Sign(request Request, key *ecdsa.PrivateKey) error {
	// if the nonce has never been created, generate one
	var nonce int64
	if request.GetNonce() == MissingNonceValue {
		nonce = GenerateNonce() // int64 -> currentTime * random int
	} else {
		// the Node will verify that the nonce is greater than the previous, otherwise
		// we risk allowing a threat actor to re-send the same nonce and signature again
		nonce = request.GetNonce() + 1
	}
	request.SetNonce(nonce)

	hash := request.GetHash()
	signature, err := ecdsa.SignASN1(rand.Reader, key, hash)
	if err != nil {
		errors.New("there was an error signing the request data")
	}
	request.SetSignature(signature)

	return nil
}
