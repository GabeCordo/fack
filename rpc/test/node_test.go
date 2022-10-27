package test

import (
	"fmt"
	"github.com/GabeCordo/fack"
	"github.com/GabeCordo/fack/rpc"
	"net/http"
	"testing"
	"time"
)

const (
	GETPort            = 8000
	SuccessMessage     = "success"
	LocalHost          = "http://127.0.0.1:"
	WaitForServerStart = 1000 * time.Millisecond
)

/*? Routing Functions */

func index(request fack.Request, response fack.Response) {
	response.Status(http.StatusOK).Description(SuccessMessage)
}

/*? Source Code to Test */

func StartupHTTPNodeWithGETEnabled() {
	a := fack.Address{"localhost", GETPort}
	n := rpc.NewNode(a, false) // pass a nil to logger pointer ~ no logging
	n.Function("/", index, []string{"GET"}, false)
	go n.Start()
}

/*? Test Function */

func TestNodeMissingAuthStruct(t *testing.T) {
	a := fack.Address{"", 8000}
	n := rpc.NewNode(a, false) // pass a nil to logger pointer ~ no logging

	// since we have not registered an auth struct to the node, we should not be allowed to create an auth-mandatory route
	err := n.Function("/", index, []string{"GET"}, true)
	if err == nil {
		t.Error("a node should not be able to create an authenticated routed path without an auth node")
	}
}

func TestAttemptAddRouteOutsideOfStartup(t *testing.T) {
	a := fack.Address{"localhost", 8000}
	n := rpc.NewNode(a, false)
	n.SetStatus(fack.Running) // simulate the n.Start() function

	err := n.Function("/", index, []string{"GET"}, false)
	if err == nil {
		t.Error("a node should not be able to dynamically assign routes while running")
	}
}

func TestNodeReceivedNonJSONRequest(t *testing.T) {
	StartupHTTPNodeWithGETEnabled()

	time.Sleep(WaitForServerStart)

	rsp, err := http.Get("http://127.0.0.1:8000/")
	if err != nil {
		t.Error("could not connect to node properly")
	}

	if rsp.StatusCode != http.StatusBadRequest {
		t.Error("node is not properly rejecting non-json core")
	}
}

func TestNodeRequestToAllowedMethod(t *testing.T) {
	StartupHTTPNodeWithGETEnabled()

	// if you are on macos, you may need to give the binary permission to use a socket port
	time.Sleep(WaitForServerStart)

	request := rpc.NewRequest("/")

	/* GET request should succeed */
	url := LocalHost + fmt.Sprint(GETPort)
	resp, err := request.Send("GET", url)
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	//if (resp.Status != http.StatusOK) || ((*resp).Data["status"] != SuccessMessage) {
	if resp.GetStatus() != http.StatusOK {
		t.Error("Did not receive the correct HTTP JSON Response")
	}
}

func TestNodeRequestToBlockedMethod(t *testing.T) {
	StartupHTTPNodeWithGETEnabled()

	// if you are on macos, you may need to give the binary permission to use a socket port
	time.Sleep(WaitForServerStart)

	request := rpc.NewRequest("/")

	/* POST request should not be supported */
	resp, err := request.Send("POST", LocalHost+fmt.Sprint(GETPort))

	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	if resp.GetStatus() != http.StatusForbidden {
		t.Error("The node is accepting unwanted HTTP method types")
	}
}
