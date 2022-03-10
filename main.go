package main

// StringService provides operations on strings.
// ===============================================================
// EXAMPLE USAGE
// ===============================================================
// $ curl -XPOST -d'{"s":"hello, world"}' localhost:8080/uppercase
// {"v":"HELLO, WORLD"}
// $ curl -XPOST -d'{"s":"hello, world"}' localhost:8080/count
// {"v":12}
// ===============================================================

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httpTransport "github.com/go-kit/kit/transport/http"
)

// ============ BUSINESS LOGIC ============
// >>> Go kit models SERVICES as an INTERFACE.

type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
}

type stringService struct{}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	return strings.ToUpper(s), nil
}

func (stringService) Count(s string) int {
	return len(s)
}

// ErrEmpty is returned when input string is empty
var ErrEmpty = errors.New("empty string")

// ============ REQUESTS / RESPONSES ============
// >>> The primary messaging parttern is RPC (remote procedure call).
// For each method, request and response structs are defined which
// capture all of the input and output parameters.

type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"` // errors don't JSON-marshal, so we use a string
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}

// ============ ENDPOINTS ============
// >>> Go kit provides most functionality through an abstraction called
// an ENDPOINT. It represents a single RPC (i.e. a single method in the
// service interface). ADAPTERS convert each service method into an
// endpoint. Each ADAPTER takes a StringService and returns an ENDPOINT
// that corresponds to one of the methods.

func makeUppercaseEndpoint(svc StringService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(uppercaseRequest)
		v, err := svc.Uppercase(req.S)
		if err != nil {
			return uppercaseResponse{v, err.Error()}, nil
		}
		return uppercaseResponse{v, ""}, nil
	}
}

func makeCountEndpoint(svc stringService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(countRequest)
		v := svc.Count(req.S)
		return countResponse{v}, nil
	}
}

// ============ TRANSPORTS ============
// >>> Expose service to the outside world! Go kit supports many TRANSPORTS
// out-of-the-box. This example uses JSON over HTTP. Go kit provides a
// helper struct in package 'transport/http'.

func main() {
	svc := stringService{}

	uppercaseHandler := httpTransport.NewServer(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		encodeResponse,
	)

	countHandler := httpTransport.NewServer(
		makeCountEndpoint(svc),
		decodeCountRequest,
		encodeResponse,
	)

	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func decodeUppercaseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request countRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
