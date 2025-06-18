package graphql

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/cloud-service-pack/go/data"
)

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
	// Schema is the content or path to retrieve the schema settings of the GraphQL API that will be
	// created dynamically
	Schema string `json:"schema"`

	// Connectors is the content or path to retrieve settings from the GraphQL API resolver that will
	// be created dynamically
	Connectors string `json:"connectors"`

	// Route represents the route that will be used by the GraphQL API (e.g. /graphql)
	Route string `json:"route"`

	// Authorization contains the authorization settings to be used by GraphQL API connectors
	Authorization Authorization

	// CloudContext is the cloud context that will be used to interact with available cloud resources
	CloudContext cloud.CloudContext

	// Session is a AWS Session used by the service to interact with the cloud
	Session *session.Session
}

// GetSchemaValue is the method responsible for retrieving the schema settings from the GraphQL API
func (c *Config) GetSchemaValue() (string, error) {
	if c.Schema == "" {
		return "", fmt.Errorf("it's necessary to specify the GraphQL API schema")
	}

	if IsPath(c.Schema) {
		cfg, err := data.ParseConfig(c.Schema)
		if err != nil {
			return "", fmt.Errorf("failed to get inline configuration of schema: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", fmt.Errorf("failed to get the schema value: %v", err)
		}
		data, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("failed to deserialize the schema data: %v", err)
		}
		return string(data), nil
	}

	return c.Schema, nil
}

// GetConnectorsValue is the method responsible for retrieving the configurations of the GraphQL API connectors
func (c *Config) GetConnectorsValue() (string, error) {
	if c.Connectors == "" {
		return "", fmt.Errorf("it's necessary to specify the GraphQL API connections")
	}

	if IsPath(c.Connectors) {
		cfg, err := data.ParseConfig(c.Connectors)
		if err != nil {
			return "", fmt.Errorf("failed to get inline configuration of connectors: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", fmt.Errorf("failed to get the connectors value: %v", err)
		}
		data, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("failed to deserialize the connectors data: %v", err)
		}
		return string(data), nil
	}

	return c.Connectors, nil
}

func (c *Config) GetTokenServiceURL() (string, error) {
	authService := c.Authorization.TokenService.TokenAuthorizationURL
	if IsPath(authService) {
		cfg, err := data.ParseConfig(authService)
		if err != nil {
			return "", fmt.Errorf("failed to get inline configuration of authorization service url: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", fmt.Errorf("failed to get the authorization service value: %v", err)
		}
		authService = value["data"].(string)
	}

	return authService, nil
}

func (c *Config) GetCredentials() (string, string, error) {
	clientID := c.Authorization.TokenService.Credentials.ClientID
	if IsPath(clientID) {
		cfg, err := data.ParseConfig(clientID)
		if err != nil {
			return "", "", fmt.Errorf("failed to get inline configuration of authorization client id: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the authorization client id: %v", err)
		}
		clientID = value["client_id"].(string)
	}

	clientSecret := c.Authorization.TokenService.Credentials.ClientSecret
	if IsPath(clientSecret) {
		cfg, err := data.ParseConfig(clientID)
		if err != nil {
			return "", "", fmt.Errorf("failed to get inline configuration of authorization client secret: %v", err)
		}
		value, err := data.GetValue(cfg, c.Session)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the authorization client secret: %v", err)
		}
		clientSecret = value["client_secret"].(string)
	}

	return clientID, clientSecret, nil
}
