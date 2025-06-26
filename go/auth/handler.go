package auth

type Handler interface {
	GetToken() (string, error)
	Start() error
	Stop()
	refreshLoop()
	refreshToken() error
}
