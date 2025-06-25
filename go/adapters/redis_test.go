package adapters

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

// Testes para RedisAdapter
func TestNewRedisAdapter(t *testing.T) {
	adapter := NewRedisAdapter("localhost:6379", "password", "user:{userId}", map[string]interface{}{
		"userId": "string",
	})

	if adapter == nil {
		t.Fatal("NewRedisAdapter retornou nil")
	}

	redisAdapterImpl, ok := adapter.(*redisAdapter)
	if !ok {
		t.Fatal("adapter não é do tipo *redisAdapter")
	}

	if redisAdapterImpl.keyPattern != "user:{userId}" {
		t.Errorf("keyPattern = %v, esperado %v", redisAdapterImpl.keyPattern, "user:{userId}")
	}

	if len(redisAdapterImpl.attr) != 1 {
		t.Errorf("len(attr) = %d, esperado 1", len(redisAdapterImpl.attr))
	}
}

func TestRedisAdapter_GetParameters(t *testing.T) {
	attributes := map[string]interface{}{
		"userId": "string",
		"type":   "string",
	}

	adapter := NewRedisAdapter("localhost:6379", "", "user:{userId}", attributes)

	args := map[string]interface{}{
		"userId": "123",
		"type":   "premium",
	}

	result, err := adapter.GetParameters(args)
	if err != nil {
		t.Fatalf("GetParameters() erro = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("len(result) = %d, esperado 2", len(result))
	}

	// Verificar se os parâmetros corretos foram retornados
	expectedParams := []AdapterAttribute{
		{Name: "userId", Type: "string", Value: "123"},
		{Name: "type", Type: "string", Value: "premium"},
	}

	for _, expected := range expectedParams {
		found := false
		for _, actual := range result {
			if actual.Name == expected.Name &&
				actual.Type == expected.Type &&
				reflect.DeepEqual(actual.Value, expected.Value) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("parâmetro esperado não encontrado: %+v", expected)
		}
	}
}

func TestRedisAdapter_GetData_Success(t *testing.T) {
	// Usar miniredis para simular Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("erro ao iniciar miniredis: %v", err)
	}
	defer mr.Close()

	// Preparar dados de teste
	testData := map[string]interface{}{
		"id":   "123",
		"name": "John Doe",
		"age":  30,
	}
	jsonData, _ := json.Marshal(testData)

	// Inserir dados no Redis mock
	mr.Set("user:123", string(jsonData))

	// Criar adapter
	adapter := &redisAdapter{
		client: redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		}),
		keyPattern: "user:{userId}",
		attr: map[string]interface{}{
			"userId": "string",
		},
	}

	// Preparar argumentos
	args := []AdapterAttribute{
		{Name: "userId", Type: "string", Value: "123"},
	}

	// Executar teste
	result, err := adapter.GetData(args)
	if err != nil {
		t.Fatalf("GetData() erro = %v", err)
	}

	// Verificar resultado
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("resultado não é map[string]interface{}")
	}

	if resultMap["id"] != "123" {
		t.Errorf("id = %v, esperado 123", resultMap["id"])
	}

	if resultMap["name"] != "John Doe" {
		t.Errorf("name = %v, esperado John Doe", resultMap["name"])
	}
}

func TestRedisAdapter_GetData_NoArgs(t *testing.T) {
	adapter := NewRedisAdapter("localhost:6379", "", "user:{userId}", map[string]interface{}{})

	_, err := adapter.GetData([]AdapterAttribute{})
	if err == nil {
		t.Fatal("esperado erro quando não há argumentos")
	}

	expectedError := "the data key value was not informed"
	if err.Error() != expectedError {
		t.Errorf("erro = %v, esperado %v", err.Error(), expectedError)
	}
}

func TestRedisAdapter_GetData_RedisError(t *testing.T) {
	// Adapter com endereço inválido para forçar erro
	adapter := &redisAdapter{
		client: redis.NewClient(&redis.Options{
			Addr: "invalid:6379",
		}),
		keyPattern: "user:{userId}",
	}

	args := []AdapterAttribute{
		{Name: "userId", Type: "string", Value: "123"},
	}

	_, err := adapter.GetData(args)
	if err == nil {
		t.Fatal("esperado erro de conexão Redis")
	}
}

func TestRedisAdapter_GetData_InvalidJSON(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("erro ao iniciar miniredis: %v", err)
	}
	defer mr.Close()

	// Inserir JSON inválido
	mr.Set("user:123", "json inválido")

	adapter := &redisAdapter{
		client: redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		}),
		keyPattern: "user:{userId}",
	}

	args := []AdapterAttribute{
		{Name: "userId", Type: "string", Value: "123"},
	}

	_, err = adapter.GetData(args)
	if err == nil {
		t.Fatal("esperado erro de JSON inválido")
	}
}
