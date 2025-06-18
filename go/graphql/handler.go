package graphql

import (
	"net/http"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/graphql-go/handler"
	"github.com/raywall/cloud-service-pack/go/graphql/middleware"
)

func (g *GraphQL) NewHandler(pretty bool, middlewares ...middleware.Middleware) http.Handler {
	// Configurar o handler GraphQL
	h := handler.New(
		&handler.Config{
			Schema:   g.Schema,
			Pretty:   pretty,
			GraphiQL: true,
		})

	// Aplicar middleware chain
	return middleware.Chain(
		h,
		// middleware.Logging,
		// middleware.Tracing,
	)
}

func (g *GraphQL) ToAmazonALB(handle http.Handler) *httpadapter.HandlerAdapterALB {
	return httpadapter.NewALB(handle)
}

func (g *GraphQL) ToAmazonAPIGateway(handle http.Handler) *httpadapter.HandlerAdapter {
	return httpadapter.New(handle)
}

func (g *GraphQL) ToAmazonAPIGatewayV2(handle http.Handler) *httpadapter.HandlerAdapterV2 {
	return httpadapter.NewV2(handle)
}
