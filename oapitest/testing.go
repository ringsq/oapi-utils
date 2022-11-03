/*
Simple library for testing API requests and responses against
an OpenAPI specification.

	import "github.com/ringsq/oapi-utils/oapitest"

	// Get the OpenAPI spec
	spec := GetSpec()

	// setup example response struct
	circuits := &models.GetSearchResponse{}
	gofakeit.Struct(circuits)

	// Create the HTTP handler
	server := getServer()

	// use a mock DB instance
	db := new(modelsfakes.FakeDBOperations)
	server.db = db

	cfg := oapitest.GetOapiTester(spec, server)

	// set the mock DB instance results
	db.SearchCircuitsReturnsOnCall(0, *circuits, fmt.Errorf("err1"))
	db.SearchCircuitsReturnsOnCall(1, *circuits, nil)

	for i := 0; i < 2; i++ {
			req := &oapitest.OapiRequest{Method: "GET", Path: "/search?page=1&per_page=1&vendor=r"}
			response, err := cfg.ValidateAPI(t, req)
			if err != nil {
					require.Fail(t, err.Error())
			}
			log.Info("testing ", i)
			log.Info("status", response.Code)
	}


*/
package oapitest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	log "github.com/ringsq/go-logger"
)

type oapiTester struct {
	server  http.Handler
	baseURL string
	Spec    *openapi3.T
}

type OapiTest interface {
	ValidateAPI(t *testing.T, req *OapiRequest) (*httptest.ResponseRecorder, error)
}

type OapiRequest struct {
	// Method is the HTTP method to test (GET, POST, DELETE, etc)
	Method string
	// Path is the server path to test
	Path string
	// Body is the request body for PUT, PATCH, etc.
	Body io.Reader
	// SkipStatus will cause the validator to not validate that the response
	// code is one of the defined responses for the request.  If set to true
	// any response code is accepted
	SkipStatus bool
}

// GetOapiTester returns a tester for the provided OpenAPI spec and handler
func GetOapiTester(spec *openapi3.T, server http.Handler) OapiTest {
	oapiTester := &oapiTester{Spec: spec, server: server}
	// Set the base URL to the first server listed in thne spec
	if len(spec.Servers) > 0 {
		oapiTester.baseURL = spec.Servers[0].URL
	}
	return oapiTester
}

// ValidateAPI will test the given request and response against the OpenAPI spec
// It returns the response recorder so that the Code and Body can be further inspected
func (tc *oapiTester) ValidateAPI(t *testing.T, req *OapiRequest) (*httptest.ResponseRecorder, error) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(req.Method, req.Path, req.Body)

	// We make a second request so that we can manipulate the path for handling the serverURL in the spec
	oapiReq := *request
	oapiURL, err := url.Parse(tc.baseURL + oapiReq.URL.String())
	if err != nil {
		return response, err
	}
	log.Debug(oapiURL)
	oapiReq.URL = oapiURL
	ctx := request.Context()
	doc := tc.Spec
	if err := doc.Validate(ctx); err != nil {
		return response, err
	}
	router, err := legacyrouter.NewRouter(doc)
	if err != nil {
		return response, err
	}
	route, pathParams, err := router.FindRoute(&oapiReq)
	if err != nil {
		return response, err
	}
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    request,
		PathParams: pathParams,
		Route:      route,
	}
	tc.server.ServeHTTP(response, request)
	// assert.Equal(t, 200, w.Code, "Unexpected return code: %d", w.Code)
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 response.Code,
		Header:                 http.Header{"Content-Type": []string{"application/json"}},
		Options:                &openapi3filter.Options{IncludeResponseStatus: !req.SkipStatus},
	}
	if response.Body != nil {
		if len(response.Body.Bytes()) > 0 {
			responseValidationInput.SetBodyBytes(response.Body.Bytes())
			// Validate response.
			if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
				t.Error(err)
				t.Errorf("ResponseBody: %v", response.Body)
				return response, err
			}
		}
	}
	return response, nil
}
