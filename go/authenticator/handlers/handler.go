// handler.go
// Copyright 2025 Raywall Malheiros de Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package handlers defines a generic interface for authentication mechanisms.

This package provides the core contracts (interfaces and data structures)
for handling different authentication strategies. The main component is the
AuthHandler interface, which must be implemented by any concrete
authentication method, such as STS (Security Token Service), LDAP, or
any other identity provider.
*/
package handlers

import (
	"context"
	"time"
)

// Principal represents the authenticated identity.
// It may contain information such as user ID, Scopes, Roles, Token, etc.
type Principal struct {
	ID          string                 `json:"id"`
	Roles       []string               `json:"roles"`
	Scopes      []string               `json:"scopes"`
	Extra       map[string]interface{} `json:"extra"`
	AccessToken *string                `json:"accessToken"`
	ExpiresAt   time.Time              `json:"expiresAt"`
}

// Credential represents the authentication keys that will be used to
// perform the system, application or user authentication.
type Credential struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// Options contains the optional attributes that can be used to customize
// the authentication process.
type Options struct {
	InsecureSkipVerify bool `json:"insecureSkipVerify"`
}

// AuthHandler is the interface that defines the contract for an
// authentication mechanism. Each implementation is responsible for
// validating a specific type of credential.
type AuthHandler interface {
	// Authenticate validates the credential and returns the [Principal] (identity)
	// in case of success, or an error otherwise.
	Authenticate(ctx context.Context) (*Principal, error)
}
