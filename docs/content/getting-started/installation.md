---
title: "Installation"
description: "Install smashingmag from a release, with go install, or from source."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/smashingmag-cli-cli/releases) carries archives for Linux, macOS,
and Windows on amd64 and arm64, plus deb, rpm, and apk packages for Linux.
Download, unpack, put `smashingmag` on your `PATH`, done. The `checksums.txt`
on each release is signed with keyless [cosign](https://docs.sigstore.dev/) if
you want to verify before running.

## With Go

```bash
go install github.com/tamnd/smashingmag-cli-cli/cmd/smashingmag@latest
```

That puts `smashingmag` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless
you moved it. Make sure that directory is on your `PATH`.

## From source

```bash
git clone https://github.com/tamnd/smashingmag-cli-cli
cd smashingmag-cli-cli
make build        # produces ./bin/smashingmag
./bin/smashingmag version
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/smashingmag:latest --help
```

## Checking the install

```bash
smashingmag version
```

prints the version and exits.
