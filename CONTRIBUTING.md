# Contributing to powerctl-cli

Thanks for taking an interest — contributions are genuinely welcome, whether
that's a bug report, a docs fix, or a new command.

## Ways to contribute

- **Report a bug** — use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).
  Include your OS, `powerctl version` output, and the exact command you ran.
- **Suggest a feature** — use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md).
- **Pick up planned work** — use the [task template](.github/ISSUE_TEMPLATE/task.md).
- **Fix the docs** — use the [documentation template](.github/ISSUE_TEMPLATE/documentation.md).
- **Send a pull request** — see below.

## The issue carries the plan

This project keeps the *thinking* in the issue and the *code* in the pull
request. Before writing anything beyond a trivial fix, open an issue that sets
out:

- **Goal** — the outcome you're after, and why. The end state, not the
  implementation.
- **Plan** — the approach, the command surface, and the shape of the
  `--format json` output if it's user-facing.
- **Tasks** — the checklist the PR will work through.

The [task template](.github/ISSUE_TEMPLATE/task.md) gives you this structure
directly; the feature template covers the same ground for ideas that aren't
scoped yet.

Two reasons this matters. It's far cheaper to disagree about a design before
the code exists than after, and it means the PR can stay short — a reviewer
who wants context follows the link rather than reconstructing it from the diff.

Every PR should then link back with `Closes #N` (if the issue is fully resolved)
or `Relates to #N`. GitHub will connect the two, and `Closes` will close the
issue on merge.

## Development setup

You need Go 1.23 or newer (`go.mod` sets `go 1.23.0`). The CI matrix also
lists 1.22.x, which works only via Go's automatic toolchain download — treat
1.23 as the real floor.

```bash
git clone https://github.com/kristofferrisa/powerctl-cli.git
cd powerctl-cli
make build
./powerctl --help
```

To run against real data you need a Tibber API token from
[developer.tibber.com/settings/access-token](https://developer.tibber.com/settings/access-token):

```bash
export TIBBER_TOKEN="your-token"
./powerctl home
```

Tibber also publishes a demo token that returns sample data, which is useful if
you don't have a Tibber subscription or a Pulse.

### Common commands

```bash
make build          # Build ./powerctl
make test           # go test -v ./...
make fmt            # Format code
make lint           # golangci-lint (optional — not run in CI)
make tidy           # go mod tidy
make build-all      # Cross-compile linux/darwin/windows
```

## Project layout

Read [ARCHITECTURE.md](ARCHITECTURE.md) for the full picture. The short version:

```
cmd/powerctl/main.go     Entry point
internal/api/            GraphQL client, queries, WebSocket
internal/commands/       One file per command
internal/config/         Config loading (flags > env > file)
internal/models/         Data structures
internal/output/         Formatter interface + pretty/json/markdown
```

The flow is: **command → API client → formatter → stdout**.

### Adding a new command

Most new features follow the same path. Using a hypothetical `powerctl foo`:

1. Add the GraphQL query to `internal/api/queries.go`.
2. Add the response types to `internal/models/types.go`.
3. Add a client method to `internal/api/client.go`.
4. Add `internal/commands/foo.go`, following `prices.go` as the model — validate
   `cfg`, build the client, call it, hand the result to `formatter`, and register
   the command in `init()`.
5. Add the method to the `Formatter` interface in `internal/output/formatter.go`
   — and implement it in **all three** formatters: `pretty.go`, `json.go` and
   `markdown.go`.
6. Add tests, and document the command in `README.md` and `ARCHITECTURE.md`.

### Design principles

Keep these in mind — they're why the tool looks the way it does:

- **Do one thing well.** One command, one responsibility.
- **Composable output.** `--format json` should be stable and pipeable into
  `jq`. Treat it as an API, not as debug output.
- **Fail fast, fail loud.** Clear error messages, exit code 1 on error.
- **Never print or log the token.**

## Code style

- Format with `gofmt -s -w .` before committing. CI enforces `gofmt -s`, and
  note that `make fmt` runs plain `go fmt` — it does *not* apply `-s`, so run
  `gofmt -s -w .` if CI complains.
- `go vet ./...` must be clean. CI enforces this.
- There's no `.golangci.yml` in the repo, so `make lint` is a local convenience
  rather than a gate. Don't feel obliged to chase its output.
- Match the surrounding code. Comment the non-obvious parts, not the obvious ones.

## Testing

```bash
go test ./...
go test -race ./...                      # what CI runs on Linux/macOS
go test -v ./internal/config -run TestLoad_EnvVarTakesPriority
```

- API tests mock HTTP responses (see `internal/api/client_test.go`).
- Formatter tests compare rendered output (see `internal/output/formatter_test.go`).
- CI runs the suite on Linux, macOS and Windows, so avoid anything
  platform-specific in path handling or terminal output.

## Branches and commits

- Branch from `main`. Name branches `feat-<something>` or `fix-<something>` —
  CI runs on pushes to those prefixes.
- Commit messages: short and descriptive. No strict convention is enforced.
- Keep a PR to one logical change. Unrelated refactors are much harder to review.

## Pull requests

Fill in the [pull request template](.github/pull_request_template.md). It asks
you to restate the Goal and mirror the Tasks from the linked issue, so a
reviewer can see at a glance what's done and what was deliberately left out.

Make sure:

- [ ] `go test ./...` passes locally
- [ ] `gofmt -s -l .` reports nothing
- [ ] `go vet ./...` is clean
- [ ] Docs updated if behaviour changed
- [ ] `--format json` output is unchanged, or the break is called out
- [ ] The related issue is linked with `Closes #N` or `Relates to #N`

CI (lint, test, build) must be green before merge.

## AI-assisted contributions

Using AI tooling to write your contribution is fine — this project is built with
it too. Two expectations:

1. **Say so in the PR description.** A single line is plenty.
2. **You own the code.** Read it, test it, and be able to explain why it works.
   PRs that clearly haven't been reviewed by their author will be sent back.

The repo has a [CLAUDE.md](CLAUDE.md) with project conventions that's useful to
point your tooling at.

## Security

- Never commit a real API token. Tokens go in `TIBBER_TOKEN` or the config file
  (`~/.tibber/config.yaml`), never in the repo.
- If you find a security issue, please don't open a public issue. See the
  [security policy](.github/SECURITY.md) for how to report it privately.

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE).
