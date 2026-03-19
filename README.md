# trust-portal

Public trust material distribution for LinnemanLabs infrastructure.

`trust-portal` is a Go service that serves cryptographic trust anchors, CA certificates, public keys, and verification metadata for the LinnemanLabs Sigstore and SPIFFE trust infrastructure. It runs at [trust.linnemanlabs.com](https://trust.linnemanlabs.com).

## What it serves

**Machine-readable discovery endpoints**

- `/.well-known/trusted-root.json` — Sigstore trusted root for cosign verification (Fulcio CA, Rekor, TesseraCT, TSA)
- `/.well-known/signing-config.json` — Sigstore signing configuration

**CA certificates**

- `/certs/root-ca.crt` — LinnemanLabs Root CA (YubiKey-backed, ECCP384, 10yr)
- `/certs/fulcio-ca.crt` — Fulcio CA for keyless code signing
- `/certs/spire-ca.crt` — SPIRE CA for workload identity (SPIFFE SVIDs)
- `/certs/tsa.crt` — Timestamp Authority certificate

**Public keys**

- `/keys/rekor-checkpoint.pub` — Rekor transparency log checkpoint signing key
- `/keys/tesseract-checkpoint.pub` — TesseraCT certificate transparency log checkpoint signing key

**Human-readable pages**

- `/` — Trust portal index
- `/cps` — Certificate Practice Statement

## Architecture

The server loads all trust material from a `data/` directory into memory at startup and serves it from an in-memory file map. No disk I/O per request. Missing files at startup cause a hard panic to catch deployment errors immediately.

Routes are organized into chi route groups (`/.well-known`, `/certs`, `/keys`) for clean OTel trace annotations and prometheus metrics.

### Key details

- **Language:** Go
- **Logging:** [linnemanlabs/go-core/log](https://github.com/linnemanlabs/go-core)
- **Observability:** OpenTelemetry instrumentation, Prometheus metrics, trace response headers
- **Deployment:** Single static binary built using my [build-system](https://github.com/keithlinneman/build-system), deployed via attested CI/CD pipeline with cosign-verified artifacts

## Trust infrastructure context

This portal is one component of a self-hosted Sigstore and SPIFFE stack:

- **Rekor** (rekor-tiles/Tessera) — Signature transparency log
- **Fulcio** — Keyless certificate authority (GitHub OIDC identity -> short-lived signing certs)
- **TesseraCT** (Tessera) — Certificate transparency log for Fulcio
- **TSA** (timestamp-authority) — RFC 3161 timestamping
- **SPIRE/SPIFFE** — Workload identity and mTLS across the environment

The `trusted_root.json` served here is the anchor clients use to verify that entire chain.

## Verification

You can verify any certificate served by this portal against the root:

```bash
# Fetch and inspect the root CA
curl -s https://trust.linnemanlabs.com/certs/root-ca.crt | openssl x509 -text -noout

# Verify a leaf cert chains to root
curl -s https://trust.linnemanlabs.com/certs/tsa.crt -o tsa.crt
curl -s https://trust.linnemanlabs.com/certs/root-ca.crt -o root-ca.crt
openssl verify -CAfile root-ca.crt tsa.crt
```

Use the trusted root with cosign:

```bash
cosign verify-blob \
  --trusted-root <(curl -s https://trust.linnemanlabs.com/.well-known/trusted-root.json) \
  --bundle artifact.bundle \
  artifact
```

## License

MIT. Copy it, steal it, modify it, learn from it, share your improvements with me. Or don't. It's code, do what you want with it.