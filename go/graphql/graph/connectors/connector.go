package connectors

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/raywall/cloud-service-pack/go/adapters"
	"github.com/raywall/cloud-service-pack/go/graphql/types"
)

type ConnectorConfig struct {
	Field         string                 `json:"field"`
	Adapter       string                 `json:"adapter"`
	AdapterConfig map[string]interface{} `json:"adapterConfig"`
	KeyPattern    string                 `json:"keyPattern"`
}

type Config struct {
	Connectors []ConnectorConfig `json:"connectors"`
}

type Connector interface {
	GetData(args map[string]interface{}) (interface{}, error)
}

type connector struct {
	adapter    adapters.Adapter
	keyPattern string
}

func NewConnector(cfg *types.Config, config ConnectorConfig) (Connector, error) {
	var adapter adapters.Adapter
	attributes := config.AdapterConfig["attr"].(map[string]interface{})

	switch config.Adapter {
	case "redis":
		endpoint, _ := config.AdapterConfig["endpoint"].(string)
		password, _ := config.AdapterConfig["password"].(string)
		adapter = adapters.NewRedisAdapter(endpoint, password, config.KeyPattern, attributes)

	case "rest":
		baseUrl, _ := config.AdapterConfig["baseUrl"].(string)
		endpoint, _ := config.AdapterConfig["endpoint"].(string)
		adapter = adapters.NewRestAdapter(baseUrl, endpoint, auth, attributes)

	case "s3":
		region, _ := config.AdapterConfig["region"].(string)
		bucket, _ := config.AdapterConfig["bucket"].(string)
		accessKeyId, _ := config.AdapterConfig["accessKeyId"].(string)
		secretAccessKey, _ := config.AdapterConfig["secretAccessKey"].(string)
		adapter = adapters.NewS3Adapter(region, bucket, accessKeyId, secretAccessKey)

	case "dynamodb":
		region, _ := config.AdapterConfig["region"].(string)
		table, _ := config.AdapterConfig["table"].(string)
		accessKeyId, _ := config.AdapterConfig["accessKeyId"].(string)
		secretAccessKey, _ := config.AdapterConfig["secretAccessKey"].(string)
		adapter = adapters.NewDynamoDBAdapter(region, table, accessKeyId, secretAccessKey)

	default:
		return nil, fmt.Errorf("unsupported adapter: %s", config.Adapter)
	}

	return &connector{
		adapter:    adapter,
		keyPattern: config.KeyPattern,
	}, nil
}

func (c *connector) GetData(args map[string]interface{}) (interface{}, error) {
	if params, err := c.adapter.GetParameters(args); err != nil {
		return nil, err
	} else {
		return c.adapter.GetData(params)
	}
}

func LoadConnectors(cfg *types.Config, connectorConfig string) (map[string]Connector, error) {
	var config Config
	if err := json.Unmarshal([]byte(connectorConfig), &config); err != nil {
		return nil, fmt.Errorf("error parsing connectors config: %v", err)
	}

	connectors := make(map[string]Connector)
	for _, connConfig := range config.Connectors {
		conn, err := NewConnector(cfg, connConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating connector for %s: %v", connConfig.Field, err)
		}
		connectors[connConfig.Field] = conn
	}

	return connectors, nil
}
