package ppassrc

import "crypto/sha512"

// Hctx hashes (ctx || nonce) into a PRF input.
func Hctx(ctx Context, nonce []byte) []byte {
	h := sha512.New()
	h.Write(ctx)
	h.Write(nonce)
	return h.Sum(nil)
}