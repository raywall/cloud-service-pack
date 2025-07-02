package controllers

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/raywall/cloud-service-pack/go/authenticator/handlers"
)

type TokenController interface {
	Start() error
	GetToken() (string, error)
	Stop()
}

type OutputAccessToken *string

// SelfControlledToken manages the life cycle of a STS token.
type SelfControlledToken struct {
	ctx               context.Context
	cancelFunc        context.CancelFunc
	mutex             sync.RWMutex
	outputAccessToken OutputAccessToken
	refreshing        bool
	handler           *handlers.STSAuthHandler
	principal         *handlers.Principal
}

// NewTokenController creates a new self managed token.
func NewTokenController(handler *handlers.STSAuthHandler, output OutputAccessToken) TokenController {
	ctx, cancel := context.WithCancel(context.Background())

	return &SelfControlledToken{
		ctx:               ctx,
		cancelFunc:        cancel,
		outputAccessToken: output,
		handler:           handler,
	}
}

// Start starts the self controlled token service and makes the first request to get the token.
// Returns error if you can't get the initial token.
func (tm *SelfControlledToken) Start() error {
	// Gets the initial token
	if err := tm.refreshToken(); err != nil {
		return err
	}

	// Starts a goroutine to update token automatically
	go tm.refreshLoop()

	return nil
}

// GetToken returns the current token, ensuring that it is valid.
func (tm *SelfControlledToken) GetToken() (string, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// Check if we have a token and if it is still valid.
	if tm.outputAccessToken == nil {
		return "", errors.New("token not available")
	}

	// Check if the token is close to expire (within 30 seconds).
	if tm.principal != nil && time.Now().Add(30*time.Second).After(tm.principal.ExpiresAt) {
		// Token is almost expiring, but Refreshloop must be treating it.
		// We returned the current token, which is still valid for at least a few seconds.
		log.Println("Token is close to expiration, but is still valid")
	}

	return *tm.outputAccessToken, nil
}

// Stop interrupts the self controlled token service.
func (tm *SelfControlledToken) Stop() {
	tm.refreshing = false
	tm.cancelFunc()
}

// refreshLoop runs in background to keep the token always updated.
func (tm *SelfControlledToken) refreshLoop() {
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
		expiresAt := tm.principal.ExpiresAt
		tm.mutex.Unlock()

		// Calculate time until we need to renew the token.
		// We renew when it is 20% of the total expiration time.
		now := time.Now()
		if expiresAt.IsZero() {
			// If we do not have a valid expiration time, we immediately try.
			_ = tm.refreshToken()

			// Avoid infinite loop in case of failure.
			time.Sleep(5 * time.Second)
			continue
		}
		totalDuration := expiresAt.Sub(now)

		// Renew when 20% of the time is missing to expire.
		refreshTime := expiresAt.Add(-totalDuration / 5)
		sleepDuration := refreshTime.Sub(now)

		if sleepDuration <= 0 {
			// The renewal time has passed, renew immediately.
			if err := tm.refreshToken(); err != nil {
				log.Printf("Error renewing token: %v. Trying again in 10 seconds.", err)

				// In case of error, wait just before trying again.
				select {
				case <-tm.ctx.Done():
					return
				case <-time.After(10 * time.Second):
					continue
				}
			}
		} else {
			// Wait for renewal.
			select {
			case <-tm.ctx.Done():
				return
			case <-time.After(sleepDuration):
				if err := tm.refreshToken(); err != nil {
					log.Printf("Error renewing token: %v. Trying again in 10 seconds.", err)

					// In case of error, wait just before trying again.
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

// refreshToken gets a new token.
func (tm *SelfControlledToken) refreshToken() error {
	resp, err := tm.handler.Authenticate(tm.ctx)
	if err != nil {
		return err
	}

	// Update the token with lock for thread safety.
	tm.mutex.Lock()
	*tm.outputAccessToken = *resp.AccessToken
	tm.mutex.Unlock()

	return nil
}
