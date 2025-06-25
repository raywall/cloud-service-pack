package adapters

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/raywall/cloud-service-pack/go/graphql/types"
)

// Mock para types.Config
type mockConfig struct {
	AccessToken string
}

func (m *mockConfig) GetAccessToken() string {
	return m.AccessToken
}

// Testes para a função getParameters
func TestGetParameters(t *testing.T) {
	tests := []struct {
		name        string
		attributes  map[string]interface{}
		args        map[string]interface{}
		expected    []AdapterAttribute
		expectError bool
	}{
		{
			name: "parâmetros básicos com valores",
			attributes: map[string]interface{}{
				"userId": "string",
				"limit":  "int",
			},
			args: map[string]interface{}{
				"userId": "123",
				"limit":  10,
			},
			expected: []AdapterAttribute{
				{Name: "userId", Type: "string", Value: "123"},
				{Name: "limit", Type: "int", Value: 10},
			},
			expectError: false,
		},
		{
			name: "parâmetros sem valores correspondentes",
			attributes: map[string]interface{}{
				"userId": "string",
				"status": "string",
			},
			args: map[string]interface{}{
				"userId": "123",
			},
			expected: []AdapterAttribute{
				{Name: "userId", Type: "string", Value: "123"},
				{Name: "status", Type: "string", Value: nil},
			},
			expectError: false,
		},
		{
			name:        "atributos vazios",
			attributes:  map[string]interface{}{},
			args:        map[string]interface{}{},
			expected:    []AdapterAttribute{},
			expectError: false,
		},
		{
			name: "args extras não utilizados",
			attributes: map[string]interface{}{
				"userId": "string",
			},
			args: map[string]interface{}{
				"userId":  "123",
				"ignored": "value",
			},
			expected: []AdapterAttribute{
				{Name: "userId", Type: "string", Value: "123"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getParameters(tt.attributes, tt.args)

			if tt.expectError && err == nil {
				t.Error("esperado erro, mas não ocorreu")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("erro inesperado: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("tamanho resultado = %d, esperado %d", len(result), len(tt.expected))
				return
			}

			// Verificar se todos os elementos esperados estão presentes
			for _, expected := range tt.expected {
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
					t.Errorf("atributo esperado não encontrado: %+v", expected)
				}
			}
		})
	}
}

// Teste de integração para verificar a interface
func TestAdapterInterface(t *testing.T) {
	// Verificar se RedisAdapter implementa Adapter
	var _ Adapter = NewRedisAdapter("localhost:6379", "", "key", nil)

	// Verificar se RestAdapter implementa Adapter
	cfg := &types.Config{AccessToken: "token"}
	var _ Adapter = NewRestAdapter(cfg, "url", "endpoint", false, nil, nil)
}

// Benchmark para getParameters
func BenchmarkGetParameters(b *testing.B) {
	attributes := map[string]interface{}{
		"userId":   "string",
		"status":   "string",
		"limit":    "int",
		"offset":   "int",
		"category": "string",
	}

	args := map[string]interface{}{
		"userId":   "123",
		"status":   "active",
		"limit":    100,
		"offset":   0,
		"category": "premium",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := getParameters(attributes, args)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Teste para verificar regex global
func TestRegexPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"{userId}", true},
		{"{status}", true},
		{"users/{userId}/posts", true},
		{"static-endpoint", false},
		{"{}", false},
		{"users/{userId}/posts/{postId}", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := re.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("regex match = %v, esperado %v para input %v", result, tt.expected, tt.input)
			}
		})
	}
}

// Teste para timeouts do RestAdapter
func TestRestAdapter_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simular resposta lenta
		time.Sleep(15 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &types.Config{AccessToken: "test-token"}
	adapter := NewRestAdapter(cfg, server.URL, "slow", false, nil, nil)

	_, err := adapter.GetData([]AdapterAttribute{})
	if err == nil {
		t.Fatal("esperado erro de timeout")
	}
}
