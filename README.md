# mogu-typo

Diff-first typo checker for pull requests. Scans only added lines in `git diff` and reports unknown words.

## Install

```bash
go install github.com/futomaru/mogu-typo/cmd/mogutypo@latest
```

## Usage

```bash
mogutypo diff --base origin/main
mogutypo diff --base $BASE_SHA --head $HEAD_SHA --format github
```

| Flag | Default | Description |
|------|---------|-------------|
| `--base` | `origin/main` | Base ref for diff |
| `--head` | `HEAD` | Head ref for diff |
| `--format` | `text` | Output format (`text` / `github`) |

Exit codes: `0` no findings / `1` findings found / `2` error

## GitHub Actions

```yaml
name: Typo Check
on: [pull_request]
jobs:
  typo:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - run: go install github.com/futomaru/mogu-typo/cmd/mogutypo@latest
      - run: mogutypo diff --base origin/${{ github.base_ref }} --format github
```

## Custom Dictionary

Add project-specific words to `.mogu-typo/allow.txt` (one word per line).

## License

MIT. See `LICENSE`.
