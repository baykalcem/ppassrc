package ppassrc

import (
	"crypto/sha512"
	"encoding/binary"
	"time"
)

// Hctx hashes (ctx || nonce) into a PRF input.
func Hctx(ctx Context, nonce []byte) []byte {
	h := sha512.New()
	h.Write(ctx)
	h.Write(nonce)
	return h.Sum(nil)
}

// NewContextTimeWindow builds a redemption context derived from the time window
// containing the provided timestamp.
func NewContextTimeWindow(now time.Time, window time.Duration) Context {
	if window <= 0 {
		panic("ppassrc: window must be positive")
	}

	bucket := now.UnixNano() / window.Nanoseconds()
	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[:8], uint64(bucket))
	binary.BigEndian.PutUint64(data[8:], uint64(window.Nanoseconds()))

	h := sha512.New()
	h.Write([]byte("ppassrc:time-window"))
	h.Write(data)
	return Context(h.Sum(nil))
}
