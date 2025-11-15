package main

import (
	"fmt"

	"ppassrc/ppassrc"
)

func main() {
	issuer, err := ppassrc.NewIssuer()
	if err != nil {
		panic(err)
	}

	client, err := ppassrc.NewClient(issuer.VerificationKey())
	if err != nil {
		panic(err)
	}

	ctx := ppassrc.NewContextRandomEpoch()

	blinded, aux, err := client.Request(ctx)
	if err != nil {
		panic(err)
	}

	eval, err := issuer.Issue(blinded)
	if err != nil {
		panic(err)
	}

	token, err := client.Finalize(eval, aux)
	if err != nil {
		panic(err)
	}

	ok, err := issuer.Redeem(ctx, token)
	if err != nil {
		panic(err)
	}
	fmt.Println("First redemption:", ok)

	ok, err = issuer.Redeem(ctx, token)
	fmt.Println("Second redemption (should be false):", ok)
}



