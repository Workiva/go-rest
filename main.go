package main

import (
	"encoding/json"
	"fmt"
	"go-rest/rest"
	"go-rest/server"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
)

type Consumer struct {
	Key    string
	Secret string
}

func (c Consumer) Authorize(urlStr string, requestType string, form url.Values) url.Values {
	return form
}

type Foo struct {
	Foo string
	Bar float64
}

func (f Foo) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"foo": f.Foo,
		"bar": f.Bar,
	})
}

type FooHandler struct{}

func (f FooHandler) EndpointName() string {
	return "foos"
}

func (f FooHandler) ReadResource(params *server.ReadParams) (interface{}, error) {
	if params.ResourceId == "42" {
		return &Foo{"hello", 42}, nil
	}

	return nil, fmt.Errorf("No resource with id %s", params.ResourceId)
}

func (f FooHandler) CreateResource(params *server.CreateParams) (interface{}, error) {
	foo := &Foo{Foo: params.Data["foo"].(string), Bar: params.Data["bar"].(float64)}
	return foo, nil
}

func (f FooHandler) UpdateResource(params *server.UpdateParams) (interface{}, error) {
	foo := &Foo{Foo: params.Data["foo"].(string), Bar: params.Data["bar"].(float64)}
	return foo, nil
}

func (f FooHandler) DeleteResource(params *server.DeleteParams) (interface{}, error) {
	foo := &Foo{}
	return foo, nil
}

func main() {
	if os.Args[1] == "1" {
		r := mux.NewRouter()
		server.RegisterResourceHandler(r, FooHandler{})
		http.Handle("/", r)
		http.ListenAndServe(":8080", nil)
	}

	rc := rest.Client{Consumer{"key", "secret"}}
	params := map[string]string{
		"foo": "bar",
	}
	resp, err := rc.Get("http://localhost:8080/api/v0.1/foos/1", params)
	fmt.Println(resp)
	fmt.Println(err)
}
