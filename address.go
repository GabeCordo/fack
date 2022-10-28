package fack

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	StringToAddressConversionError = "could not convert invalid string to address structure"
	MinPortValue                   = 1
	MaxPortValue                   = 65535
)

type Address struct {
	host string `json:"host"`
	port int    `json:"port"`
}

func NewAddress(ip string) (*Address, error) {
	if !strings.Contains(ip, ":") {
		return nil, errors.New(StringToAddressConversionError)
	}
	split := strings.Split(ip, ":")
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, errors.New(StringToAddressConversionError)
	}

	address := new(Address)
	address.host = split[0]
	address.port = port

	return address, nil
}

// EmptyAddress
// Defaults to localhost port 8080
func EmptyAddress() *Address {
	address := new(Address)

	address.host = ""
	address.port = 1

	return address
}

func LocalHost() *Address {
	address := new(Address)

	address.host = "localhost"
	address.port = 1

	return address
}

func (a *Address) Host(host string) *Address {
	runes := []rune(host)

	for _, rune := range runes {
		if !unicode.IsDigit(rune) && !unicode.IsLetter(rune) && (rune != '_') && (rune != '.') {
			panic(host + " is an invalid domain, this cannot be assigned as type host")
		}
	}

	a.host = host

	return a
}

func (a Address) GetHost() string {
	return a.host
}

func (a Address) IsLocalHost() bool {
	return (a.host == "127.0.0.1") || (a.host == "localhost")
}

func (a *Address) Port(port int) *Address {
	if (port < MinPortValue) || (port > MaxPortValue) {
		panic("a port must be within (inclusive) of 1 to 65535")
	}
	a.port = port

	return a
}

func (a Address) GetPort() int {
	return a.port
}

func (a Address) ToString() string {
	return fmt.Sprintf("%s:%d", a.host, a.port)
}
