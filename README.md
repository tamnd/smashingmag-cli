# smashingmag

Browse Smashing Magazine web design articles

`smashingmag` is a single pure-Go binary. It reads Smashing Magazine through its
public RSS feed, shapes the responses into clean records, and pipes into the rest
of your tools. No API key, nothing to run alongside it.

## Install

```bash
go install github.com/tamnd/smashingmag-cli/cmd/smashingmag@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/smashingmag-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/smashingmag:latest --help
```

## Usage

```bash
smashingmag articles          # list latest articles (table on TTY, JSONL piped)
smashingmag articles -n 5     # limit to 5 articles
smashingmag articles -o json  # JSON output
smashingmag articles -o url   # print URLs only
smashingmag --help
smashingmag version
```

## Development

```
cmd/smashingmag/   thin main, wires cli.Root into fang
cli/               the cobra command tree
smashingmag/       the library: HTTP client and data models
pkg/render/        output renderer (table/json/jsonl/csv/tsv/url/raw)
docs/              documentation site
```

```bash
make build      # ./bin/smashingmag
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
