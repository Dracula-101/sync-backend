package common

const (
	payloadApiKey   string = "apikey"
	payloadUser     string = "user"
)

type ContextPayload interface {
}

type payload struct{}

func NewContextPayload() ContextPayload {
	return &payload{}
}
