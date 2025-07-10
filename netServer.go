package main

var routes = map[string]map[string]HandlerFunc{
	"GET":  {},
	"POST": {},
}

func Handle(method, path string, h HandlerFunc) {
	if routes[method] == nil {
		routes[method] = make(map[string]HandlerFunc)
	}
	routes[method][path] = h
}
