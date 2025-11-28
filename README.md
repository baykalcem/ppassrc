# PPass-RC: A Ristretto255 VOPRF-Based Privacy Pass Construction

This repository contains a complete Go implementation of **PPass-RC**, a Privacy-Pass–compatible protocol described in:

**"Security Analysis of Privately Verifiable Privacy Pass" (ePrint 2025/1847)**  
https://eprint.iacr.org/2025/1847

The protocol provides unlinkable, verifiable, context-bound tokens using an **RFC-compatible Ristretto255 VOPRF**.

This implementation follows the algorithms in the paper as closely as possible and includes:
- Security tests  
- Benchmarks  
- A runnable end-to-end demo

---


###  Security Property Tests

Includes tests modeling the formal properties defined in the paper:
- Unlinkability  
- One-More Unforgeability  
- Targeted-Context Unforgeability  
- Robustness

###  Benchmarks

Benchmarks issuance and redemption performance using Go's native testing framework.

###  Minimal and Reproducible

Uses an actively maintained VOPRF implementation:
```go
github.com/bytemare/voprf
```

which implements the standard **Ristretto255-SHA512** VOPRF ciphersuite from the IETF draft/RFC.

---

##  Repository Structure

```
ppassrc/
├── go.mod
├── go.sum
├── main.go                    # end-to-end example (issue + redeem)
├── ppassrc/
│   ├── client.go              # client token request + finalize logic
│   ├── issuer.go              # issuer keygen, issuance, redemption
│   ├── context.go             # hashing utilities for H(ctx || nonce)
│   ├── types.go               # Token struct and shared definitions
│   └── crypto.go              # intentionally minimal (VOPRF handles crypto internally)
└── tests/
    ├── security_test.go       # formal security property tests
    └── benchmark_test.go      # issuance + redemption benchmarks
```

---

##  Running Tests

```bash
go test ./...
```

---

##  Running Benchmarks

```bash
go test -bench=. ./...
```


##  Running the End-to-End Demo

```bash
go run main.go
```

Expected output:
```
First redemption: true
Second redemption (should be false): false
```

This confirms correctness and one-time-use enforcement.

---

##  Security Model

PPass-RC inherits its security from the underlying VOPRF primitive:

- **Unlinkability** — guaranteed by VOPRF blindness  
- **One-More Unforgeability** — follows from VOPRF PRF-security  
- **Targeted-Context Unforgeability** — from binding `ctx` and `nonce` into the PRF input  
- **Robustness** — guaranteed by `VerifyFinalize` and the spent-list

The implementation directly mirrors the formal algorithms provided in the paper.

---

##  Reference

**Konrad Hanff, Anja Lehmann, and Cavit Özbay.**  
*Security Analysis of Privately Verifiable Privacy Pass.*  
Hasso Plattner Institute, University of Potsdam.  
Cryptology ePrint Archive, Paper 2025/1847.  
https://eprint.iacr.org/2025/1847

---


##  License

MIT License

Cem Baykal: cbaykal@cs.unc.edu
Jiangyuan Yuan: jyuan@cs.unc.edu

