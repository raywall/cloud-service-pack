package auth

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// AuthRequest representa os dados enviados para a API de autenticação
type AuthRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// ManagedToken gerencia o ciclo de vida do token STS
type ManagedToken struct {
	apiURL      string
	authRequest AuthRequest
	client      *http.Client

	accessToken *string
	expiresAt   time.Time
	mutex       sync.RWMutex

	ctx        context.Context
	cancelFunc context.CancelFunc
	refreshing bool
}

// TokenResponse representa a resposta da API de autenticação
type TokenResponse struct {
	Token        string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Active       bool   `json:"active"`
}
