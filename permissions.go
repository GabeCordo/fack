package fack

import "fmt"

type HTTPMethod uint8

const (
	GET    HTTPMethod = 0
	POST              = 1
	PULL              = 2
	DELETE            = 3
)

func HTTPMethodFromString(method string) HTTPMethod {
	if method == "GET" {
		return GET
	} else if method == "POST" {
		return POST
	} else if method == "PULL" {
		return PULL
	} else {
		return DELETE
	}
}

func IsValidHTTPMethod(method string) bool {
	return (method == "GET") || (method == "POST") || (method == "PULL") || (method == "DELETE")
}

type Permission [4]bool

func NewPermission(method ...bool) *Permission {
	permission := new(Permission)

	if len(method) > 4 {
		panic("only 4 permission options possible (GET, POST, PULL, DELETE)")
	}

	//
	for i, enabled := range method {
		permission[i] = enabled
	}

	return permission
}

func (permission *Permission) FullAccess() *Permission {
	for i := range permission {
		permission[i] = true
	}

	return permission
}

func (permission *Permission) NoAccess() *Permission {
	for i := range permission {
		permission[i] = false
	}

	return permission
}

func (permission *Permission) Enable(method HTTPMethod) *Permission {
	permission[method] = true

	return permission
}

func (permission *Permission) Disable(method HTTPMethod) *Permission {
	permission[method] = false

	return permission
}

func (permission Permission) IsEnabled(method HTTPMethod) bool {
	return permission[method]
}

func (permission Permission) String() string {
	template := "Permission[%t, %t, %t, %t]"
	return fmt.Sprintf(template, permission[GET], permission[POST], permission[PULL], permission[DELETE])
}
