package server

import (
	"encoding/json"
	"fmt"
	"go-rest/server/context"
	"net/http"
)

type Response map[string]interface{}

// ResponseSerializer is responsible for serializing REST responses and sending
// them back to the client.
type ResponseSerializer interface {
	SendSuccessResponse(http.ResponseWriter, Response, int)
	SendErrorResponse(http.ResponseWriter, error, int)
}

type JsonSerializer struct{}

func (j JsonSerializer) SendErrorResponse(w http.ResponseWriter, err error, errorCode int) {
	response := makeErrorResponse(err)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	w.Write(jsonResponse)
}

func (j JsonSerializer) SendSuccessResponse(w http.ResponseWriter, r Response, status int) {
	jsonResponse, err := json.Marshal(r)
	if err != nil {
		j.SendErrorResponse(w, err, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}

func makeSuccessResponse(r interface{}, cursor string) Response {
	response := Response{
		"success": true,
		"result":  r,
	}

	if cursor != "" {
		response["next"] = cursor
	}

	return response
}

func makeErrorResponse(err error) Response {
	return Response{
		"success": false,
		"error":   err.Error(),
	}
}

func sendResponse(w http.ResponseWriter, ctx context.RequestContext) {
	status := ctx.Status()
	requestError := ctx.Error()
	cursor := ctx.Cursor()
	result := ctx.Result()

	serializer, err := responseSerializer(ctx.ResponseFormat())
	if err != nil {
		// Fall back to json serialization.
		serializer = JsonSerializer{}
		status = http.StatusNotImplemented
		requestError = err
	}

	if requestError != nil {
		if status < 400 {
			status = http.StatusInternalServerError
		}
		serializer.SendErrorResponse(w, requestError, status)
		return
	}

	serializer.SendSuccessResponse(w, makeSuccessResponse(result, cursor), status)
}

// responseSerializer returns a ResponseSerializer for the given format type. If the format
// is not implemented, the returned serializer will be nil and the error set.
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
