package fack

type Route struct {
	path         string
	access       Permission
	Debug        bool
	RequiresAuth bool
}

func NewRoute(path string) *Route {
	route := new(Route)
	route.path = path

	return route
}

func EmptyRoute() *Route {
	route := new(Route)
	return route
}

func (route *Route) Path(path string) *Route {
	route.path = path

	return route
}

func (route *Route) Auth(enable bool) *Route {
	route.RequiresAuth = true

	return route
}

func (route *Route) Method(method HTTPMethod) *Route {
	route.access.Enable(method)

	return route
}

func (route Route) IsMethodSupported(method HTTPMethod) bool {
	return route.access.IsEnabled(method)
}
