package types

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/cloud-service-pack/go/auth"
	"github.com/raywall/cloud-service-pack/go/data"
)

type MetricCollectorType int

const (
	DatadogCollector MetricCollectorType = iota
	OpenTelemetryCollector
)

// Basic represents the basic config necessary to the graphql observability
type Basic struct {
	// Team indicates the name of the team/squad responsable for this service
	Team string `json:"team"`

	// Solution indicates the name of the solution/application
	Solution string `json:"solution"`

	// Domain indicates the domain of the solution (DDD)
	Domain string `json:"domain"`

	// Product indicates the name of the product responsable for this service
	Product string `json:"product"`

	// Tags is an additional field to create custom tags not covered by the default object
	Tags map[string]string `json:"tags"`
}

// Credentials represents the credentials that will be used to generate a new token
type Credentials struct {
	// ClientID indicates the registered client application id
	ClientID string `json:"client_id"`

	// ClientSecret indicates the password of the registered client application
	ClientSecret string `json:"client_secret"`
}

// Tokenservice contains the settings needed to generate a token STS
type TokenService struct {
	// TokenAuthorizationURL represents the URL for the Token service
	TokenAuthorizationURL string `json:"token_authorization_url"`

	// Credentials representa as credenciais que serão utilizadas para gerar um novo token
	Credentials Credentials
}

// Authorization contains the authorization settings to be used by the Graphql API connectors
type Authorization struct {
	// RequireTokenSTS indicates whether your API will use an STS token to generate authentication tokens
	// for API connectors
	RequireTokenSTS bool `json:"require_token_sts"`

	// TokenService contains the settings required to generate an STS token
	TokenService TokenService
}

// Config contains all the configuration required to create and instantiate a dynamic GraphQL API
type Config struct {
	// Authorization contains the authorization settings to be used by GraphQL API connectors
	Authorization Authorization

	// BasicData is the basic information necessary to register the library observability
	BasicData *Basic `json:"basic"`

	// CloudContext is the cloud context that will be used to interact with available cloud resources
	CloudContext cloud.CloudContext

	// Connectors is the content or path to retrieve settings from the GraphQL API resolver that will
	// be created dynamically
	Connectors string `json:"connectors"`

	// Metrics indicates the metrics platform that will be used by the GraphQL Datadog or OpenTelemetry
	Metrics MetricCollectorType `json:"metrics"`

	// Route represents the route that will be used by the GraphQL API (e.g. /graphql)
	Route string `json:"route"`

	// Schema is the content or path to retrieve the schema settings of the GraphQL API that will be
	// created dynamically
	Schema string `json:"schema"`

	// Session is a AWS Session used by the service to interact with the cloud
	Session *session.Session

	// TokenManager is a auto managed token, responsable for manage and mantains the token always updated
	TokenManager *auth.TokenManager

	// AccessToken is the access token updated by the TokenManager
	AccessToken string
}

// GetSchemaValue is the method responsible for retrieving the schema settings from the GraphQL API
func (c *Config) GetSchemaValue() (string, error) {
	if c.Schema == "" {
		return "", fmt.Errorf("it's necessary to specify the GraphQL API schema")
	}

	if data.IsConfig(c.Schema) {
		cfg, err := data.ParseConfig(c.Schema)
		if err != nil {
			return "", fmt.Errorf("failed to get inline configuration of schema: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", fmt.Errorf("failed to get the schema value: %v", err)
		}
		return string(value), nil
	}

	return c.Schema, nil
}

// GetConnectorsValue is the method responsible for retrieving the configurations of the GraphQL API connectors
func (c *Config) GetConnectorsValue() (string, error) {
	if c.Connectors == "" {
		return "", fmt.Errorf("it's necessary to specify the GraphQL API connections")
	}

	if data.IsConfig(c.Connectors) {
		cfg, err := data.ParseConfig(c.Connectors)
		if err != nil {
			return "", fmt.Errorf("failed to get inline configuration of connectors: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", fmt.Errorf("failed to get the connectors value: %v", err)
		}
		return string(value), nil
	}

	return c.Connectors, nil
}

func (c *Config) GetTokenServiceURL() (string, error) {
	authService := c.Authorization.TokenService.TokenAuthorizationURL
	if data.IsConfig(authService) {
		cfg, err := data.ParseConfig(authService)
		if err != nil {
			return "", fmt.Errorf("failed to get inline configuration of authorization service url: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", fmt.Errorf("failed to get the authorization service value: %v", err)
		}
		return string(value), nil
	}

	return authService, nil
}

func (c *Config) GetCredentials() (string, string, error) {
	clientID := c.Authorization.TokenService.Credentials.ClientID
	if data.IsConfig(clientID) {
		cfg, err := data.ParseConfig(clientID)
		if err != nil {
			return "", "", fmt.Errorf("failed to get inline configuration of authorization client id: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the authorization client id: %v", err)
		}
		data, err := value.CastTo(data.JSON)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the authorization client id: %v", err)
		}
		clientID = ((data.(map[string]interface{}))["client_id"]).(string)
	}

	clientSecret := c.Authorization.TokenService.Credentials.ClientSecret
	if data.IsConfig(clientSecret) {
		cfg, err := data.ParseConfig(clientSecret)
		if err != nil {
			return "", "", fmt.Errorf("failed to get inline configuration of authorization client secret: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the authorization client secret: %v", err)
		}
		data, err := value.CastTo(data.JSON)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the authorization client secret: %v", err)
		}
		clientSecret = ((data.(map[string]interface{}))["client_secret"]).(string)
	}

	return clientID, clientSecret, nil
}
