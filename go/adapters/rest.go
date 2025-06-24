package adapters

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/raywall/cloud-service-pack/go/graphql/types"
)

type RestAdapter interface {
	Adapter
}

type restAdapter struct {
	client      *http.Client
	accessToken *string
	baseUrl     string
	endpoint    string
	auth        bool
	attr        map[string]interface{}
	headers     map[string]interface{}
}

func NewRestAdapter(cfg *types.Config, baseUrl, endpoint string, auth bool, attributes, headers map[string]interface{}) RestAdapter {
	return &restAdapter{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		accessToken: &cfg.AccessToken,
		baseUrl:     baseUrl,
		endpoint:    endpoint,
		headers:     headers,
		attr:        attributes,
		auth:        auth,
	}
}

func (r *restAdapter) GetData(args []AdapterAttribute) (interface{}, error) {
	route := r.endpoint
	if re.MatchString(route) {
		for _, attr := range args {
			route = strings.ReplaceAll(
				route,
				fmt.Sprintf("{%s}", attr.Name),
				fmt.Sprintf("%v", attr.Value))
		}
	}

	url := fmt.Sprintf("%s/%s", r.baseUrl, route)
	req, _ := http.NewRequest("GET", url, nil)

	for key, value := range r.headers {
		finalValue := value.(string)
		if re.MatchString(finalValue) {
			for _, attr := range args {
				finalValue = strings.ReplaceAll(
					finalValue,
					fmt.Sprintf("{%s}", attr.Name),
					fmt.Sprintf("%v", attr.Value))
			}
		}
		req.Header.Add(key, finalValue)
	}

	if r.auth {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *r.accessToken))
	}

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from REST API %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("REST API returned status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read REST API response: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to decode REST API response: %v", err)
	}

	// validar se exige ou nao data
	return data["data"], nil
}

func (r *restAdapter) GetParameters(args map[string]interface{}) ([]AdapterAttribute, error) {
	return getParameters(r.attr, args)
}
