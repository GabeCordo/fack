package fack

type Router func(request Request, response Response)

type Node interface {
	Start()
	Shutdown()
	AddFunction(path string, handler Router) Route
}

type Request interface {
	GetEndpoint() string
	GetSignature() []byte
	SetSignature(bytes []byte)
	GetHash() []byte
	GetNonce() int64
	SetNonce(nonce int64)
}

type ResponseData map[string]any

type Response interface {
	GetStatus() int
	SetStatus(int) Response
	GetDescription() string
	SetDescription(string) Response
	GetData() ResponseData
	Pair(string, any) Response
}
