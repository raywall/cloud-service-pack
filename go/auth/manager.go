// Copyright 2025 Raywall Malheiros de Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package auth provides an auto managed token client, capable of monitoring token expiration
and refresh access token automatically without any intervention of the user.

It defines a type, [Handler], wich provides several methods (such as [Handler.GetToken], [Handler.Start]
and [Handler.Stop]) to enable the interaction with the token client.
*/
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
	"sync"
	"time"
)

// TokenRequest representa os dados enviados para a API de autenticação
type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// response representa a resposta da API de autenticação
type response struct {
	Active       bool   `json:"active"`
	ExpiresAt    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Token        string `json:"access_token"`
	TokenType    string `json:"token_type"`
}

// ManagedToken gerencia o ciclo de vida do token STS
type ManagedToken struct {
	apiURL      string
	accessToken *string
	cancelFunc  context.CancelFunc
	client      *http.Client
	ctx         context.Context
	expiresAt   time.Time
	mutex       sync.RWMutex
	refreshing  bool
	request     TokenRequest
}

func New(apiURL string, req TokenRequest, insecureSkipVerify bool, accessToken *string) Handler {
	ctx, cancel := context.WithCancel(context.Background())
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify, // WARNING: Use with caution in production
		},
	}
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	if insecureSkipVerify {
		httpClient.Transport = transport
	}

	return &ManagedToken{
		apiURL:      apiURL,
		accessToken: accessToken,
		cancelFunc:  cancel,
		client:      httpClient,
		ctx:         ctx,
		request:     req,
	}
}

// Start inicio o gerenciador de token e faz a primeira requisição para obter o token
// Retorna erro se não conseguir obter o token inicial
func (tm *ManagedToken) Start() error {
	// Obter o token inicial
	if err := tm.refreshToken(); err != nil {
		return err
	}

	// Iniciar goroutine para atualização automática
	go tm.refreshLoop()

	return nil
}

// GetToken retorna o token atual, garantindo que seja válido
func (tm *ManagedToken) GetToken() (string, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// Verifica se temos um token e se ele ainda é válido
	if tm.accessToken == nil {
		return "", errors.New("token não disponível")
	}

	// Verifica se o token está próximo de expirar (dentro de 30 segundos)
	if time.Now().Add(30 * time.Second).After(tm.expiresAt) {
		// Token está quase expirando, mas o refreshLoop deve estar tratando isso
		// Retornamos o token atual, que ainda é válido por pelo menos alguns segundos
		log.Println("Token está próximo de expiração, mas ainda é válido")
	}

	return *tm.accessToken, nil
}

// Stop interrompe o gerenciador de token
func (tm *ManagedToken) Stop() {
	tm.refreshing = false
	tm.cancelFunc()
}

// refreshLoop executa em background para manter o token atualizado
func (tm *ManagedToken) refreshLoop() {
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
			tm.refreshToken()
			time.Sleep(5 * time.Second) // Evita loop infinito em caso de falha
			continue
		}

		totalDuration := expiresAt.Sub(now)
		// Renovar quando faltar 20% do tempo para expirar
		refreshTime := expiresAt.Add(-totalDuration / 5)
		sleepDuration := refreshTime.Sub(now)

		if sleepDuration <= 0 {
			// Já passou o tempo de renovação, renovar imediatamente
			if err := tm.refreshToken(); err != nil {
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
				if err := tm.refreshToken(); err != nil {
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
func (tm *ManagedToken) refreshToken() error {
	// Preparar o payload da requisição
	payload := url.Values{}
	payload.Add("grant_type", "client_credentials")
	payload.Add("client_id", tm.request.ClientID)
	payload.Add("client_secret", tm.request.ClientSecret)

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
	var tokenResp response
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	// Atualizar o token com lock para thread safety
	tm.mutex.Lock()
	*tm.accessToken = tokenResp.Token

	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresAt) * time.Second)
	tm.mutex.Unlock()

	return nil
}
