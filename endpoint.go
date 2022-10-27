package fack

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/json"
)

const (
	Localhost    = "127.0.0.1"
	MissingNonce = 0
)

type Endpoint struct {
	Name              string `json:"name"`
	X509              string `json:"publicKey"`
	PublicKey         *ecdsa.PublicKey
	LastNonce         int64
	GlobalPermissions Permission            `json:"globalPermissions"`
	LocalPermissions  map[string]Permission `json:"localPermissions"`
}

func NewEndpoint(name string, publicKey *ecdsa.PublicKey) *Endpoint {
	endpoint := new(Endpoint)

	endpoint.Name = name
	endpoint.PublicKey = publicKey
	endpoint.LastNonce = MissingNonce
	endpoint.LocalPermissions = make(map[string]Permission)

	return endpoint
}

func (endpoint *Endpoint) AddGlobalPermission(permission Permission) {
	endpoint.GlobalPermissions = permission
}

func (endpoint *Endpoint) AddLocalPermission(route string, permission Permission) bool {
	if _, found := endpoint.LocalPermissions[route]; !found {
		endpoint.LocalPermissions[route] = permission
		return true
	}
	return false
}

func (endpoint *Endpoint) GetPublicKey() (*ecdsa.PublicKey, bool) {
	if (len(endpoint.X509) == 0) && (endpoint.PublicKey == nil) {
		return nil, false
	}

	if endpoint.PublicKey != nil {
		return endpoint.PublicKey, true
	}

	publicKeyByteArray, ok := StringToByte(endpoint.X509)
	if !ok {
		return nil, false
	}
	endpoint.GeneratePublicKey(publicKeyByteArray)

	return endpoint.PublicKey, true
}

func (endpoint *Endpoint) GeneratePublicKey(data []byte) bool {
	endpoint.X509 = ByteToString(data)

	// any ECDSA key stored in a byte format should be encoded using the x509 scheme
	// rather than the default ecdsa.Marshal encoding scheme
	publicKey, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return false
	}

	endpoint.PublicKey = publicKey.(*ecdsa.PublicKey)

	return true
}

func (endpoint *Endpoint) PublicKeyToBytes() []byte {
	if endpoint.PublicKey == nil {
		return []byte{}
	}

	b := elliptic.Marshal(endpoint.PublicKey.Curve, endpoint.PublicKey.X, endpoint.PublicKey.Y)
	return b
}

func (endpoint *Endpoint) ValidateSource(request Request) bool {
	// if we do not have a public key we cannot verify the ECDSA signature
	if endpoint.PublicKey == nil {
		return false
	}
	// we cannot accept the last received or previous nonce, or we risk a threat actor
	// resending an intercepted nonce/signature to forge credentials
	if request.Nonce() <= endpoint.LastNonce {
		return false
	}
	hash := request.Hash()
	return ecdsa.VerifyASN1(endpoint.PublicKey, hash[:], request.Signature())
}

func (endpoint Endpoint) HasPermissionToUseMethod(route, method string) bool {
	if localPermission, ok := endpoint.LocalPermissions[route]; ok {
		return localPermission.Check(method)
	} else {
		return endpoint.GlobalPermissions.Check(method)
	}
}

func (endpoint Endpoint) String() string {
	j, _ := json.Marshal(endpoint)
	return string(j)
}
