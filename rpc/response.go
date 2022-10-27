package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/GabeCordo/fack"
	"net/http"
)

const (
	Success        = "success"
	BadArgument    = "the syntax was correct but the types provided did not match"
	SyntaxMismatch = "the required number of parameters was not satisfied"
	Failure        = "there is an unspecified internal error"
	ByeBye         = "bad authentication"
)

type Response struct {
	status      int               `json:"status"`
	description string            `json:"description"`
	data        fack.ResponseData `json:"data"`
}

func NewResponse() *Response {
	response := new(Response)
	response.status = http.StatusNoContent // if the status is never populated, this will be returned by the node
	response.data = make(fack.ResponseData)
	return response
}

// interface methods

func (r Response) GetStatus() int {
	return r.status
}

func (r Response) Status(status int) fack.Response {
	if status < 0 {
		panicMessage := fmt.Sprintf("status cannot be negative %d", status)
		panic(panicMessage)
	}

	dataCopy := make(fack.ResponseData)
	for key, value := range r.data {
		dataCopy[key] = value
	}
	return Response{status, r.description, dataCopy}
}

func (r Response) GetDescription() string {
	return r.description
}

func (r Response) Description(description string) fack.Response {
	if len(description) == 0 {
		panic("description cannot be an empty string")
	}

	dataCopy := make(fack.ResponseData)
	for key, value := range r.data {
		dataCopy[key] = value
	}
	return Response{r.status, description, dataCopy}
}

func (r Response) GetData() fack.ResponseData {
	return r.data
}

func (r Response) Pair(key string, value any) fack.Response {
	if _, found := r.data[key]; found {
		panic("the key already exists")
	}

	dataCopy := make(fack.ResponseData)
	for key, value := range r.data {
		dataCopy[key] = value
	}
	dataCopy[key] = value

	return Response{r.status, r.description, dataCopy}
}

func (r *Response) AddStatus(httpResponseCode int, message ...string) {
	r.status = httpResponseCode
	if len(message) > 0 {
		r.description = message[0]
	}
}

func (r *Response) Send(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	json.NewEncoder(w).Encode(r)
}
