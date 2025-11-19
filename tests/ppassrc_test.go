package tests

import (
	"bytes"
	"ppassrc/ppassrc"
	"testing"
	"time"
)

// 1. Unlinkability (sanity check on blinded requests)
func TestUnlinkability(t *testing.T) {
	issuer, _ := ppassrc.NewIssuer()
	client, _ := ppassrc.NewClient(issuer.VerificationKey())

	ctx1 := ppassrc.NewContextRandomEpoch()
	ctx2 := ppassrc.NewContextRandomEpoch()

	b1, _, _ := client.Request(ctx1)
	b2, _, _ := client.Request(ctx2)

	if bytes.Equal(b1.Blinded, b2.Blinded) {
		t.Fatal("unlinkability violated: blinded requests are equal")
	}
}

// 2. One-more unforgeability style: double spend rejected
func TestOMUF(t *testing.T) {
	issuer, _ := ppassrc.NewIssuer()
	client, _ := ppassrc.NewClient(issuer.VerificationKey())
	ctx := ppassrc.NewContextRandomEpoch()

	b, aux, _ := client.Request(ctx)
	eval, _ := issuer.Issue(b)
	tok, _ := client.Finalize(eval, aux)

	ok, _ := issuer.Redeem(ctx, tok)
	if !ok {
		t.Fatal("first redemption failed")
	}

	ok, _ = issuer.Redeem(ctx, tok)
	if ok {
		t.Fatal("double redemption should fail")
	}
}

// 3. Targeted-context UF: preminted tokens fail on fresh unpredictable context
func TestTCUF(t *testing.T) {
	issuer, _ := ppassrc.NewIssuer()
	client, _ := ppassrc.NewClient(issuer.VerificationKey())

	var preminted []*ppassrc.Token

	for i := 0; i < 5; i++ {
		ctx := ppassrc.NewContextRandomEpoch()
		b, aux, _ := client.Request(ctx)
		ev, _ := issuer.Issue(b)
		tok, _ := client.Finalize(ev, aux)
		preminted = append(preminted, tok)
	}

	ctxStar := ppassrc.NewContextRandomEpoch()

	for _, tok := range preminted {
		ok, _ := issuer.Redeem(ctxStar, tok)
		if ok {
			t.Fatal("preminted token redeemed under fresh ctx* (violates targeted-context UF)")
		}
	}
}

func TestTimeWindowContextIsolation(t *testing.T) {
	issuer, _ := ppassrc.NewIssuer()
	client, _ := ppassrc.NewClient(issuer.VerificationKey())

	window := 24 * time.Hour
	monday := time.Date(2025, time.November, 17, 0, 0, 0, 0, time.UTC)
	ctxMon := ppassrc.NewContextTimeWindow(monday, window)
	ctxTue := ppassrc.NewContextTimeWindow(monday.Add(window), window)

	b, aux, _ := client.Request(ctxMon)
	ev, _ := issuer.Issue(b)
	tok, _ := client.Finalize(ev, aux)

	if ok, _ := issuer.Redeem(ctxTue, tok); ok {
		t.Fatal("redemption should fail when context shifts to the next window")
	}

	if ok, _ := issuer.Redeem(ctxMon, tok); !ok {
		t.Fatal("redemption failed with the original window context")
	}
}

// 4. Robustness: honest redemption succeeds despite noisy clients
func TestRobustness(t *testing.T) {
	issuer, _ := ppassrc.NewIssuer()
	clientA, _ := ppassrc.NewClient(issuer.VerificationKey())
	clientB, _ := ppassrc.NewClient(issuer.VerificationKey())

	ctx := ppassrc.NewContextRandomEpoch()

	b, aux, _ := clientA.Request(ctx)
	ev, _ := issuer.Issue(b)
	tok, _ := clientA.Finalize(ev, aux)

	// Noise
	for i := 0; i < 10; i++ {
		cx := ppassrc.NewContextRandomEpoch()
		b2, aux2, _ := clientB.Request(cx)
		ev2, _ := issuer.Issue(b2)
		tok2, _ := clientB.Finalize(ev2, aux2)
		issuer.Redeem(cx, tok2)
	}

	ok, _ := issuer.Redeem(ctx, tok)
	if !ok {
		t.Fatal("robustness violated: honest token rejected after noise")
	}
}
