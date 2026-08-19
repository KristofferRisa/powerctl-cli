# Security Policy

## Supported Versions

This project is pre-1.0. Security fixes land on `main` and go out in the next
release; older releases are not patched.

## Reporting a Vulnerability

**Please do not report security issues in a public GitHub issue.**

Report privately through GitHub's
[private vulnerability reporting](https://github.com/kristofferrisa/powerctl-cli/security/advisories/new)
on this repository's Security tab. If that isn't available to you, open a
regular issue asking for a private contact channel — without including any
details of the vulnerability.

Please include:

- The version affected (`powerctl version`)
- Steps to reproduce
- What an attacker could achieve

You can expect an initial response within a week.

## Scope

Things that are in scope:

- Leaking or logging the Tibber API token
- Insecure handling of `~/.tibber/config.yaml` (permissions, disclosure)
- TLS or certificate verification weaknesses in the HTTP or WebSocket clients
- Command or path injection through flags, config values, or API responses

Out of scope:

- Vulnerabilities in the Tibber API itself — report those to
  [Tibber](https://developer.tibber.com/)
- Issues that require an attacker to already have write access to your machine
  or your config file

## Handling Your Token

Your Tibber API token grants access to your home's energy data. Treat it like a
password:

- Prefer the `TIBBER_TOKEN` environment variable or `~/.tibber/config.yaml`
- Never commit it, never paste it into an issue, and redact it from any logs or
  terminal output you share
- Revoke and reissue at
  [developer.tibber.com/settings/access-token](https://developer.tibber.com/settings/access-token)
  if you think it's been exposed
