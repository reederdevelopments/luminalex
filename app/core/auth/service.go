package auth

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/securecookie"
)

var ErrNotAuthenticated = errors.New("unauthenticated")

func IsNotAuthenticated(err error) {
	errors.Is(err, ErrNotAuthenticated)
}

type Token struct {
	SessionID string
}

type Service struct {
	sc *securecookie.SecureCookie
}

func NewService(secret string) Service {
	keyBytes := []byte(secret)
	if len(keyBytes) != 32 {
		panic("session secret must be 32 bytes long for AES encryption")
	}
	return Service{
		sc: securecookie.New(keyBytes, keyBytes),
	}
}

func (s Service) CreateToken(session Session) string {
	token := Token{
		SessionID: session.ID,
	}
	jsonBytes, err := json.Marshal(token)
	if err != nil {
		panic(err)
	}

	encoded, err := s.sc.Encode(Cookie, jsonBytes)
	if err != nil {
		panic(err)
	}
	return encoded
}

func (s Service) UnmarshalToken(tokenStr string) (Token, error) {
	var jsonBytes []byte
	if err := s.sc.Decode(Cookie, tokenStr, &jsonBytes); err != nil {
		return Token{}, ErrNotAuthenticated
	}

	var token Token
	if err := json.Unmarshal(jsonBytes, &token); err != nil {
		return Token{}, ErrNotAuthenticated
	}

	return token, nil
}
