package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type CreateParams struct {
	Data    map[string]interface{}
	Version string
}

type ReadParams struct {
	ResourceId string
	Version    string
}

type UpdateParams struct {
	ResourceId string
	Data       map[string]interface{}
	Version    string
}

type DeleteParams struct {
	ResourceId string
	Version    string
}

type ResourceHandler interface {
	EndpointName() string
	ReadResource(*ReadParams) (interface{}, error)
	CreateResource(*CreateParams) (interface{}, error)
	UpdateResource(*UpdateParams) (interface{}, error)
	DeleteResource(*DeleteParams) (interface{}, error)
}

func RegisterResourceHandler(router *mux.Router, r ResourceHandler) {
	urlBase := fmt.Sprintf("/api/v{version:[^/]+}/%s", r.EndpointName())
	router.HandleFunc(urlBase, handleCreate(r.CreateResource)).Methods("POST")
	router.HandleFunc(urlBase+"/{resource_id}", handleRead(r.ReadResource)).Methods("GET")
	router.HandleFunc(urlBase+"/{resource_id}", handleUpdate(r.UpdateResource)).Methods("PUT")
	router.HandleFunc(urlBase+"/{resource_id}", handleDelete(r.DeleteResource)).Methods("DELETE")
}

func handleCreate(createFunc func(*CreateParams) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		version := vars["version"]
		format := "json"
		if format, ok := r.URL.Query()["format"]; ok {
			format = format
		}

		serializer, err := responseSerializer(format)
		if err != nil {
			sendResponse(nil, w, nil, err, http.StatusInternalServerError)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			sendResponse(serializer, w, nil, err, http.StatusInternalServerError)
			return
		}

		createParams := &CreateParams{Data: data, Version: version}
		resource, err := createFunc(createParams)

		sendResponse(serializer, w, resource, err, http.StatusCreated)
	}
}

func handleRead(readFunc func(*ReadParams) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		resource_id := vars["resource_id"]
		version := vars["version"]
		format := "json"
		if format, ok := r.URL.Query()["format"]; ok {
			format = format
		}

		serializer, err := responseSerializer(format)
		if err != nil {
			sendResponse(nil, w, nil, err, http.StatusInternalServerError)
			return
		}

		readParams := &ReadParams{ResourceId: resource_id, Version: version}
		resource, err := readFunc(readParams)

		sendResponse(serializer, w, resource, err, http.StatusOK)
	}
}

func handleUpdate(updateFunc func(*UpdateParams) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		resource_id := vars["resource_id"]
		version := vars["version"]
		format := "json"
		if format, ok := r.URL.Query()["format"]; ok {
			format = format
		}

		serializer, err := responseSerializer(format)
		if err != nil {
			sendResponse(nil, w, nil, err, http.StatusInternalServerError)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			sendResponse(serializer, w, nil, err, http.StatusInternalServerError)
			return
		}

		updateParams := &UpdateParams{
			ResourceId: resource_id,
			Data:       data,
			Version:    version,
		}
		resource, err := updateFunc(updateParams)

		sendResponse(serializer, w, resource, err, http.StatusOK)
	}
}

func handleDelete(deleteFunc func(*DeleteParams) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		resource_id := vars["resource_id"]
		version := vars["version"]
		format := "json"
		if format, ok := r.URL.Query()["format"]; ok {
			format = format
		}

		serializer, err := responseSerializer(format)
		if err != nil {
			sendResponse(nil, w, nil, err, http.StatusInternalServerError)
			return
		}

		deleteParams := &DeleteParams{ResourceId: resource_id, Version: version}
		resource, err := deleteFunc(deleteParams)

		sendResponse(serializer, w, resource, err, http.StatusOK)
	}
}

func responseSerializer(format string) (ResponseSerializer, error) {
	var serializer ResponseSerializer
	switch format {
	case "json":
		serializer = JsonSerializer{}
	default:
		return nil, fmt.Errorf("Format not implemented: %s", format)
	}

	return serializer, nil
}
