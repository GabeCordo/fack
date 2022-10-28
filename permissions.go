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

type Permission [4]bool

func NewPermission(get, post, pull, delete bool) *Permission {
	permission := new(Permission)

	permission[0] = get
	permission[1] = post
	permission[2] = pull
	permission[3] = delete

	return permission
}
func (permission *Permission) Enable(method HTTPMethod) {
	permission[method] = true
}

func (permission *Permission) Disable(method HTTPMethod) {
	permission[method] = false
}

func (permission Permission) IsEnabled(method HTTPMethod) bool {
	return permission[method]
}

func (permission Permission) String() string {
	template := "Permission[%t, %t, %t, %t]"
	return fmt.Sprintf(template, permission[GET], permission[POST], permission[PULL], permission[DELETE])
}
