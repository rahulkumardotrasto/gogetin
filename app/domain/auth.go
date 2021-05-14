package domain

import "time"

//AuthTokenResponse ...
type AuthTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
