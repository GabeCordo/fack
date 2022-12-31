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
	Host string `json:"Host"`
	Port int    `json:"Port"`
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
	address.Host = split[0]
	address.Port = port

	return address, nil
}

// EmptyAddress
// Defaults to localhost Port 8080
func EmptyAddress() *Address {
	address := new(Address)

	address.Host = ""
	address.Port = 1

	return address
}

func LocalHost() *Address {
	address := new(Address)

	address.Host = "localhost"
	address.Port = 1

	return address
}

func (a *Address) SetHost(host string) *Address {
	runes := []rune(host)

	for _, rune := range runes {
		if !unicode.IsDigit(rune) && !unicode.IsLetter(rune) && (rune != '_') && (rune != '.') {
			panic(host + " is an invalid domain, this cannot be assigned as type Host")
		}
	}

	a.Host = host

	return a
}

func (a Address) GetHost() string {
	return a.Host
}

func (a Address) IsLocalHost() bool {
	return (a.Host == "127.0.0.1") || (a.Host == "localhost")
}

func (a *Address) SetPort(port int) *Address {
	if (port < MinPortValue) || (port > MaxPortValue) {
		panic("a Port must be within (inclusive) of 1 to 65535")
	}
	a.Port = port

	return a
}

func (a Address) GetPort() int {
	return a.Port
}

func (a Address) ToString() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
