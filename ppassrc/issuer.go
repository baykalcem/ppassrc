package ppassrc

import (
	"sync"

	"github.com/bytemare/voprf"
)

type Issuer struct {
	cs    voprf.Identifier
	srv   *voprf.Server
	pk    []byte
	mu    sync.Mutex
	spent map[string]bool
}

// NewIssuer runs Kg: generate a VOPRF key pair and server instance.
func NewIssuer() (*Issuer, error) {
	cs := voprf.Ristretto255Sha512

	kp := cs.KeyGen() // returns *KeyPair with PublicKey, SecretKey
	srv, err := cs.Server(voprf.VOPRF, kp.SecretKey)
	if err != nil {
		return nil, err
	}

	return &Issuer{
		cs:    cs,
		srv:   srv,
		pk:    kp.PublicKey,
		spent: make(map[string]bool),
	}, nil
}

// VerificationKey returns the issuer's encoded public key (to give to clients).
func (iss *Issuer) VerificationKey() []byte {
	return iss.pk
}

// Issue runs the VOPRF evaluation on the blinded input.
func (iss *Issuer) Issue(b BlindedToken) (*Evaluation, error) {
	eval, err := iss.srv.Evaluate(b.Blinded, nil)
	if err != nil {
		return nil, err
	}
	return &Evaluation{Eval: eval.Serialize()}, nil
}

// Redeem verifies the PRF output and enforces one-time-use (double-spend prevention).
func (iss *Issuer) Redeem(ctx Context, tok *Token) (bool, error) {
	msg := Hctx(ctx, tok.Nonce)

	// Check PRF validity for this msg under issuer's key.
	if !iss.srv.VerifyFinalize(msg, nil, tok.Value) {
		return false, nil
	}

	key := string(tok.Value)

	iss.mu.Lock()
	defer iss.mu.Unlock()

	if iss.spent[key] {
		return false, nil
	}

	iss.spent[key] = true
	return true, nil
}

// ResetForBench just clears spent state for a given token (used by benchmarks).
func (iss *Issuer) ResetForBench(tok *Token) {
	iss.mu.Lock()
	delete(iss.spent, string(tok.Value))
	iss.mu.Unlock()
}