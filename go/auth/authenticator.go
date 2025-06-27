package auth

import (
	"context"
	"fmt"
)

// Authenticator is the input point for the authentication logic.
// It delegates the validation to the configured [AuthHandler].
type Authenticator struct {
	handler    AuthHandler
	controller TokenController
}

// NewAuthenticator creates a new instance of the authenticator with
// the supplied handler.
func NewAuthenticator(handler AuthHandler) *Authenticator {
	return &Authenticator{
		handler: handler,
	}
}

// WithController creates a [TokenController], which will keep the token
// always updated and valid.
func (a *Authenticator) WithController(output OutputAccessToken) *Authenticator {
	if handler, ok := a.handler.(*STSAuthHandler); ok {
		a.controller = handler.NewTokenController(output)
	}
	return a
}

// Authenticate performs the authentication process using the configured handler.
func (a *Authenticator) Authenticate(ctx context.Context) (*Principal, error) {
	if a.handler == nil {
		return nil, fmt.Errorf("auth handler was not configured")
	}
	return a.handler.Authenticate(ctx)
}

func (a *Authenticator) Start() error {
	if a.handler == nil {
		return fmt.Errorf("auth handler was not configured")
	}
	return a.controller.Start()
}

func (a *Authenticator) GetToken() (string, error) {
	if a.handler == nil {
		return "", fmt.Errorf("auth handler was not configured")
	}
	return a.controller.GetToken()
}

func (a *Authenticator) Stop() {
	if a.handler != nil {
		a.controller.Stop()
	}
}
