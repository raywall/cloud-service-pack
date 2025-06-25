package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewManagedToken(t *testing.T) {
	tests := []struct {
		name               string
		apiURL             string
		authRequest        AuthRequest
		insecureSkipVerify bool
		accessToken        *string
	}{
		{
			name:   "criar token manager com configurações básicas",
			apiURL: "https://api.example.com/token",
			authRequest: AuthRequest{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			insecureSkipVerify: false,
			accessToken:        new(string),
		},
		{
			name:   "criar token manager com insecure skip verify",
			apiURL: "https://api.example.com/token",
			authRequest: AuthRequest{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			insecureSkipVerify: true,
			accessToken:        new(string),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := NewManagedToken(tt.apiURL, tt.authRequest, tt.insecureSkipVerify, tt.accessToken)

			managedToken, ok := token.(*ManagedToken)
			if !ok {
				t.Fatal("esperado *ManagedToken")
			}

			if managedToken.apiURL != tt.apiURL {
				t.Errorf("apiURL = %v, esperado %v", managedToken.apiURL, tt.apiURL)
			}

			if managedToken.authRequest.ClientID != tt.authRequest.ClientID {
				t.Errorf("ClientID = %v, esperado %v", managedToken.authRequest.ClientID, tt.authRequest.ClientID)
			}

			if managedToken.accessToken != tt.accessToken {
				t.Errorf("accessToken pointer diferente do esperado")
			}
		})
	}
}

func TestManagedToken_RefreshToken_Success(t *testing.T) {
	// Mock server que retorna um token válido
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("esperado POST, recebido %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type incorreto: %s", r.Header.Get("Content-Type"))
		}

		// Verificar o payload
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		if r.Form.Get("grant_type") != "client_credentials" {
			t.Errorf("grant_type incorreto: %s", r.Form.Get("grant_type"))
		}

		if r.Form.Get("client_id") != "test-client" {
			t.Errorf("client_id incorreto: %s", r.Form.Get("client_id"))
		}

		if r.Form.Get("client_secret") != "test-secret" {
			t.Errorf("client_secret incorreto: %s", r.Form.Get("client_secret"))
		}

		response := TokenResponse{
			Token:     "test-access-token",
			TokenType: "Bearer",
			ExpiresAt: 3600,
			Active:    true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	accessToken := ""
	token := NewManagedToken(server.URL, AuthRequest{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}, false, &accessToken)

	managedToken := token.(*ManagedToken)

	err := managedToken.RefreshToken()
	if err != nil {
		t.Fatalf("RefreshToken() erro = %v", err)
	}

	if accessToken != "test-access-token" {
		t.Errorf("accessToken = %v, esperado %v", accessToken, "test-access-token")
	}

	// Verificar se expiresAt foi definido corretamente
	expectedExpiry := time.Now().Add(3600 * time.Second)
	if managedToken.expiresAt.Before(expectedExpiry.Add(-time.Minute)) ||
		managedToken.expiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("expiresAt fora do intervalo esperado")
	}
}

func TestManagedToken_RefreshToken_HTTPError(t *testing.T) {
	// Mock server que retorna erro HTTP
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	accessToken := ""
	token := NewManagedToken(server.URL, AuthRequest{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}, false, &accessToken)

	managedToken := token.(*ManagedToken)

	err := managedToken.RefreshToken()
	if err == nil {
		t.Fatal("esperado erro, mas não houve")
	}

	expectedError := "falha na autenticação: status 401 Unauthorized"
	if err.Error() != expectedError {
		t.Errorf("erro = %v, esperado %v", err.Error(), expectedError)
	}
}

func TestManagedToken_RefreshToken_InvalidJSON(t *testing.T) {
	// Mock server que retorna JSON inválido
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("json inválido"))
	}))
	defer server.Close()

	accessToken := ""
	token := NewManagedToken(server.URL, AuthRequest{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}, false, &accessToken)

	managedToken := token.(*ManagedToken)

	err := managedToken.RefreshToken()
	if err == nil {
		t.Fatal("esperado erro de JSON inválido")
	}
}

