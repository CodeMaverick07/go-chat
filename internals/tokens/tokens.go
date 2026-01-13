package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"time"
)

const (
	ScopeAuth string = "login"
)

type Token struct {
	PlainText string    `json:"token"`
	Hash      string    `json:"-"`
	UserId    string    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func GenerateToken(UserId string, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserId: UserId,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	emptyBytes := make([]byte, 32)
	_, err := rand.Read(emptyBytes)
	if err != nil {
		return nil, err
	}
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(emptyBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hex.EncodeToString(hash[:])
	return token, nil
}
