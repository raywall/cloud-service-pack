// controller_test.go
package controllers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/raywall/cloud-service-pack/go/authenticator/controllers"
	"github.com/raywall/cloud-service-pack/go/authenticator/handlers"
)

// mockStsServer it is a helper to simulate the STS server.
// It allows us to control the answers and synchronize with tests using a channel.
type mockStsServer struct {
	server       *httptest.Server
	requestCount int32
	requestChan  chan struct{} // Signals when a request is received
	tokenPrefix  string
	expiresIn    int // in seconds
}

func newMockStsServer(tokenPrefix string, expiresIn int) *mockStsServer {
	mock := &mockStsServer{
		requestChan: make(chan struct{}, 10), // Buffer not to block
		tokenPrefix: tokenPrefix,
		expiresIn:   expiresIn,
	}

	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increase the request accountant safely
		count := atomic.AddInt32(&mock.requestCount, 1)

		// Assembles token based on the request number
		tokenPayload := map[string]interface{}{
			"access_token": fmt.Sprintf("%s-%d", mock.tokenPrefix, count),
			"expires_in":   mock.expiresIn,
			"scope":        "test",
			"token_type":   "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenPayload)

		// Signals that a request has been processed
		mock.requestChan <- struct{}{}
	}))

	return mock
}

func (m *mockStsServer) URL() string {
	return m.server.URL
}

func (m *mockStsServer) Close() {
	m.server.Close()
	close(m.requestChan)
}

// waitForRequest waits for a request or failure for timeout.
func (m *mockStsServer) waitForRequest(t *testing.T, timeout time.Duration) {
	t.Helper()
	select {
	case <-m.requestChan:
		// Sucesso
	case <-time.After(timeout):
		t.Fatal("timed out waiting for the server to receive a request")
	}
}

// TestToKenController_startandGetToken checks the basic startup flow.
func TestTokenController_StartAndGetToken(t *testing.T) {
	mockServer := newMockStsServer("initial-token", 3600) // Long-term token
	defer mockServer.Close()

	creds := &handlers.Credential{ClientID: "test-client"}
	handler := handlers.NewSTSAuthHandler(mockServer.URL(), creds, nil)

	var token string
	controller := controllers.NewTokenController(handler, &token)
	defer controller.Stop()

	// Start the controller
	err := controller.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Waiting for the first request to be processed
	mockServer.waitForRequest(t, 2*time.Second)

	// Gets the token
	retrievedToken, err := controller.GetToken()
	if err != nil {
		t.Fatalf("GetToken() failed: %v", err)
	}

	if retrievedToken != "initial-token-1" {
		t.Errorf("expected token 'initial-token-1', got '%s'", retrievedToken)
	}

	if token != "initial-token-1" {
		t.Errorf("expected output variable to be 'initial-token-1', got '%s'", token)
	}
}

// TestTokenController_StartFailure checks error treatment in startup.
func TestTokenController_StartFailure(t *testing.T) {
	// Servidor que sempre falha
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	creds := &handlers.Credential{ClientID: "test-client"}
	handler := handlers.NewSTSAuthHandler(server.URL, creds, nil)
	var token string
	controller := controllers.NewTokenController(handler, &token)

	err := controller.Start()
	if err == nil {
		t.Fatal("expected Start() to fail, but it succeeded")
	}
}

// TestTokenController_AutomaticRefresh check if the token is renewed automatically.
func TestTokenController_AutomaticRefresh(t *testing.T) {
	// Token with very short expiration to force renewal
	mockServer := newMockStsServer("refresh-token", 1) // Expires in 1 second
	defer mockServer.Close()

	creds := &handlers.Credential{ClientID: "test-client"}
	handler := handlers.NewSTSAuthHandler(mockServer.URL(), creds, nil)

	var token string
	controller := controllers.NewTokenController(handler, &token)
	defer controller.Stop()

	err := controller.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// 1. Check the initial token
	mockServer.waitForRequest(t, time.Second)
	currentToken, _ := controller.GetToken()
	if currentToken != "refresh-token-1" {
		t.Fatalf("expected initial token 'refresh-token-1', got '%s'", currentToken)
	}

	// 2. Waiting for automatic renewal
	// Logic renews with 20% of the remaining time, so for 1s, it should be ~800ms.
	mockServer.waitForRequest(t, 2*time.Second)

	// 3. Check the renewed token
	refreshedToken, _ := controller.GetToken()
	if refreshedToken != "refresh-token-2" {
		t.Fatalf("expected refreshed token 'refresh-token-2', got '%s'", refreshedToken)
	}
}

// TestTokenController_Stop Checks if the renewal gouritine for.
func TestTokenController_Stop(t *testing.T) {
	mockServer := newMockStsServer("stoppable-token", 1) // Expires in 1 second
	defer mockServer.Close()

	creds := &handlers.Credential{}
	handler := handlers.NewSTSAuthHandler(mockServer.URL(), creds, nil)
	var token string
	controller := controllers.NewTokenController(handler, &token)

	controller.Start()
	mockServer.waitForRequest(t, time.Second)

	// Guarantees that the first token was received
	if atomic.LoadInt32(&mockServer.requestCount) != 1 {
		t.Fatal("initial token request was not made")
	}

	// For the controller
	controller.Stop()

	// Waits longer than the renewal interval to ensure
	// that no new request was made.
	time.Sleep(1500 * time.Millisecond)

	// The request counter should still be 1
	if atomic.LoadInt32(&mockServer.requestCount) != 1 {
		t.Errorf("expected 1 request, but got %d. Stop() did not work.", mockServer.requestCount)
	}
}

// TestTokenController_GetTokenBeforeStart Check the behavior before startup.
func TestTokenController_GetTokenBeforeStart(t *testing.T) {
	mockServer := newMockStsServer("never-used", 3600)
	defer mockServer.Close()

	creds := &handlers.Credential{}
	handler := handlers.NewSTSAuthHandler(mockServer.URL(), creds, nil)
	var token string
	controller := controllers.NewTokenController(handler, &token)

	// Try to get the token before calling Start ()
	_, err := controller.GetToken()
	if err == nil {
		t.Fatal("expected an error when calling GetToken() before Start(), but got nil")
	}
}
