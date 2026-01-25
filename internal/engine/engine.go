package engine

import (
	"bufio"
	"bytes"
	_ "embed"
	"io"
	"os"
	"strings"
)

//go:embed dict/en-basic.txt
var embeddedDict []byte

type Engine struct {
	typos   map[string][]string
	allowed map[string]struct{}
}

func New() *Engine {
	return &Engine{
		typos:   make(map[string][]string),
		allowed: make(map[string]struct{}),
	}
}

func (e *Engine) LoadBuiltin() {
	_ = e.loadTypos(bytes.NewReader(embeddedDict))
}

func (e *Engine) LoadAllowlist(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		e.allowed[strings.ToLower(word)] = struct{}{}
	}
	return scanner.Err()
}

func (e *Engine) Check(word string) (bool, []string) {
	lower := strings.ToLower(word)
	if _, ok := e.allowed[lower]; ok {
		return false, nil
	}
	if corrections, ok := e.typos[lower]; ok {
		return true, corrections
	}
	return false, nil
}

func (e *Engine) loadTypos(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "->", 2)
		if len(parts) != 2 {
			continue
		}
		typo := strings.ToLower(strings.TrimSpace(parts[0]))
		if typo == "" {
			continue
		}
		var corrections []string
		for _, c := range strings.Split(parts[1], ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				corrections = append(corrections, c)
			}
		}
		if len(corrections) > 0 {
			e.typos[typo] = corrections
		}
	}
	return scanner.Err()
}
