package fack

import "crypto/ecdsa"

type Router func(request Request, response Response)

type Node interface {
	Start()
	Shutdown()
	AddFunction(path string, handler Router) Route
}

type Request interface {
	Endpoint() string
	Signature() []byte
	Hash() [32]byte
	Nonce() int64
	Sign(key *ecdsa.PrivateKey) error
}

type ResponseData map[string]interface{}

type Response interface {
	GetStatus() int
	SetStatus(int) Response
	GetDescription() string
	SetDescription(string) Response
	GetData() ResponseData
	Pair(string, any) Response
}
