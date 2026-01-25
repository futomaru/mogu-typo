package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/futomaru/mogu-typo/internal/diff"
	"github.com/futomaru/mogu-typo/internal/engine"
	"github.com/futomaru/mogu-typo/internal/extract"
	"github.com/futomaru/mogu-typo/internal/report"
)

const (
	exitOK    = 0
	exitFound = 1
	exitError = 2
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	opts, err := parseArgs(args)
	if err != nil {
		if err == flag.ErrHelp {
			return exitOK
		}
		fmt.Fprintln(os.Stderr, err)
		usage(os.Stderr)
		return exitError
	}

	eng, err := loadEngine(".mogu-typo/allow.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}

	lines, err := diff.RunGitDiff(opts.base, opts.head)
	if err != nil {
		fmt.Fprintln(os.Stderr, "git diff error:", err)
		return exitError
	}

	findings := collectFindings(lines, eng)
	if err := writeFindings(os.Stdout, opts.format, findings); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}

	if len(findings) > 0 {
		return exitFound
	}
	return exitOK
}

type runOptions struct {
	base   string
	head   string
	format string
}

func parseArgs(args []string) (runOptions, error) {
	if len(args) < 2 || args[1] != "diff" {
		return runOptions{}, fmt.Errorf("expected 'diff' subcommand")
	}

	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := fs.String("base", "origin/main", "base ref")
	head := fs.String("head", "HEAD", "head ref")
	format := fs.String("format", "text", "output format (text/github)")
	if err := fs.Parse(args[2:]); err != nil {
		return runOptions{}, err
	}
	return runOptions{
		base:   *base,
		head:   *head,
		format: strings.ToLower(strings.TrimSpace(*format)),
	}, nil
}

func loadEngine(allowlist string) (*engine.Engine, error) {
	eng := engine.New()
	eng.LoadBuiltin()
	if err := eng.LoadAllowlist(allowlist); err != nil {
		return nil, fmt.Errorf("allowlist load error: %w", err)
	}
	return eng, nil
}

func collectFindings(lines []diff.Line, eng *engine.Engine) []report.Finding {
	extractOpts := extract.Options{
		IgnoreURLs:    true,
		IgnoreAllCaps: true,
	}

	inCodeFence := make(map[string]bool)

	findings := make([]report.Finding, 0, len(lines))
	for _, line := range lines {
		if line.File == "" {
			continue
		}

		if extract.IsCodeFence(line.Text) {
			inCodeFence[line.File] = !inCodeFence[line.File]
			continue
		}
		if inCodeFence[line.File] {
			continue
		}

		words := extract.ExtractWordsWithOptions(line.Text, extractOpts)
		for _, word := range words {
			if isTypo, corrections := eng.Check(word.Text); isTypo {
				findings = append(findings, report.Finding{
					File:        line.File,
					Line:        line.LineNo,
					Col:         word.Col,
					Word:        word.Text,
					Corrections: corrections,
				})
			}
		}
	}

	return findings
}

func writeFindings(w io.Writer, format string, findings []report.Finding) error {
	var err error
	switch format {
	case "text":
		err = report.WriteText(w, findings)
	case "github":
		err = report.WriteGitHub(w, findings)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("output error: %w", err)
	}
	return nil
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  mogutypo diff [--base origin/main] [--head HEAD] [--format text|github]")
}
