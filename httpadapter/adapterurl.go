package httpadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/e2u/aws-lambda-go-function-url/core"
)

type HandlerAdapter struct {
	core.RequestAccessor
	handler http.Handler
}

func New(handler http.Handler) *HandlerAdapter {
	return &HandlerAdapter{
		handler: handler,
	}
}

func (h *HandlerAdapter) Proxy(event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	req, err := h.EventToRequest(event)
	return h.proxyInternal(req, err)
}

func (h *HandlerAdapter) ProxyWithContext(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	req, err := h.EventToRequestWithContext(ctx, event)
	return h.proxyInternal(req, err)
}

func (h *HandlerAdapter) proxyInternal(req *http.Request, err error) (events.LambdaFunctionURLResponse, error) {
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewResponseWriter()
	h.handler.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return resp, nil
}