func TestManagedToken_GetToken_Success(t *testing.T) {
	accessToken := "valid-token"
	token := NewManagedToken("https://api.example.com", AuthRequest{}, false, &accessToken)

	managedToken := token.(*ManagedToken)
	// Definir uma data de expiração futura
	managedToken.expiresAt = time.Now().Add(time.Hour)

	result, err := managedToken.GetToken()
	if err != nil {
		t.Fatalf("GetToken() erro = %v", err)
	}

	if result != "valid-token" {
		t.Errorf("GetToken() = %v, esperado %v", result, "valid-token")
	}
}

func TestManagedToken_GetToken_NoToken(t *testing.T) {
	token := NewManagedToken("https://api.example.com", AuthRequest{}, false, nil)
	managedToken := token.(*ManagedToken)

	_, err := managedToken.GetToken()
	if err == nil {
		t.Fatal("esperado erro quando não há token")
	}

	expectedError := "token não disponível"
	if err.Error() != expectedError {
		t.Errorf("erro = %v, esperado %v", err.Error(), expectedError)
	}
}

func TestManagedToken_GetToken_NearExpiry(t *testing.T) {
	accessToken := "expiring-soon-token"
	token := NewManagedToken("https://api.example.com", AuthRequest{}, false, &accessToken)

	managedToken := token.(*ManagedToken)
	// Definir uma data de expiração próxima (dentro de 30 segundos)
	managedToken.expiresAt = time.Now().Add(15 * time.Second)

	result, err := managedToken.GetToken()
	if err != nil {
		t.Fatalf("GetToken() erro = %v", err)
	}

	// Mesmo próximo da expiração, deve retornar o token
	if result != "expiring-soon-token" {
		t.Errorf("GetToken() = %v, esperado %v", result, "expiring-soon-token")
	}
}

func TestManagedToken_Start_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TokenResponse{
			Token:     "initial-token",
			TokenType: "Bearer",
			ExpiresAt: 3600,
			Active:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	accessToken := ""
	token := NewManagedToken(server.URL, AuthRequest{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}, false, &accessToken)

	err := token.Start()
	if err != nil {
		t.Fatalf("Start() erro = %v", err)
	}

	// Verificar se o token foi obtido
	if accessToken != "initial-token" {
		t.Errorf("accessToken = %v, esperado %v", accessToken, "initial-token")
	}

	// Cleanup
	token.Stop()
}

func TestManagedToken_Start_InitialRefreshFails(t *testing.T) {
	// Mock server que sempre falha
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	accessToken := ""
	token := NewManagedToken(server.URL, AuthRequest{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}, false, &accessToken)

	err := token.Start()
	if err == nil {
		t.Fatal("esperado erro ao iniciar com falha na obtenção inicial do token")
	}
}

