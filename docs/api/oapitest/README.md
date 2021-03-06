<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# oapitest

```go
import "github.com/ringsq/oapi-utils/oapitest"
```

Simple library for testing API requests and responses against an OpenAPI specification.  The example below uses a
mock DB interface created with https://github.com/maxbrunsfeld/counterfeiter.

```
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
```

## Index

- [type OapiRequest](<#type-oapirequest>)
- [type OapiTest](<#type-oapitest>)
  - [func GetOapiTester(spec *openapi3.T, server http.Handler) OapiTest](<#func-getoapitester>)


## type OapiRequest

```go
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
```

## type OapiTest

```go
type OapiTest interface {
    ValidateAPI(t *testing.T, req *OapiRequest) (*httptest.ResponseRecorder, error)
}
```

### func GetOapiTester

```go
func GetOapiTester(spec *openapi3.T, server http.Handler) OapiTest
```

GetOapiTester returns a tester for the provided OpenAPI spec and handler



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
