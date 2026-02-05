package token

import (
	"time"

	"aidanwoods.dev/go-paseto"
)

const (
	PayloadKey = "payload"
)

type PasetoMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

func NewPasetoMaker(secretKey string) (Maker, error) {
	keyBytes := []byte(secretKey)
	v4SymmetricKey, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
	if err != nil {
		return nil, err
	}
	maker := &PasetoMaker{
		symmetricKey: v4SymmetricKey,
	}
	return maker, nil
}

func (maker *PasetoMaker) CreateToken(
	id int64,
	email string,
	role string,
	name string,
	instituteID int32,
	duration time.Duration,
) (string, *TokenPayload, error) {

	payload := &TokenPayload{
		ID:          id,
		Email:       email,
		Role:        role,
		Name:        name,
		InstituteID: instituteID,
		IssuedAt:    time.Now(),
		ExpiredAt:   time.Now().Add(duration),
	}

	t := paseto.NewToken()
	t.SetIssuedAt(payload.IssuedAt)
	t.SetExpiration(payload.ExpiredAt)
	t.Set(PayloadKey, payload)

	token := t.V4Encrypt(maker.symmetricKey, nil)
	return token, payload, nil
}

func (maker *PasetoMaker) VerifyToken(tokenString string) (*TokenPayload, error) {
	parser := paseto.NewParser()
	t, err := parser.ParseV4Local(maker.symmetricKey, tokenString, nil)
	if err != nil {
		return nil, err
	}
	payload := TokenPayload{}
	t.Get(PayloadKey, &payload)
	payload.ExpiredAt, _ = t.GetExpiration()
	payload.IssuedAt, _ = t.GetIssuedAt()
	return &payload, nil
}