func TestManagedToken_RefreshLoop_Concurrency(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	// Mock server que conta as chamadas
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		currentCount := callCount
		mu.Unlock()

		response := TokenResponse{
			Token:     "token-" + string(rune(currentCount)),
			TokenType: "Bearer",
			ExpiresAt: 1, // 1 segundo para forçar renovações frequentes
			Active:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	accessToken := ""
	token := NewManagedToken(server.URL, AuthRequest{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}, false, &accessToken)

	err := token.Start()
	if err != nil {
		t.Fatalf("Start() erro = %v", err)
	}

	// Aguardar algumas renovações automáticas
	time.Sleep(3 * time.Second)

	// Verificar se houve múltiplas chamadas
	mu.Lock()
	finalCallCount := callCount
	mu.Unlock()

	if finalCallCount < 2 {
		t.Errorf("esperado pelo menos 2 chamadas, recebido %d", finalCallCount)
	}

	token.Stop()
}

func TestManagedToken_Stop(t *testing.T) {
	accessToken := ""
	token := NewManagedToken("https://api.example.com", AuthRequest{}, false, &accessToken)

	managedToken := token.(*ManagedToken)

	// Verificar se o contexto não está cancelado inicialmente
	select {
	case <-managedToken.ctx.Done():
		t.Fatal("contexto cancelado prematuramente")
	default:
	}

	token.Stop()

	// Verificar se o contexto foi cancelado
	select {
	case <-managedToken.ctx.Done():
		// OK, contexto foi cancelado
	case <-time.After(100 * time.Millisecond):
		t.Fatal("contexto não foi cancelado após Stop()")
	}
}

func TestManagedToken_RefreshLoop_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TokenResponse{
			Token:     "test-token",
			TokenType: "Bearer",
			ExpiresAt: 3600,
			Active:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	accessToken := ""
	token := NewManagedToken(server.URL, AuthRequest{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}, false, &accessToken)

	managedToken := token.(*ManagedToken)

	// Iniciar o refresh loop
	go managedToken.RefreshLoop()

	// Aguardar um pouco para garantir que o loop começou
	time.Sleep(100 * time.Millisecond)

	// Cancelar o contexto
	token.Stop()

	// Aguardar um pouco para garantir que o loop terminou
	time.Sleep(100 * time.Millisecond)

	// Verificar se refreshing foi definido como false
	managedToken.mutex.RLock()
	refreshing := managedToken.refreshing
	managedToken.mutex.RUnlock()

	if refreshing {
		t.Error("refreshing ainda é true após cancelamento do contexto")
	}
}

func TestManagedToken_RefreshLoop_ErrorHandling(t *testing.T) {
	failCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failCount++
		if failCount <= 2 {
			// Falhar nas primeiras duas tentativas
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Sucesso na terceira tentativa
		response := TokenResponse{
			Token:     "success-token",
			TokenType: "Bearer",
			ExpiresAt: 3600,
			Active:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	accessToken := ""
	token := NewManagedToken(server.URL, AuthRequest{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}, false, &accessToken)

	managedToken := token.(*ManagedToken)

	// Definir expiresAt no passado para forçar renovação imediata
	managedToken.expiresAt = time.Now().Add(-time.Hour)

	// Executar o refresh loop por um tempo limitado
	done := make(chan bool)
	go func() {
		managedToken.RefreshLoop()
		done <- true
	}()

	// Aguardar o suficiente para várias tentativas
	time.Sleep(25 * time.Second)

	// Parar o loop
	token.Stop()

	// Aguardar o loop terminar
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RefreshLoop não terminou após Stop()")
	}

	// Verificar se eventualmente obteve sucesso
	if accessToken != "success-token" {
		t.Errorf("accessToken = %v, esperado %v após recuperação de erros", accessToken, "success-token")
	}
}

// Benchmark para medir performance do GetToken
func BenchmarkManagedToken_GetToken(b *testing.B) {
	accessToken := "benchmark-token"
	token := NewManagedToken("https://api.example.com", AuthRequest{}, false, &accessToken)

	managedToken := token.(*ManagedToken)
	managedToken.expiresAt = time.Now().Add(time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := managedToken.GetToken()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Test para verificar thread safety
func TestManagedToken_ThreadSafety(t *testing.T) {
	accessToken := "thread-safe-token"
	token := NewManagedToken("https://api.example.com", AuthRequest{}, false, &accessToken)

	managedToken := token.(*ManagedToken)
	managedToken.expiresAt = time.Now().Add(time.Hour)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Múltiplas goroutines chamando GetToken simultaneamente
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, err := managedToken.GetToken()
				if err != nil {
					errors <- err
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Verificar se houve erros
	for err := range errors {
		t.Errorf("erro em acesso concorrente: %v", err)
	}
}
