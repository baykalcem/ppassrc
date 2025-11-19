package tests

import (
	"bytes"
	"runtime"
	"testing"

	"ppassrc/ppassrc"
)

// ------------------------------
// Helper: generate n tokens
// ------------------------------

func makeTokens(b *testing.B, n int) (*ppassrc.Issuer, *ppassrc.Client, ppassrc.Context, []*ppassrc.Token) {
	issuer, err := ppassrc.NewIssuer()
	if err != nil {
		b.Fatalf("NewIssuer: %v", err)
	}
	client, err := ppassrc.NewClient(issuer.VerificationKey())
	if err != nil {
		b.Fatalf("NewClient: %v", err)
	}

	ctx := ppassrc.NewContextRandomEpoch()
	toks := make([]*ppassrc.Token, 0, n)

	for i := 0; i < n; i++ {
		bl, aux, err := client.Request(ctx)
		if err != nil {
			b.Fatalf("Request: %v", err)
		}
		eval, err := issuer.Issue(bl)
		if err != nil {
			b.Fatalf("Issue: %v", err)
		}
		tok, err := client.Finalize(eval, aux)
		if err != nil {
			b.Fatalf("Finalize: %v", err)
		}
		toks = append(toks, tok)
	}

	return issuer, client, ctx, toks
}

// Mark tokens as unspent again
func resetTokens(issuer *ppassrc.Issuer, toks []*ppassrc.Token) {
	for _, tok := range toks {
		issuer.ResetForBench(tok)
	}
}

// ------------------------------
// Batch issuance benchmarks
// ------------------------------

func BenchmarkIssuanceBatch(b *testing.B) {
	batchSizes := []int{1, 5, 10, 25, 50}

	for _, n := range batchSizes {
		b.Run(funcName("batch", n), func(b *testing.B) {
			issuer, err := ppassrc.NewIssuer()
			if err != nil {
				b.Fatalf("NewIssuer: %v", err)
			}
			client, err := ppassrc.NewClient(issuer.VerificationKey())
			if err != nil {
				b.Fatalf("NewClient: %v", err)
			}
			ctx := ppassrc.NewContextRandomEpoch()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < n; j++ {
					bl, aux, err := client.Request(ctx)
					if err != nil {
						b.Fatalf("Request: %v", err)
					}
					eval, err := issuer.Issue(bl)
					if err != nil {
						b.Fatalf("Issue: %v", err)
					}
					_, err = client.Finalize(eval, aux)
					if err != nil {
						b.Fatalf("Finalize: %v", err)
					}
				}
			}
		})
	}
}

// cheap name builder
func funcName(prefix string, n int) string {
	var buf bytes.Buffer
	buf.WriteString(prefix)
	buf.WriteByte('-')

	if n == 0 {
		buf.WriteByte('0')
		return buf.String()
	}

	tmp := [10]byte{}
	i := len(tmp)
	for n > 0 {
		i--
		tmp[i] = byte('0' + (n % 10))
		n /= 10
	}
	buf.Write(tmp[i:])
	return buf.String()
}

// ------------------------------
// Batch redemption
// ------------------------------

func BenchmarkRedeemBatch(b *testing.B) {
	batchSizes := []int{1, 5, 10, 25, 50}

	for _, n := range batchSizes {
		b.Run(funcName("batch-redeem", n), func(b *testing.B) {
			issuer, client, ctx, toks := makeTokens(b, n)
			_ = client

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, tok := range toks {
					ok, err := issuer.Redeem(ctx, tok)
					if err != nil {
						b.Fatalf("Redeem: %v", err)
					}
					if !ok {
						b.Fatalf("Redeem unexpectedly false")
					}
				}
				resetTokens(issuer, toks)
			}
		})
	}
}

// ------------------------------
// Context size scaling
// ------------------------------

func BenchmarkContextSizes(b *testing.B) {
	sizes := []int{8, 64, 1024, 10240}

	for _, size := range sizes {
		b.Run(funcName("ctx-bytes", size), func(b *testing.B) {
			issuer, err := ppassrc.NewIssuer()
			if err != nil {
				b.Fatalf("NewIssuer: %v", err)
			}
			client, err := ppassrc.NewClient(issuer.VerificationKey())
			if err != nil {
				b.Fatalf("NewClient: %v", err)
			}

			ctxBytes := bytes.Repeat([]byte("c"), size)
			ctx := ppassrc.NewContext(ctxBytes)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bl, aux, err := client.Request(ctx)
				if err != nil {
					b.Fatalf("Request: %v", err)
				}
				eval, err := issuer.Issue(bl)
				if err != nil {
					b.Fatalf("Issue: %v", err)
				}
				_, err = client.Finalize(eval, aux)
				if err != nil {
					b.Fatalf("Finalize: %v", err)
				}
			}
		})
	}
}

