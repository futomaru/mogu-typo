package extract

import (
	"regexp"
	"strings"
	"unicode"
)

type Word struct {
	Text string
	Col  int
}

type Options struct {
	IgnoreURLs    bool
	IgnoreAllCaps bool
}

var (
	urlRegex = regexp.MustCompile(`https?://\S+|ftp://\S+`)

	wordRegex = regexp.MustCompile(`[a-zA-Z]+`)
)

func ExtractWordsWithOptions(text string, opts Options) []Word {
	cleaned := text
	if opts.IgnoreURLs && strings.Contains(text, "://") {
		cleaned = urlRegex.ReplaceAllStringFunc(text, func(match string) string {
			return strings.Repeat(" ", len(match))
		})
	}

	matches := wordRegex.FindAllStringIndex(cleaned, -1)
	words := make([]Word, 0, len(matches))
	for _, match := range matches {
		start, end := match[0], match[1]
		word := cleaned[start:end]

		if len(word) < 2 {
			continue
		}

		if opts.IgnoreAllCaps && isAllUpper(word) {
			continue
		}

		if hasAdjacentDigit(text, start, len(word)) {
			continue
		}

		words = append(words, Word{
			Text: strings.ToLower(word),
			Col:  start + 1,
		})
	}

	return words
}

func isAllUpper(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func hasAdjacentDigit(text string, pos, length int) bool {
	if pos > 0 && text[pos-1] >= '0' && text[pos-1] <= '9' {
		return true
	}
	if end := pos + length; end < len(text) && text[end] >= '0' && text[end] <= '9' {
		return true
	}
	return false
}

func IsCodeFence(text string) bool {
	trimmed := strings.TrimSpace(text)
	return strings.HasPrefix(trimmed, "```")
}
