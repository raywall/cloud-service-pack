package authenticator

import (
	"context"
	"fmt"

	"github.com/raywall/cloud-service-pack/go/authenticator/controllers"
	"github.com/raywall/cloud-service-pack/go/authenticator/handlers"
)

// Authenticator is the input point for the authentication logic.
// It delegates the validation to the configured [AuthHandler].
type Authenticator struct {
	handler    handlers.AuthHandler
	controller controllers.TokenController
}

// NewAuthenticator creates a new instance of the authenticator with
// the supplied handler.
func NewAuthenticator(handler handlers.AuthHandler) *Authenticator {
	return &Authenticator{
		handler: handler,
	}
}

// WithController creates a [TokenController], which will keep the token
// always updated and valid.
func (a *Authenticator) WithController(output controllers.OutputAccessToken) *Authenticator {
	if handler, ok := a.handler.(*handlers.STSAuthHandler); ok {
		a.controller = controllers.NewTokenController(handler, output)
	}
	return a
}

// Authenticate performs the authentication process using the configured handler.
func (a *Authenticator) Authenticate(ctx context.Context) (*handlers.Principal, error) {
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
