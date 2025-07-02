// authenticator_test.go
package authenticator_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/raywall/cloud-service-pack/go/authenticator"
	"github.com/raywall/cloud-service-pack/go/authenticator/handlers"
)

// mockAuthHandler is a mock implementation of the handlers.AuthHandler interface
// for isolated testing of the Authenticator.
type mockAuthHandler struct {
	// The principal to be returned by Authenticate
	principalToReturn *handlers.Principal
	// The error to be returned by Authenticate
	errorToReturn error
	// A flag to check if the Authenticate method was called
	authenticateCalled bool
}

// Authenticate mocks the authentication process.
func (m *mockAuthHandler) Authenticate(ctx context.Context) (*handlers.Principal, error) {
	m.authenticateCalled = true
	return m.principalToReturn, m.errorToReturn
}

// TestAuthenticator_Authenticate tests the direct delegation to the handler.
func TestAuthenticator_Authenticate(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		// Setup mock handler to return a valid principal
		expectedPrincipal := &handlers.Principal{ID: "mock-user"}
		mockHandler := &mockAuthHandler{principalToReturn: expectedPrincipal}

		auth := authenticator.NewAuthenticator(mockHandler)
		p, err := auth.Authenticate(context.Background())

		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if !mockHandler.authenticateCalled {
			t.Error("Expected Authenticate method on the handler to be called, but it wasn't")
		}
		if p.ID != "mock-user" {
			t.Errorf("Expected principal ID 'mock-user', but got '%s'", p.ID)
		}
	})

	t.Run("failed authentication", func(t *testing.T) {
		// Setup mock handler to return an error
		expectedError := errors.New("authentication failed")
		mockHandler := &mockAuthHandler{errorToReturn: expectedError}

		auth := authenticator.NewAuthenticator(mockHandler)
		_, err := auth.Authenticate(context.Background())

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !errors.Is(err, expectedError) {
			t.Errorf("Expected error '%v', but got '%v'", expectedError, err)
		}
	})

	t.Run("no handler configured", func(t *testing.T) {
		// Manually create an authenticator without a handler
		auth := &authenticator.Authenticator{}
		_, err := auth.Authenticate(context.Background())

		if err == nil {
			t.Fatal("Expected an error for nil handler, but got nil")
		}
	})
}

// TestAuthenticator_WithController tests the creation and delegation to the TokenController.
func TestAuthenticator_WithController(t *testing.T) {
	// This test uses a real STSAuthHandler and TokenController, backed by a mock HTTP server.
	t.Run("success with STSAuthHandler", func(t *testing.T) {
		// Setup a mock STS server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"access_token": "controlled-token-123", "expires_in": 3600}`)
		}))
		defer mockServer.Close()

		// Create a real STS handler pointing to our mock server
		stsHandler := handlers.NewSTSAuthHandler(mockServer.URL, &handlers.Credential{}, nil)
		var token string

		// Create the authenticator and wire up the controller
		auth := authenticator.NewAuthenticator(stsHandler).WithController(&token)

		// Defer Stop() to ensure the controller's goroutine is cleaned up
		defer auth.Stop()

		// Test Start() delegation
		if err := auth.Start(); err != nil {
			t.Fatalf("auth.Start() failed: %v", err)
		}

		// Allow time for the initial token fetch
		time.Sleep(50 * time.Millisecond)

		// Test GetToken() delegation
		retrievedToken, err := auth.GetToken()
		if err != nil {
			t.Fatalf("auth.GetToken() failed: %v", err)
		}
		if retrievedToken != "controlled-token-123" {
			t.Errorf("Expected token 'controlled-token-123', got '%s'", retrievedToken)
		}
		if token != "controlled-token-123" {
			t.Errorf("Expected output variable to be 'controlled-token-123', got '%s'", token)
		}
	})

	// This test verifies the behavior when the handler is not of the required type.
	t.Run("failure with incorrect handler type", func(t *testing.T) {
		// Use a mock handler that is not an *STSAuthHandler
		mockHandler := &mockAuthHandler{}
		var token string

		auth := authenticator.NewAuthenticator(mockHandler).WithController(&token)

		// We expect a panic because the internal controller is nil.
		// We test each controller-delegated method for this panic.
		testPanic := func(methodName string, action func()) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic when calling %s with a nil controller", methodName)
				}
			}()
			action()
		}

		testPanic("Start", func() { auth.Start() })
		testPanic("GetToken", func() { auth.GetToken() })
		testPanic("Stop", func() { auth.Stop() })
	})
}
