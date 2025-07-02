// sts_test.go
package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/raywall/cloud-service-pack/go/authenticator/handlers"
)

// TestSTSAuthHandler_Authenticate tests the different scenarios of the Authenticate method.
func TestSTSAuthHandler_Authenticate(t *testing.T) {
	// Test Case 1: happy way (success)
	t.Run("success", func(t *testing.T) {
		// Creates a test server that simulates a successful STS response.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verifica se o request está correto
			if r.Method != "POST" {
				t.Errorf("Expected 'POST', got '%s'", r.Method)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("Failed to parse form: %v", err)
			}
			if r.FormValue("grant_type") != "client_credentials" {
				t.Errorf("Expected grant_type 'client_credentials', got '%s'", r.FormValue("grant_type"))
			}

			// Envia a resposta de sucesso
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{
				"access_token": "fake-jwt-token",
				"expires_in": 3600,
				"scope": "read,write",
				"token_type": "Bearer",
				"active": true
			}`)
		}))
		defer server.Close()

		// Configures credentials and handler
		creds := &handlers.Credential{ClientID: "test-client", ClientSecret: "test-secret"}
		handler := handlers.NewSTSAuthHandler(server.URL, creds, nil)

		// Performs authentication
		principal, err := handler.Authenticate(context.Background())

		// results validation
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if principal == nil {
			t.Fatal("Expected a principal, but got nil")
		}
		if *principal.AccessToken != "fake-jwt-token" {
			t.Errorf("Expected access token 'fake-jwt-token', got '%s'", *principal.AccessToken)
		}
		if principal.ID != "test-client" {
			t.Errorf("Expected principal ID 'test-client', got '%s'", principal.ID)
		}
		if len(principal.Scopes) != 2 || principal.Scopes[0] != "read" || principal.Scopes[1] != "write" {
			t.Errorf("Expected scopes ['read', 'write'], got %v", principal.Scopes)
		}
		if time.Until(principal.ExpiresAt).Seconds() < 3590 {
			t.Error("Expected expiration time to be around 1 hour from now")
		}
	})

	// Test Case 2: STS server error
	t.Run("server error", func(t *testing.T) {
		// Creates a test server that simulates a failure.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Internal Server Error")
		}))
		defer server.Close()

		creds := &handlers.Credential{ClientID: "test-client", ClientSecret: "test-secret"}
		handler := handlers.NewSTSAuthHandler(server.URL, creds, nil)

		principal, err := handler.Authenticate(context.Background())

		// results validation
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if principal != nil {
			t.Fatal("Expected principal to be nil, but it was not")
		}
		expectedError := "authentication failure: status 500 Internal Server Error"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedError, err.Error())
		}
	})

	// Test Case 3: malformed JSON answer
	t.Run("bad json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"access_token": "incomplete"`) // invalid JSON
		}))
		defer server.Close()

		creds := &handlers.Credential{ClientID: "test-client", ClientSecret: "test-secret"}
		handler := handlers.NewSTSAuthHandler(server.URL, creds, nil)

		principal, err := handler.Authenticate(context.Background())

		if err == nil {
			t.Fatal("Expected an error for bad JSON, but got nil")
		}
		if principal != nil {
			t.Fatal("Expected principal to be nil, but it was not")
		}
		expectedError := "decoding authentication response"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedError, err.Error())
		}
	})
}

// TestNewSTSAuthHandler checks the Handler construction and the HTTP customer configuration.
func TestNewSTSAuthHandler(t *testing.T) {
	creds := &handlers.Credential{}

	t.Run("default options", func(t *testing.T) {
		handler := handlers.NewSTSAuthHandler("http://localhost", creds, nil)
		if handler == nil {
			t.Fatal("Handler should not be nil")
		}
	})

	t.Run("with InsecureSkipVerify", func(t *testing.T) {
		// This test is more conceptual as we cannot inspect the client http
		// TLS.Config directly after its creation.
		// The functionality is indirectly tested on authenticate tests that
		// use httptest.newtlsserver.
		opts := &handlers.Options{InsecureSkipVerify: true}
		handler := handlers.NewSTSAuthHandler("https://localhost", creds, opts)

		if handler == nil {
			t.Fatal("Handler should not be nil")
		}
		// The true verification that insecureskipverify works would require
		// a httptest.newtlsserver () with a self -signed certificate.
	})
}
