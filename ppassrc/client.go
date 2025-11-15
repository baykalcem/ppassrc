package ppassrc

import (
	"crypto/rand"

	"github.com/bytemare/voprf"
)

type Client struct {
	cs   voprf.Identifier
	impl *voprf.Client
}

// NewClient takes the issuer's public key (as bytes) and instantiates a VOPRF client.
func NewClient(pubKey []byte) (*Client, error) {
	cs := voprf.Ristretto255Sha512
	cli, err := cs.Client(voprf.VOPRF, pubKey)
	if err != nil {
		return nil, err
	}
	return &Client{
		cs:   cs,
		impl: cli,
	}, nil
}

// Request implements the PPass-RC Request algorithm:
// - sample nonce
// - compute msg = Hctx(ctx, nonce)
// - blind msg with VOPRF client
func (c *Client) Request(ctx Context) (BlindedToken, RequestAux, error) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)

	msg := Hctx(ctx, nonce)

	blinded := c.impl.Blind(msg, nil) // VOPRF handles blind scalar internally

	aux := RequestAux{
		Nonce: nonce,
	}
	return BlindedToken{Blinded: blinded}, aux, nil
}

// Finalize unblinds the issuer's evaluation and returns the usable token.
func (c *Client) Finalize(eval *Evaluation, aux RequestAux) (*Token, error) {
	ev := new(voprf.Evaluation)
	if err := ev.Deserialize(eval.Eval); err != nil {
		return nil, err
	}

	out, err := c.impl.Finalize(ev, nil)
	if err != nil {
		return nil, err
	}

	return &Token{
		Value: out,
		Nonce: aux.Nonce,
	}, nil
}