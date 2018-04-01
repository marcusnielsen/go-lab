package main

import (
	"encoding/json"
	"fmt"
	"net/HTTP"
	"strings"
)

type route struct {
	url     string
	modelFn string
	method  string
	args    []string
}

type provider struct {
	message string
	drivers drivers
}

type httpServer struct {
	mux            *http.ServeMux
	httpServerPort string
}

type drivers struct {
	httpServer
}

type component struct {
	routes   []route
	provider provider
}

type modelFnResult struct{ data string }

type model map[string]func(routeData) modelFnResult

type routeData map[string]string

type routeHandler struct {
	modelFn func(r routeData) modelFnResult
	route   route
}

type envVars struct {
	httpServerPort string
}

func getUser(routeData routeData) modelFnResult {
	return modelFnResult{data: "id: " + routeData["id"]}
}

func createUser(routeData routeData) modelFnResult {
	return modelFnResult{data: "name: " + routeData["name"]}
}

func initUserComponent(provider provider) model {
	model := model{
		"getUser":    getUser,
		"createUser": createUser,
	}

	routes := []route{
		route{url: "/users/create/", modelFn: "createUser", method: "post", args: []string{"name"}},
		route{url: "/users/get/", modelFn: "getUser", method: "get", args: []string{"id"}},
	}

	for _, route := range routes {
		provider.drivers.httpServer.handleRoute(route, model)
	}

	return model
}

func (httpServer httpServer) handleRoute(route route, model model) {
	modelFn := model[route.modelFn]

	routeHandler := routeHandler{
		modelFn,
		route,
	}

	httpServer.mux.Handle(route.url, routeHandler)
}

func (routeHandler routeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// @TODO: Get data from body or URL
	routeData := routeData{}

	if routeHandler.route.method == "get" {
		urlPath := r.URL.Path
		fmt.Println(urlPath)

		urlArgString := strings.TrimPrefix(urlPath, routeHandler.route.url)
		fmt.Println(urlArgString)

		urlArgs := strings.Split(urlArgString, "/")

		for i, urlArg := range urlArgs {
			routeData[routeHandler.route.args[i]] = urlArg
		}

		fmt.Println(routeData)
	}
	if routeHandler.route.method == "post" {
		var bodyArgs []string
		err := json.NewDecoder(r.Body).Decode(&bodyArgs)

		if err != nil {
			fmt.Println(err)
		}

		for i, bodyArg := range bodyArgs {
			routeData[routeHandler.route.args[i]] = bodyArg
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, routeHandler.modelFn(routeData).data)
}

func (httpServer httpServer) listenAndServe() {
	err := http.ListenAndServe(":"+httpServer.httpServerPort, httpServer.mux)
	if err != nil {
		fmt.Println(err)
	}
}

func makeProvider(envVars envVars) provider {
	mux := http.NewServeMux()
	httpServerPort := envVars.httpServerPort

	httpServer := httpServer{
		mux,
		httpServerPort,
	}

	drivers := drivers{
		httpServer,
	}

	provider := provider{
		drivers: drivers,
	}

	return provider
}

func main() {
	envVars := envVars{"4000"}

	provider := makeProvider(envVars)
	initUserComponent(provider)

	provider.drivers.httpServer.listenAndServe()
}
