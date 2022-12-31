package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
	WaitForServerStart = 2 * time.Second
)

/*? Test Function */

func AuthenticatedIndex(request fack.Request, response fack.Response) {
	response.SetStatus(http.StatusOK).SetDescription("authenticated")
}

func TestECDSABytesToString(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("could not generate an ECDSA key")
	}

	b := elliptic.Marshal(key.Curve, key.X, key.Y)
	fmt.Println("original:")
	fmt.Println(b)

	s := fack.ByteToString(b)
	fmt.Println("to string:")
	fmt.Println(s)

	bGenerated, ok := fack.StringToByte(s)
	if !ok {
		t.Error("failed to convert from string to bytes")
	}
	fmt.Println("to bytes:")
	fmt.Println(bGenerated)

	res := bytes.Compare(b, bGenerated)
	if res != 0 {
		fmt.Println("the original and generated byte arrays do NOT match")
	}
}

func TestECDSAKeyGeneration(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("could not generate an ECDSA key")
	}

	endpoint := fack.Endpoint{}
	endpoint.PublicKey = &privateKey.PublicKey

	b := endpoint.PublicKeyToBytes()
	sb := string(b)

	b = []byte(sb)
	endpoint.GeneratePublicKey(b)
}

// since there is no global or local Permission bitmap assigned to the endpoint
// the auth function should not grant access to the route
func TestAuthNoGlobalOrLocalPermissionsPresent(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("Could not generate an ECDSA key pair")
	}

	var na *fack.Auth = fack.NewAuth()
	var ne *fack.Endpoint = fack.NewEndpoint("test", &privateKey.PublicKey)
	na.AddTrusted("127.0.0.1", ne)

	a := fack.LocalHost().SetPort(8000)
	n := rpc.NewNode(a, na) // pass a nil to logger pointer ~ no logging
	route := n.Function("/", AuthenticatedIndex).Method(fack.GET)
	route.RequiresAuth = true

	go n.Start()

	// if you are on macos, you may need to give the binary permission to use a socket port
	time.Sleep(WaitForServerStart)

	request := rpc.NewRequest("/")
	err = fack.Sign(request, privateKey)
	if err != nil {
		t.Error(err)
	}

	/* GET request should succeed */
	resp, err := request.Send("GET", LocalHost+fmt.Sprint(GETPort))
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	if resp.GetStatus() != http.StatusUnauthorized {
		t.Error("node not rejecting unauthorized endpoints properly")
	}
}

// the auth function should default to the GlobalPermission bitmap and grant
// the request access to the route
func TestAuthGlobalPermissionPresent(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("Could not generate an ECDSA key pair")
	}

	var na *fack.Auth = fack.NewAuth()
	var ne *fack.Endpoint = fack.NewEndpoint("test", &privateKey.PublicKey)

	globalPermissionMap := fack.Permission{true, false, false, false}
	ne.AddGlobalPermission(&globalPermissionMap)
	na.AddTrusted("127.0.0.1", ne)

	a := fack.LocalHost().SetPort(8000)
	n := rpc.NewNode(a, na) // pass a nil to logger pointer ~ no logging
	route := n.Function("/", AuthenticatedIndex).Method(fack.GET)
	route.RequiresAuth = true

	go n.Start()

	// if you are on macos, you may need to give the binary permission to use a socket port
	time.Sleep(WaitForServerStart)

	request := rpc.NewRequest("/")
	err = fack.Sign(request, privateKey)
	if err != nil {
		t.Error(err)
	}

	/* GET request should succeed */
	resp, err := request.Send("GET", LocalHost+fmt.Sprint(GETPort))
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	if (*resp).GetDescription() != "authenticated" {
		t.Error("Node could not authenticate a valid host")
	}
}

// despite an endpoint holding no global Permission bitmap, the auth function
// should default to the local Permission bitmap to determine authorization
// to the HTTP method and route
func TestAuthLocalPermissionPresent(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("Could not generate an ECDSA key pair")
	}

	var na *fack.Auth = fack.NewAuth()
	var ne *fack.Endpoint = fack.NewEndpoint("test", &privateKey.PublicKey)

	localPermissionMap := fack.Permission{true, false, false, false}
	ne.AddLocalPermission("/", &localPermissionMap)
	na.AddTrusted("127.0.0.1", ne)

	a := fack.LocalHost().SetPort(8000)
	n := rpc.NewNode(a, na) // pass a nil to logger pointer ~ no logging
	n.Function("/", AuthenticatedIndex).Method(fack.GET)

	go n.Start()

	// if you are on macos, you may need to give the binary permission to use a socket port
	time.Sleep(WaitForServerStart)

	request := rpc.NewRequest("/")
	err = fack.Sign(request, privateKey)
	if err != nil {
		t.Error(err)
	}

	/* GET request should succeed */
	resp, err := request.Send("GET", LocalHost+fmt.Sprint(GETPort))
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	if (*resp).GetDescription() != "authenticated" {
		t.Error("Node could not authenticate a valid host")
	}
}

// in the case where the user has both a global and local Permission bitmap set
// for a route, even if the global Permission bitmap denies access to an HTTP method
// the local Permission bitmap should take priority as an edge-case of permission elevation
func TestAuthGlobalAndLocalPermissionsPresent(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Error("Could not generate an ECDSA key pair")
	}

	address := fack.LocalHost().SetPort(8000)

	var auth *fack.Auth = fack.NewAuth()
	node := rpc.NewNode(address, auth) // pass a nil to logger pointer ~ no logging

	var endpoint *fack.Endpoint = fack.NewEndpoint("test", &privateKey.PublicKey)
	auth.AddTrusted("127.0.0.1", endpoint)

	globalPermissionMap := fack.NewPermission().NoAccess()
	endpoint.AddGlobalPermission(globalPermissionMap)

	localPermissionMap := fack.NewPermission().Enable(fack.GET)
	endpoint.AddLocalPermission("/", localPermissionMap)

	route := node.Function("/", AuthenticatedIndex).Method(fack.GET).Method(fack.POST).Auth(true)
	fmt.Println(route)

	go node.Start()

	// if you are on macos, you may need to give the binary permission to use a socket port
	time.Sleep(WaitForServerStart)

	request := rpc.NewRequest("/")
	err = fack.Sign(request, privateKey)
	if err != nil {
		t.Error(err.Error())
	}

	/* GET request should succeed */
	resp, err := request.Send("GET", LocalHost+fmt.Sprint(GETPort))
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	if (*resp).GetDescription() != "authenticated" {
		t.Error("Node could not authenticate a valid host")
	}

	/* POST request should NOT succeed */
	resp, err = request.Send("POST", LocalHost+fmt.Sprint(GETPort))
	if err != nil {
		t.Error("Failed to startup an HTTP GET route.")
	}

	if resp.GetStatus() != http.StatusUnauthorized {
		t.Error("Node was let into a permission ")
	}

}
