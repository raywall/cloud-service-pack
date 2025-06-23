package adapters

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type RedisAdapter interface {
	Adapter
}

type redisAdapter struct {
	client *redis.Client
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
		attr: attributes,
		keyPattern: keyPattern,
	}
}

func (r *redisAdapter) GetData(args []AdapterAttribute) (interface{}, error) {
	if arts == nil || len(args) == 0 {
		return nil, errors.New("the data key value was not informed")
	}

	searchKey := strings.ReplaceAll(
		r.keyPattern,
		fmt.Sprintf("(%s)", args[0].Name),
		fmt.Sprintf("%w", args[0].Value))

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
