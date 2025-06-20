package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewManagedToken(apiURL string, authRequest AuthRequest, certSkipVerify bool) *TokenManager {
	ctx, cancel := context.WithCancel(context.Background())
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // WARNING: Use with caution in production
		},
	}
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	if certSkipVerify {
		httpClient.Transport = transport
	}

	return &TokenManager{
		apiURL:      apiURL,
		authRequest: authRequest,
		client:      httpClient,
		ctx:         ctx,
		cancelFunc:  cancel,
	}
}

// Start inicio o gerenciador de token e faz a primeira requisição para obter o token
// Retorna erro se não conseguir obter o token inicial
func (tm *TokenManager) Start() error {
	// Obter o token inicial
	if err := tm.RefreshToken(); err != nil {
		return err
	}

	// Iniciar goroutine para atualização automática
	go tm.RefreshLoop()

	return nil
}

// GetToken retorna o token atual, garantindo que seja válido
func (tm *TokenManager) GetToken() (string, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// Verifica se temos um token e se ele ainda é válido
	if tm.token == "" {
		return "", errors.New("token não disponível")
	}

	// Verifica se o token está próximo de expirar (dentro de 30 segundos)
	if time.Now().Add(30 * time.Second).After(tm.expiresAt) {
		// Token está quase expirando, mas o refreshLoop deve estar tratando isso
		// Retornamos o token atual, que ainda é válido por pelo menos alguns segundos
		log.Println("Token está próximo de expiração, mas ainda é válido")
	}

	return tm.token, nil
}

// Stop interrompe o gerenciador de token
func (tm *TokenManager) Stop() {
	tm.cancelFunc()
}

// refreshLoop executa em background para manter o token atualizado
func (tm *TokenManager) RefreshLoop() {
	tm.mutex.Lock()
	tm.refreshing = true
	tm.mutex.Unlock()

	defer func() {
		tm.mutex.Lock()
		tm.refreshing = false
		tm.mutex.Unlock()
	}()

	for {
		tm.mutex.Lock()
		expiresAt := tm.expiresAt
		tm.mutex.Unlock()

		// Calcular o tempo até precisarmos renovar o token
		// Renovamos quando estiver a 20% do tempo total de expiração
		now := time.Now()
		if expiresAt.IsZero() {
			// Se não temos um tempo de expiração válido, tentamos imediatamente
			tm.RefreshToken()
			time.Sleep(5 * time.Second) // Evita loop infinito em caso de falha
			continue
		}

		totalDuration := expiresAt.Sub(now)
		// Renovar quando faltar 20% do tempo para expirar
		refreshTime := expiresAt.Add(-totalDuration / 5)
		sleepDuration := refreshTime.Sub(now)

		if sleepDuration <= 0 {
			// Já passou o tempo de renovação, renovar imediatamente
			if err := tm.RefreshToken(); err != nil {
				log.Printf("Erro ao renovar token: %v. Tentando novamente em 10 segundos.", err)
				// Em caso de erro, esperar um pouco antes de tentar novamente
				select {
				case <-tm.ctx.Done():
					return
				case <-time.After(10 * time.Second):
					continue
				}
			}
		} else {
			// Esperar até o momento calculado para renovação
			select {
			case <-tm.ctx.Done():
				return
			case <-time.After(sleepDuration):
				if err := tm.RefreshToken(); err != nil {
					log.Printf("Erro ao renovar token: %v. Tentando novamente em 10 segundos.", err)
					// Em caso de erro, esperar um pouco antes de tentar novamente
					select {
					case <-tm.ctx.Done():
						return
					case <-time.After(10 * time.Second):
						continue
					}
				}
			}
		}
	}
}

// refreshToken faz uma chamada à API para obter um novo token
func (tm *TokenManager) RefreshToken() error {
	// Preparar o payload da requisição
	payload := url.Values{}
	payload.Add("grant_type", "client_credentials")
	payload.Add("client_id", tm.authRequest.ClientID)
	payload.Add("client_secret", tm.authRequest.ClientSecret)

	// Criar a requisição
	req, err := http.NewRequestWithContext(tm.ctx, "POST", tm.apiURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Fazer a requisição
	resp, err := tm.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verificar o status da resposta
	if resp.StatusCode != http.StatusOK {
		return errors.New("falha na autenticação: status " + resp.Status)
	}

	// Decodificar a resposta
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	// Atualizar o token com lock para thread safety
	tm.mutex.Lock()
	tm.token = tokenResp.Token
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresAt) * time.Second)
	tm.mutex.Unlock()

	return nil
}
