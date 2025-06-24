package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/cloud-service-pack/go/graphql"
	"github.com/raywall/cloud-service-pack/go/graphql/types"
)

var (
	adapter        *httpadapter.HandlerAdapterALB
	wrappedHandler http.Handler
	err            error
	api            *graphql.GraphQL
)

func init() {
	resources := &cloud.CloudContextList{
		cloud.SSMContext,
		cloud.SecretsManagerContext,
	}

	config := &types.Config{
		Authorization: types.Authorization{
			RequireTokenSTS: false,
			TokenService: types.TokenService{
				TokenAuthorizationURL: "https://sts.teste.net/api/oauth/token",
				Credentials: types.Credentials{
					ClientID:     "aws::secrets::/graphql/dev/credentials::json",
					ClientSecret: "aws::secrets::/graphql/dev/credentials::json",
				},
				InsecureSkipVerify: false,
			},
		},
		BasicData: &types.Info{
			Team:     "Squad",
			Solution: "Solution",
			Domain:   "Domain",
			Product:  "Product",
		},
		Metrics:    types.OpenTelemetryCollector,
		Connectors: "local::file::/Users/macmini/Documents/workspace/projetos/cloud-service-pack/examples/config/graphql/connectors.json",
		Route:      "/graphql",
		Schema:     "local::file::/Users/macmini/Documents/workspace/projetos/cloud-service-pack/examples/config/graphql/schema.json",
	}

	api, err = graphql.New(config, resources, "us-east-1", "http://localhost:4566")
	if err != nil {
		panic(err)
	}

	// Configurar o handler GraphQL
	wrappedHandler = api.NewHandler(true)

	// Adaptar o handler para Lambda
	adapter = api.ToAmazonALB(wrappedHandler)
}

func requestHandler(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	method := req.HTTPMethod
	path := req.Path

	if path == "/health" && method == http.MethodGet {
		return events.ALBTargetGroupResponse{
			StatusCode: http.StatusOK,
			Body:       `{"status": "ok"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil

	} else if path != api.Config.Route && method != http.MethodPost {
		return events.ALBTargetGroupResponse{
			StatusCode: 404,
			Body:       `{"message": "rota não encontrada ou método não permitido"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	if _, ok := os.LookupEnv("ENVIRONMENT"); ok {
		lambda.Start(requestHandler)
	} else {
		http.Handle("/graphql", wrappedHandler)
		fmt.Println("Server running at http://localhost:8080/graphql")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
