package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raywall/cloud-service-pack/go/graphql/types"
)

// Testes para RestAdapter
func TestNewRestAdapter(t *testing.T) {
	cfg := &types.Config{
		AccessToken: "test-token",
	}

	attributes := map[string]interface{}{
		"userId": "string",
	}

	headers := map[string]interface{}{
		"Content-Type": "application/json",
	}

	adapter := NewRestAdapter(cfg, "https://api.example.com", "users/{userId}", true, attributes, headers)

	if adapter == nil {
		t.Fatal("NewRestAdapter retornou nil")
	}

	restAdapterImpl, ok := adapter.(*restAdapter)
	if !ok {
		t.Fatal("adapter não é do tipo *restAdapter")
	}

	if restAdapterImpl.baseUrl != "https://api.example.com" {
		t.Errorf("baseUrl = %v, esperado %v", restAdapterImpl.baseUrl, "https://api.example.com")
	}

	if restAdapterImpl.endpoint != "users/{userId}" {
		t.Errorf("endpoint = %v, esperado %v", restAdapterImpl.endpoint, "users/{userId}")
	}

	if !restAdapterImpl.auth {
		t.Error("auth deveria ser true")
	}
}

func TestRestAdapter_GetParameters(t *testing.T) {
	cfg := &types.Config{AccessToken: "test-token"}
	attributes := map[string]interface{}{
		"userId": "string",
		"limit":  "int",
	}

	adapter := NewRestAdapter(cfg, "https://api.example.com", "users", false, attributes, nil)

	args := map[string]interface{}{
		"userId": "123",
		"limit":  10,
	}

	result, err := adapter.GetParameters(args)
	if err != nil {
		t.Fatalf("GetParameters() erro = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("len(result) = %d, esperado 2", len(result))
	}
}

func TestRestAdapter_GetData_Success(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/123" {
			t.Errorf("path = %v, esperado /users/123", r.URL.Path)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization header = %v, esperado Bearer test-token", r.Header.Get("Authorization"))
		}

		if r.Header.Get("X-Custom") != "value-123" {
			t.Errorf("X-Custom header = %v, esperado value-123", r.Header.Get("X-Custom"))
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "123",
				"name": "John Doe",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &types.Config{AccessToken: "test-token"}
	headers := map[string]interface{}{
		"X-Custom": "value-{userId}",
	}

	adapter := NewRestAdapter(cfg, server.URL, "users/{userId}", true, nil, headers)

	args := []AdapterAttribute{
		{Name: "userId", Type: "string", Value: "123"},
	}

	result, err := adapter.GetData(args)
	if err != nil {
		t.Fatalf("GetData() erro = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("resultado não é map[string]interface{}")
	}

	if resultMap["id"] != "123" {
		t.Errorf("id = %v, esperado 123", resultMap["id"])
	}
}

func TestRestAdapter_GetData_WithoutAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("Authorization header não deveria estar presente")
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"public": "data",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &types.Config{AccessToken: "test-token"}
	adapter := NewRestAdapter(cfg, server.URL, "public", false, nil, nil)

	result, err := adapter.GetData([]AdapterAttribute{})
	if err != nil {
		t.Fatalf("GetData() erro = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("resultado não é map[string]interface{}")
	}

	if resultMap["public"] != "data" {
		t.Errorf("public = %v, esperado data", resultMap["public"])
	}
}

func TestRestAdapter_GetData_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &types.Config{AccessToken: "test-token"}
	adapter := NewRestAdapter(cfg, server.URL, "users/123", false, nil, nil)

	_, err := adapter.GetData([]AdapterAttribute{})
	if err == nil {
		t.Fatal("esperado erro HTTP 404")
	}

	expectedError := fmt.Sprintf("REST API returned status 404 for %s/users/123", server.URL)
	if err.Error() != expectedError {
		t.Errorf("erro = %v, esperado %v", err.Error(), expectedError)
	}
}

func TestRestAdapter_GetData_NetworkError(t *testing.T) {
	cfg := &types.Config{AccessToken: "test-token"}
	adapter := NewRestAdapter(cfg, "http://invalid-url", "users", false, nil, nil)

	_, err := adapter.GetData([]AdapterAttribute{})
	if err == nil {
		t.Fatal("esperado erro de rede")
	}
}

func TestRestAdapter_GetData_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("json inválido"))
	}))
	defer server.Close()

	cfg := &types.Config{AccessToken: "test-token"}
	adapter := NewRestAdapter(cfg, server.URL, "users", false, nil, nil)

	_, err := adapter.GetData([]AdapterAttribute{})
	if err == nil {
		t.Fatal("esperado erro de JSON inválido")
	}
}

func TestRestAdapter_GetData_NoParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static" {
			t.Errorf("path = %v, esperado /static", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": "static content",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &types.Config{AccessToken: "test-token"}
	adapter := NewRestAdapter(cfg, server.URL, "static", false, nil, nil)

	result, err := adapter.GetData([]AdapterAttribute{})
	if err != nil {
		t.Fatalf("GetData() erro = %v", err)
	}

	if result != "static content" {
		t.Errorf("result = %v, esperado static content", result)
	}
}