// ------------------------------
// FIXED: GOMAXPROCS scaling
// ------------------------------

func BenchmarkIssuanceScalingGOMAXPROCS(b *testing.B) {
	procs := []int{1, 2, 4, 8}

	for _, p := range procs {
		b.Run(funcName("procs", p), func(b *testing.B) {
			old := runtime.GOMAXPROCS(p)
			defer runtime.GOMAXPROCS(old)

			issuer, err := ppassrc.NewIssuer()
			if err != nil {
				b.Fatalf("NewIssuer: %v", err)
			}

			ctx := ppassrc.NewContextRandomEpoch()

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				// FIX: Each goroutine must use its own client
				localClient, err := ppassrc.NewClient(issuer.VerificationKey())
				if err != nil {
					b.Fatalf("NewClient: %v", err)
				}

				for pb.Next() {
					bl, aux, err := localClient.Request(ctx)
					if err != nil {
						b.Fatalf("Request: %v", err)
					}
					eval, err := issuer.Issue(bl)
					if err != nil {
						b.Fatalf("Issue: %v", err)
					}
					_, err = localClient.Finalize(eval, aux)
					if err != nil {
						b.Fatalf("Finalize: %v", err)
					}
				}
			})
		})
	}
}

// ------------------------------
// Valid vs invalid redemption
// ------------------------------

func BenchmarkRedeemValid(b *testing.B) {
	issuer, client, ctx, toks := makeTokens(b, 1)
	tok := toks[0]
	_ = client

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok, err := issuer.Redeem(ctx, tok)
		if err != nil {
			b.Fatalf("Redeem: %v", err)
		}
		if !ok {
			b.Fatalf("Redeem unexpectedly false")
		}
		issuer.ResetForBench(tok)
	}
}

func BenchmarkRedeemInvalid(b *testing.B) {
	issuer, client, ctx, toks := makeTokens(b, 1)
	tok := toks[0]
	_ = client

	bad := *tok
	if len(bad.Value) > 0 {
		bad.Value[0] ^= 1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok, err := issuer.Redeem(ctx, &bad)
		if err != nil {
			b.Fatalf("Redeem: %v", err)
		}
		if ok {
			b.Fatalf("Invalid token accepted")
		}
	}
}

// ------------------------------
// Combined issuance + redemption
// ------------------------------

func BenchmarkIssuanceRedeemCombined(b *testing.B) {
	issuer, err := ppassrc.NewIssuer()
	if err != nil {
		b.Fatalf("NewIssuer: %v", err)
	}
	client, err := ppassrc.NewClient(issuer.VerificationKey())
	if err != nil {
		b.Fatalf("NewClient: %v", err)
	}

	ctx := ppassrc.NewContextRandomEpoch()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bl, aux, err := client.Request(ctx)
		if err != nil {
			b.Fatalf("Request: %v", err)
		}
		eval, err := issuer.Issue(bl)
		if err != nil {
			b.Fatalf("Issue: %v", err)
		}
		tok, err := client.Finalize(eval, aux)
		if err != nil {
			b.Fatalf("Finalize: %v", err)
		}
		ok, err := issuer.Redeem(ctx, tok)
		if err != nil {
			b.Fatalf("Redeem: %v", err)
		}
		if !ok {
			b.Fatalf("Redeem unexpectedly false")
		}
		issuer.ResetForBench(tok)
	}
}

// ------------------------------
// Memory overhead per batch
// ------------------------------

func BenchmarkIssuanceMemoryOverhead(b *testing.B) {
	const batchSize = 10

	issuer, err := ppassrc.NewIssuer()
	if err != nil {
		b.Fatalf("NewIssuer: %v", err)
	}
	client, err := ppassrc.NewClient(issuer.VerificationKey())
	if err != nil {
		b.Fatalf("NewClient: %v", err)
	}
	ctx := ppassrc.NewContextRandomEpoch()

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < batchSize; j++ {
			bl, aux, err := client.Request(ctx)
			if err != nil {
				b.Fatalf("Request: %v", err)
			}
			eval, err := issuer.Issue(bl)
			if err != nil {
				b.Fatalf("Issue: %v", err)
			}
			_, err = client.Finalize(eval, aux)
			if err != nil {
				b.Fatalf("Finalize: %v", err)
			}
		}
	}
	b.StopTimer()

	runtime.ReadMemStats(&after)
	if b.N > 0 {
		delta := float64(after.TotalAlloc-before.TotalAlloc) / float64(b.N)
		b.ReportMetric(delta, "bytes/op")
	}
}
