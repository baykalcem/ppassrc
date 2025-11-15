package ppassrc

import "crypto/rand"

// BlindedToken is the blinded message sent from client to issuer.
type BlindedToken struct {
	Blinded []byte
}

// Evaluation is the issuer's VOPRF evaluation response (serialized).
type Evaluation struct {
	Eval []byte
}

// Token is the finalized, unblinded token the client uses for redemption.
type Token struct {
	Value []byte
	Nonce []byte
}

// RequestAux stores client-side state needed between Request and Finalize.
type RequestAux struct {
	Nonce []byte
}

// Context is the redemption context (epoch, origin, etc.).
type Context []byte

func NewContext(b []byte) Context { return Context(b) }

// NewContextRandomEpoch generates a fresh random epoch-like context.
func NewContextRandomEpoch() Context {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return Context(append([]byte("epoch:"), b...))
}