package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GabeCordo/fack"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type NodeStatus int

const (
	Startup NodeStatus = iota
	Running
	Frozen
	Killed
)

type Node struct {
	Name    string       `json:"name"`
	Address fack.Address `json:"address"`
	Debug   bool         `json:"debug"`
	Status  NodeStatus   `json:"status"`

	Auth *fack.Auth

	mux    *http.ServeMux
	server *http.Server
	mutex  sync.Mutex
}

// NewNode
// address : Address -> defines the listening host and port
// auth : *Auth ->
//
// LEGACY: func NewNode(address Address, debug bool, auth *Auth, logger *logger.Logger) *Node {
func NewNode(address fack.Address, optional ...interface{}) *Node {
	node := new(Node)

	for _, o := range optional {
		switch val := o.(type) {
		case *fack.Auth:
			node.Auth = val // default: nil
		case bool:
			node.Debug = val // default: false
		}
	}

	node.Name = fack.GenerateRandomString(int(fack.GenerateNonce()))
	node.mux = http.NewServeMux()
	node.server = new(http.Server)
	node.Address = address
	node.Status = Startup

	return node
}

func (node Node) IsAuthAttached() bool {
	return node.Auth != nil
}

func (node Node) MissingModules() bool {
	return !node.IsAuthAttached()
}

func (node *Node) SetStatus(status NodeStatus) {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	node.Status = status // if we don't lock this, two core attempting to change the status can cause a race condition
}

func (node *Node) SetName(name string) {
	// a node name should be static during runtime given that during a time interval from (0 to inf)
	// if N logs are stored, using the node name as an id, then a dynamic name change at time t
	// would render new logs created from (t to inf) detached from logs created from (0 to t)
	if node.Status == Startup {
		node.Name = name
	}
}

func (node *Node) Function(path string, handler fack.Router, methods []string, auth bool) error {

	// the user should not be able to create a route that requires ECDSA or permission bitmap
	// authentication if they have not registered an auth structure
	if (node.Auth == nil) && auth {
		return errors.New("cannot create a route that requires authentication with a nil auth struct")
	}

	// the developer should know that they're breaking the pattern by calling this during runtime
	if node.Status != Startup {
		return errors.New("Endpoints should not be added dynamically to the Node during runtime")
	}

	// by passing a hash-map to the handler function instead of a list of methods, we can perform a method
	// look-up in O(1) time versus O(node) needed to iterate over a list of methods
	methodsHashTable := fack.ArrayToLookupHashTable(methods)

	node.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		response := NewResponse()
		defer response.Send(w)

		HTTPRequestAllowedOnHandler := false
		if _, ok := methodsHashTable[r.Method]; ok {
			HTTPRequestAllowedOnHandler = true
		}

		if HTTPRequestAllowedOnHandler {
			if node.Debug {
				log.Printf(node.Name, "Request %s used an approved %s method.\n", r.Host, r.Method)
			}

			if !fack.IsUsingJSONContent(r) {
				response.AddStatus(http.StatusBadRequest, "Only JSON Content permitted")
				return
			}

			// we will see if the IP address has a mapped local or global permission to the endpoint
			sender, error := fack.GetInternetProtocol(r)
			if error != nil {
				response.AddStatus(http.StatusInternalServerError, "Internet Protocol Parser Failed")
				return
			}

			/** Unmarshal the JSON body to Request Struct */
			body := new(Request)

			httpBodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				response.AddStatus(http.StatusInternalServerError, err.Error())
				return
			}
			s := string(httpBodyBytes)
			err = json.Unmarshal([]byte(s), body)
			if err != nil {
				if node.Debug {
					log.Printf(node.Name, "Request %s contained a malformed HTTP body\n", r.Host)
				}
				response.AddStatus(http.StatusBadRequest, err.Error())
				return
			}

			if auth {
				// Why not pass the lambda provided by the request to IsEndpointAuthorized?
				//		-> the user is not forced to use the request.Send() method and can
				//		   direct the request to an url they do not have permission for while
				//		   inserting an url path as the lambda for a route they do have permission
				//		   for
				// Why not place method into request type as well?
				//		-> a lambda can support > 1 HTTP method
				//		-> it is safer to use a server-defined method that the node has control over
				if (node.Debug && sender.Host == fack.Localhost) || node.Auth.IsEndpointAuthorized(sender, body, path, r.Method) {
					// the request IP destination either had local or global permission
					handler(body, response)
				} else {
					// the request IP destination does not have local or global permission
					if node.Debug {
						log.Printf("Request %s attempted to submit a request to %s (%s); did not have permission\n", path, r.Method)
					}
					response.AddStatus(http.StatusUnauthorized, "Bye Bye.")
				}
			} else {
				// the endpoint does not require the destination ip of the request to have local or global
				// permission to send messages to the Node
				handler(body, response)
			}
		} else {
			if node.Debug {
				log.Printf(node.Name, "Request to %s failed; Path does not support %s\n", path, r.Method)
			}
			response.AddStatus(http.StatusForbidden, "HTTP Method Not Allowed")
		}

		// this exists in the event that an unintended error or unforeseen error has been improperly handled
		// by a user-defined handler function or a packet has corrupted IP / body data
		defer func() {
			if err := recover(); err != nil {
				response.AddStatus(http.StatusInternalServerError, "Node panic")
			}
		}()
	})

	return nil
}

func (node *Node) Start() {
	node.SetStatus(Running) // thread safe

	http.ListenAndServe(node.Address.ToString(), node.mux)
}

func (node *Node) Shutdown() {
	node.SetStatus(Killed) // TODO - kill HTTP server

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	node.server.Shutdown(ctx)
}

func (node Node) ToString() string {
	j, _ := json.Marshal(node)
	return string(j)
}
