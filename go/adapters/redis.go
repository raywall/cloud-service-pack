package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

type RedisAdapter interface {
	Adapter
}

type redisAdapter struct {
	client     *redis.Client
	keyPattern string
	attr       map[string]interface{}
}

func NewRedisAdapter(endpoint, pass, keyPattern string, attributes map[string]interface{}) RedisAdapter {
	return &redisAdapter{
		client: redis.NewClient(
			&redis.Options{
				Addr:     endpoint,
				Password: pass,
				DB:       0,
			},
		),
		attr:       attributes,
		keyPattern: keyPattern,
	}
}

func (r *redisAdapter) GetData(args []AdapterAttribute) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("the data key value was not informed")
	}

	searchKey := strings.ReplaceAll(
		r.keyPattern,
		fmt.Sprintf("{%s}", args[0].Name),
		fmt.Sprintf("%v", args[0].Value))

	data, err := r.client.Get(context.Background(), searchKey).Result()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *redisAdapter) GetParameters(args map[string]interface{}) ([]AdapterAttribute, error) {
	return getParameters(r.attr, args)
}
