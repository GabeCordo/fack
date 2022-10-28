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
	Status      int               `json:"status"`
	Description string            `json:"description,omitempty"`
	Data        fack.ResponseData `json:"data,omitempty"`
}

func NewResponse() *Response {
	response := new(Response)
	response.Status = http.StatusNoContent // if the status is never populated, this will be returned by the node
	response.Data = make(fack.ResponseData)
	return response
}

// interface methods

func (r Response) GetStatus() int {
	return r.Status
}

func (r *Response) SetStatus(status int) fack.Response {
	if status < 0 {
		panicMessage := fmt.Sprintf("status cannot be negative %d", status)
		panic(panicMessage)
	}
	r.Status = status
	return r
}

func (r Response) GetDescription() string {
	return r.Description
}

func (r *Response) SetDescription(description string) fack.Response {
	if len(description) == 0 {
		panic("description cannot be an empty string")
	}

	r.Description = description
	return r
}

func (r Response) GetData() fack.ResponseData {
	return r.Data
}

func (r *Response) Pair(key string, value any) fack.Response {
	if _, found := r.Data[key]; found {
		panic("the key already exists")
	}
	r.Data[key] = value

	return r
}

func (r *Response) AddStatus(httpResponseCode int, message ...string) {
	r.Status = httpResponseCode
	if len(message) > 0 {
		r.Description = message[0]
	}
}

func (r *Response) Send(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)
	err := json.NewEncoder(w).Encode(r)
	if err != nil {
		fmt.Println(err.Error())
	}
}
