package rpc

import (
	"context"
	"encoding/json"
	"fmt"
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
	name string `json:"name"`

	address *fack.Address `json:"address"`
	status  NodeStatus    `json:"status"`

	auth *fack.Auth

	mux    *http.ServeMux
	server *http.Server
	mutex  sync.Mutex
}

// NewNode
// address : address -> defines the listening host and port
// auth : *auth ->
//
// LEGACY: func NewNode(address address, debug bool, auth *auth, logger *logger.Logger) *Node {
func NewNode(address *fack.Address, optional ...any) *Node {
	node := new(Node)

	for _, o := range optional {
		switch val := o.(type) {
		case string:
			node.name = val
		case *fack.Auth:
			node.auth = val // default: nil
		}
	}

	// if a name is never passed to the node, generate a random string of chars
	if len(node.name) == 0 {
		node.name = fack.GenerateRandomString(int(fack.GenerateNonce()))
	}

	// if no auth node is passed in, generate an empty one
	if node.auth == nil {
		node.auth = fack.NewAuth()
	}

	node.mux = http.NewServeMux()
	node.server = new(http.Server)

	node.address = address
	node.status = Startup

	return node
}

func (node *Node) Status(status NodeStatus) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	// if we don't lock this, two core attempting to change the status can cause a race condition
	node.status = status
}

func (node *Node) Name(name string) {
	// a node name should be static during runtime given that during a time interval from (0 to inf)
	// if N logs are stored, using the node name as an id, then a dynamic name change at time t
	// would render new logs created from (t to inf) detached from logs created from (0 to t)
	if node.status == Startup {
		node.name = name
	}
}

func (node *Node) Auth(auth *fack.Auth) {
	// there is no reason to switch auth nodes at the moment during runtime
	if node.status == Startup {
		node.auth = auth
	}
}

func (node *Node) Function(path string, handler fack.Router) *fack.Route {

	// functions should be assigned before the node is running
	if node.status != Startup {
		panic("Endpoints should not be added dynamically to the Node during runtime")
	}

	route := fack.EmptyRoute()

	//node.mux.HandleFunc(path, TestHandler)
	node.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		response := NewResponse()
		defer response.Send(w)

		fmt.Printf(r.Method)

		method := fack.HTTPMethodFromString(r.Method)
		if !route.IsMethodSupported(method) {
			log.Printf("[%s] Request to %s failed; Path does not support %s\n", node.name, path, r.Method)
			response.AddStatus(http.StatusForbidden, "HTTP Method Not Allowed")
			return
		} else {
			log.Printf("[%s] Request %s used an approved %s method.\n", node.name, r.Host, r.Method)
		}

		if !fack.IsUsingJSONContent(r) {
			log.Println("not using JSON content")
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
		var body *Request = &Request{}

		httpBodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			response.AddStatus(http.StatusInternalServerError, err.Error())
			return
		}
		s := string(httpBodyBytes)
		err = json.Unmarshal([]byte(s), body)
		if err != nil {
			log.Println(err.Error())
			if route.Debug {
				log.Printf("[%s] Request %s contained a malformed HTTP body\n", node.name, r.Host)
			}
			response.AddStatus(http.StatusBadRequest, err.Error())
			return
		}

		if route.RequiresAuth {
			// Why not pass the lambda provided by the request to IsEndpointAuthorized?
			//		-> the user is not forced to use the request.Send() method and can
			//		   direct the request to an url they do not have permission for while
			//		   inserting an url path as the lambda for a route they do have permission
			//		   for
			// Why not place method into request type as well?
			//		-> a lambda can support > 1 HTTP method
			//		-> it is safer to use a server-defined method that the node has control over
			if (route.Debug && sender.IsLocalHost()) || node.auth.IsEndpointAuthorized(sender, body, path, method) {
				// the request IP destination either had local or global permission
				handler(body, response)
			} else {
				// the request IP destination does not have local or global permission
				if route.Debug {
					log.Printf("Request %s attempted to submit a request to %s (%s); did not have permission\n", path, r.Method)
				}
				response.AddStatus(http.StatusUnauthorized, "Bye Bye.")
			}
		} else {
			// the endpoint does not require the destination ip of the request to have local or global
			// permission to send messages to the Node
			handler(body, response)
		}

		// this exists in the event that an unintended error or unforeseen error has been improperly handled
		// by a user-defined handler function or a packet has corrupted IP / body data
		defer func() {
			if err := recover(); err != nil {
				response.AddStatus(http.StatusInternalServerError, "Node panic")
			}
		}()
	})

	return route
}

func (node *Node) Start() {
	node.Status(Running) // thread safe

	log.Printf("(!) http node started on %s\n", node.address.ToString())

	http.ListenAndServe(node.address.ToString(), node.mux)
}

func (node *Node) Shutdown() {
	node.Status(Killed) // TODO - kill HTTP server

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
