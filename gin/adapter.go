package ginadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/e2u/aws-lambda-go-function-url/core"
	"github.com/gin-gonic/gin"
)

type GinLambda struct {
	core.RequestAccessor
	ginEngine *gin.Engine
}

func New(gin *gin.Engine) *GinLambda {
	return &GinLambda{ginEngine: gin}
}

func (g *GinLambda) Proxy(req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	ginRequest, err := g.ProxyEventToHTTPRequest(req)
	return g.proxyInternal(ginRequest, err)
}

func (g *GinLambda) ProxyWithContext(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	ginRequest, err := g.EventToRequestWithContext(ctx, req)
	return g.proxyInternal(ginRequest, err)
}

func (g *GinLambda) proxyInternal(req *http.Request, err error) (events.LambdaFunctionURLResponse, error) {
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert LambdaFunctionURLRequest to request: %v", err)
	}
	respWriter := core.NewResponseWriter()
	g.ginEngine.ServeHTTP(http.ResponseWriter(respWriter), req)
	proxyResponse, err := respWriter.GetResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating LambdaFunctionURLResponse response: %v", err)
	}
	return proxyResponse, nil
}
