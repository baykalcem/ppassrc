package tests

import (
	"ppassrc/ppassrc"
	"testing"
)

func BenchmarkIssuance(b *testing.B) {
	issuer, _ := ppassrc.NewIssuer()
	client, _ := ppassrc.NewClient(issuer.VerificationKey())
	ctx := ppassrc.NewContextRandomEpoch()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bl, aux, _ := client.Request(ctx)
		ev, _ := issuer.Issue(bl)
		client.Finalize(ev, aux)
	}
}

func BenchmarkRedeem(b *testing.B) {
	issuer, _ := ppassrc.NewIssuer()
	client, _ := ppassrc.NewClient(issuer.VerificationKey())

	ctx := ppassrc.NewContextRandomEpoch()
	bl, aux, _ := client.Request(ctx)
	ev, _ := issuer.Issue(bl)
	tok, _ := client.Finalize(ev, aux)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok, _ := issuer.Redeem(ctx, tok)
		if !ok {
			b.Fatal("unexpected redemption failure in benchmark")
		}
		issuer.ResetForBench(tok)
	}
}