package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

const CustomHostVariable = "GO_API_HOST"

const DefaultServerAddress = "https://aws-serverless-go-api.com"

type RequestAccessor struct {
	stripBasePath string
}

func (r *RequestAccessor) StripBasePath(basePath string) string {
	if strings.Trim(basePath, " ") == "" {
		r.stripBasePath = ""
		return ""
	}
	newBasePath := basePath

	for {
		if !strings.HasPrefix(newBasePath, "/") {
			break
		}
		newBasePath = strings.TrimPrefix(newBasePath, "/")
	}

	for {
		if !strings.HasSuffix(newBasePath, "/") {
			break
		}
		newBasePath = strings.TrimSuffix(newBasePath, "/")
	}

	newBasePath = "/" + newBasePath
	r.stripBasePath = newBasePath
	return newBasePath
}

func (r *RequestAccessor) ProxyEventToHTTPRequest(req events.LambdaFunctionURLRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeader(httpRequest, req)
}

func (r *RequestAccessor) EventToRequestWithContext(ctx context.Context, req events.LambdaFunctionURLRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContext(ctx, httpRequest, req), nil
}

func (r *RequestAccessor) EventToRequest(req events.LambdaFunctionURLRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		decodedBody = base64Body
	}

	path := req.RawPath
	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	serverAddress := DefaultServerAddress
	if customAddress, ok := os.LookupEnv(CustomHostVariable); ok {
		serverAddress = customAddress
	}

	path = serverAddress + path

	qs := url.Values{}
	mqu := func(es string) string {
		qes, err := url.QueryUnescape(es)
		if err != nil {
			log.Println(err)
			fmt.Printf("QueryUnescape error=%v", err)
			return es
		}
		return qes
	}

	if req.QueryStringParameters != nil && len(req.QueryStringParameters) > 0 {
		for q, v := range req.QueryStringParameters {
			qs.Add(mqu(q), mqu(v))
		}
	}

	if len(qs) > 0 {
		path += "?" + qs.Encode()
	}

	httpRequest, err := http.NewRequestWithContext(
		context.TODO(),
		strings.ToUpper(req.RequestContext.HTTP.Method),
		path,
		bytes.NewReader(decodedBody),
	)

	if err != nil {
		log.Printf("Could not convert request %s:%s to http.Request\n", req.RequestContext.HTTP.Method, req.RawPath)
		return nil, err
	}

	for h := range req.Headers {
		httpRequest.Header.Add(h, req.Headers[h])
	}

	return httpRequest, nil
}

func addToHeader(req *http.Request, apiGwRequest events.LambdaFunctionURLRequest) (*http.Request, error) {
	return req, nil
}

func addToContext(ctx context.Context, req *http.Request, apiGwRequest events.LambdaFunctionURLRequest) *http.Request {
	lc, _ := lambdacontext.FromContext(ctx)
	rc := requestContext{lambdaContext: lc, gatewayProxyContext: apiGwRequest.RequestContext}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

type ctxKey struct{}

type requestContext struct {
	lambdaContext       *lambdacontext.LambdaContext
	gatewayProxyContext events.LambdaFunctionURLRequestContext
}
