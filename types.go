package fack

import "crypto/ecdsa"

type Router func(request Request, response Response)

type Node interface {
	Start()
	Shutdown()
	AddFunction(path string, handler Router, methods []string, auth bool)
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
	Status(int) Response
	GetDescription() string
	Description(string) Response
	GetData() ResponseData
	Pair(string, any) Response
}
