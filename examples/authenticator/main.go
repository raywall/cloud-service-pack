package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/raywall/cloud-service-pack/go/authenticator"
	"github.com/raywall/cloud-service-pack/go/authenticator/controllers"
	"github.com/raywall/cloud-service-pack/go/authenticator/handlers"
)

// stsState stores the state of our fake STS server.
type stsState struct {
	mutex            sync.Mutex
	tokenCounter     int32
	lastIssuedToken  string
	expectedClientID string
	expectedSecret   string
}

// createMockSTSServer creates a fake STS server that issues short-lived tokens.
func createMockSTSServer(state *stsState) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the form data from the request
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Validate the client credentials
		if r.FormValue("client_id") != state.expectedClientID || r.FormValue("client_secret") != state.expectedSecret {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Generate a new token that changes with each call
		newTokenNumber := atomic.AddInt32(&state.tokenCounter, 1)
		tokenValue := fmt.Sprintf("fake-token-%d", newTokenNumber)

		// Store the last issued token for validation in the protected API
		state.mutex.Lock()
		state.lastIssuedToken = tokenValue
		state.mutex.Unlock()

		log.Printf("[STS SERVER] Emitindo novo token: %s", tokenValue)

		// Respond with the new token and a very short expiration time (3 seconds)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": tokenValue,
			"expires_in":   3, // <-- Token expires in 3 seconds!
			"token_type":   "Bearer",
		})
	}))
}

// createMockProtectedAPIServer creates a fake API server that requires a valid token.
func createMockProtectedAPIServer(sts *stsState) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract the token from the "Bearer <token>" header
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		receivedToken := parts[1]

		// Check if the received token is the last one issued by the STS
		sts.mutex.Lock()
		isValid := receivedToken == sts.lastIssuedToken
		sts.mutex.Unlock()

		if !isValid {
			log.Printf("[API SERVER] Acesso NEGADO para o token: %s", receivedToken)
			http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		log.Printf("[API SERVER] Acesso PERMITIDO para o token: %s", receivedToken)
		io.WriteString(w, fmt.Sprintf("Olá! Você acessou com o token %s.", receivedToken))
	}))
}

func main() {
	log.Println("--- Iniciando Exemplo do Authenticator ---")

	// --- 1. Simulation Environment Setup ---
	stsServerState := &stsState{
		expectedClientID: "my-app-id",
		expectedSecret:   "my-app-secret",
	}
	mockSTS := createMockSTSServer(stsServerState)
	defer mockSTS.Close()

	mockAPI := createMockProtectedAPIServer(stsServerState)
	defer mockAPI.Close()

	log.Printf("Mock STS Server rodando em: %s", mockSTS.URL)
	log.Printf("Mock Protected API rodando em: %s", mockAPI.URL)

	// --- 2. Authenticator Library Setup ---
	creds := &handlers.Credential{
		ClientID:     "my-app-id",
		ClientSecret: "my-app-secret",
	}
	stsHandler := handlers.NewSTSAuthHandler(mockSTS.URL, creds, nil)

	var accessToken string
	auth := authenticator.NewAuthenticator(stsHandler).
		WithController(controllers.OutputAccessToken(&accessToken))

	if err := auth.Start(); err != nil {
		log.Fatalf("Falha ao iniciar o authenticator: %v", err)
	}
	defer auth.Stop()

	// --- 3. Client Application Logic Execution ---
	log.Println("\n--- Iniciando chamadas para a API a cada segundo ---")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("\n--- Exemplo Concluído ---")
			return
		case <-ticker.C:
			// The application simply requests the token. The library ensures it is valid.
			token, err := auth.GetToken()
			if err != nil {
				log.Printf("[CLIENT] Erro ao obter token: %v", err)
				continue
			}

			log.Printf("[CLIENT] Tentando chamar a API com o token: %s", token)

			// Make the call to the protected API
			req, _ := http.NewRequest("GET", mockAPI.URL, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("[CLIENT] Erro ao chamar a API: %v", err)
				continue
			}

			body, _ := io.ReadAll(resp.Body)
			log.Printf("[CLIENT] Resposta da API: Status %d, Body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			resp.Body.Close()
		}
	}
}
