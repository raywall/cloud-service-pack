package graphql

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	gp "github.com/graphql-go/graphql"

	"github.com/raywall/cloud-easy-connector/pkg/auth"
	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/cloud-service-pack/go/graphql/graph"
)

// route is the API route name that will be used by default
const route string = "/graphql"

type GraphQL struct {
	tokenManager auth.AutoManagedToken `json:"-"`
	AccessToken  *string               `json:"token"`
	Config       Config                `json:"config"`
	Resolver     *graph.Resolver       `json:"resolver"`
	Schema       *gp.Schema            `json:"schema"`
}

func New(config *Config, resources *cloud.CloudContextList, region, endpoint string) (*GraphQL, error) {
	var (
		err error
		api = GraphQL{}
	)

	// route
	if config.Route == "" {
		config.Route = route
	}

	// cloud context
	if region == "" {
		return nil, fmt.Errorf("it's necessary to inform the AWS region you want to use")
	}

	// aws session
	config.Session, err = session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao criar sessão AWS: %w", err)
	}
	if endpoint != "" {
		config.Session.Config.Endpoint = aws.String(endpoint)
	}

	config.CloudContext, err = cloud.NewAwsCloudContext(region, endpoint, resources)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new AWS Cloud Context: %v", err)
	}

	// connections
	connectionsConfig, err := config.GetConnectorsValue()
	if err != nil {
		return nil, fmt.Errorf("failed to get the connections config: %v", err)
	}

	res, err := graph.NewResolver(connectionsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a resolver: %v", err)
	}
	api.Resolver = &res

	// schema
	schemaConfig, err := config.GetSchemaValue()
	if err != nil {
		return nil, fmt.Errorf("failed to get the schema config: %v", err)
	}

	api.Schema, err = graph.CreateSchema(res, schemaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a schema: %v", err)
	}

	// token
	if config.Authorization.RequireTokenSTS {
		// auth_service_url
		authServiceUrl, err := config.GetTokenServiceURL()
		if err != nil {
			return nil, err
		}

		// credentials
		clientID, clientSecret, err := config.GetCredentials()
		if err != nil {
			return nil, err
		}

		config.CloudContext.NewAutoManagedToken(
			authServiceUrl,
			clientID,
			clientSecret,
			false)
	}

	return &api, nil
}
