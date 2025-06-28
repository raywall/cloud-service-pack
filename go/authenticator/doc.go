// Copyright 2025 Raywall Malheiros de Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package authenticator provides a comprehensive and extensible framework for client
authentication and automated token lifecycle management in Go applications.

It decouples the authentication process into distinct, manageable components,
making it easy to support different authentication strategies and to handle
expiring tokens for long-running services.

# Core Concepts

The package is built around three main components that work together:

1. Authenticator:
The `Authenticator` is the primary entry point and facade of the package.
It orchestrates the authentication flow by delegating tasks to a configured
authentication handler and an optional token controller.

2. AuthHandler:
Defined in the `handlers` sub-package, the `handlers.AuthHandler` is an
interface that represents a specific authentication strategy. Implementations
are responsible for validating a set of credentials and returning a `Principal`
(a representation of the authenticated identity) upon success.
A concrete implementation, `handlers.STSAuthHandler`, is provided for
authenticating against a Security Token Service using the OAuth 2.0
Client Credentials flow.

3. TokenController:
Defined in the `controllers` sub-package, the `controllers.TokenController`
is an optional but powerful component that manages the lifecycle of a token.
When configured using `WithController`, it automatically refreshes the token
in the background before it expires. This is essential for long-running
applications that need to maintain a valid authentication token at all times.

# Usage

The following example demonstrates how to create an authenticator that uses an
STS handler to fetch a token and a controller to keep it automatically refreshed.

	package main

	import (
		"fmt"
		"log"
		"time"

		"github.com/raywall/cloud-service-pack/go/authenticator"
		"github.com/raywall/cloud-service-pack/go/authenticator/controllers"
		"github.com/raywall/cloud-service-pack/go/authenticator/handlers"
	)

	func main() {
		// 1. Define the credentials for the service.
		creds := &handlers.Credential{
			ClientID:     "your-client-id",
			ClientSecret: "your-client-secret",
		}

		// 2. Create the specific authentication handler. In this case, for an STS.
		// The URL should point to your identity provider's token endpoint.
		stsHandler := handlers.NewSTSAuthHandler(
			"https://your-sts-service.com/oauth/token",
			creds,
			nil, // Use nil for default options.
		)

		// 3. Declare a variable that will hold the access token.
		// The controller will update this variable automatically.
		var accessToken string
		output := controllers.OutputAccessToken(&accessToken)

		// 4. Create the authenticator and chain it with the token controller.
		auth := authenticator.NewAuthenticator(stsHandler).WithController(output)

		// 5. Start the controller. It will fetch the initial token and
		// start the background refresh loop.
		if err := auth.Start(); err != nil {
			log.Fatalf("Failed to start authenticator: %v", err)
		}
		defer auth.Stop() // Ensure the background goroutine is stopped on exit.

		// 6. Your application can now get the token whenever needed.
		// The controller ensures the token is always valid.
		for i := 0; i < 3; i++ {
			token, err := auth.GetToken()
			if err != nil {
				log.Fatalf("Failed to get token: %v", err)
			}
			fmt.Printf("Successfully retrieved token: %s\n", token)
			time.Sleep(1 * time.Second)
		}
	}
*/
package authenticator
