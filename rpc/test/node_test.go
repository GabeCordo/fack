package test

import (
	"fmt"
	"github.com/GabeCordo/fack"
	"github.com/GabeCordo/fack/rpc"
	"log"
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
	response.SetStatus(http.StatusOK).SetDescription(SuccessMessage)
}

/*? Source Code to Test */

func StartupHTTPNodeWithGETEnabled() {
	a := fack.EmptyAddress().Host("localhost").Port(GETPort)

	n := rpc.NewNode(a)
	n.Function("/", index).Method(fack.GET)

	go n.Start()
}

/*? Test Function */

func TestAttemptAddRouteOutsideOfStartup(t *testing.T) {
	a := fack.LocalHost().Port(8000)

	n := rpc.NewNode(a, false)
	n.Status(fack.Running) // simulate the n.Start() function

	n.Function("/", index).Method(fack.GET)

	defer func() {
		if err := recover(); err != nil {
			log.Println("SUCCESS panicked on bad function assignment", err)
		}
	}()
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

func TestNode(t *testing.T) {
	address := fack.EmptyAddress().Host("localhost").Port(8080)

	node := rpc.NewNode(address, "test node", true)

	node.Function("/", func(request fack.Request, response fack.Response) {
		log.Println("received request")
		response.SetStatus(200).SetDescription("test\n")
	}).Method(fack.GET)

	go node.Start()

	time.Sleep(5 * time.Second)
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

	//if (resp.status != http.StatusOK) || ((*resp).Data["status"] != SuccessMessage) {
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
