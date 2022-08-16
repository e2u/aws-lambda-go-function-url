package core

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-playground/assert/v2"
)

func TestEventToRequest(t *testing.T) {
	readData := func(filename string) events.LambdaFunctionURLRequest {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("read test data file error=%v", err)
		}
		rs := events.LambdaFunctionURLRequest{}
		if err := json.Unmarshal(b, &rs); err != nil {
			t.Fatalf("parse test data file error=%v", err)
		}
		return rs
	}

	t.Run("get-01-Multi value headers-Enabled", func(t *testing.T) {
		t.Setenv(CustomHostVariable, "https://www.custom-domain.com")
		event := readData("../data/get-01.json")
		r := &RequestAccessor{}
		req, err := r.EventToRequest(event)
		if err != nil {
			t.Fatal(err)
		}
		//out.Method
		assert.Equal(t, req.Method, "GET")
		assert.Equal(t, req.Host, "www.custom-domain.com")
		// assert.Equal(t, req.Header.Get("Cache-Control"), "no-cache")
	})
}
