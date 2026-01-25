package diff

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Line struct {
	File   string
	LineNo int
	Text   string
}

var hunkHeaderRegex = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

func Parse(r io.Reader) ([]Line, error) {
	var lines []Line
	var currentFile string
	var lineNo int
	inHunk := false

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "diff ") {
			currentFile = ""
			lineNo = 0
			inHunk = false
			continue
		}

		if strings.HasPrefix(line, "+++ ") {
			path := strings.TrimPrefix(line, "+++ ")
			if strings.HasPrefix(path, "b/") {
				currentFile = strings.TrimPrefix(path, "b/")
			} else if path == "/dev/null" {
				currentFile = ""
			} else {
				currentFile = path
			}
			inHunk = false
			continue
		}

		if matches := hunkHeaderRegex.FindStringSubmatch(line); matches != nil {
			lineNo, _ = strconv.Atoi(matches[1])
			inHunk = true
			continue
		}

		if !inHunk || currentFile == "" {
			continue
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			lines = append(lines, Line{
				File:   currentFile,
				LineNo: lineNo,
				Text:   strings.TrimPrefix(line, "+"),
			})
			lineNo++
			continue
		}

		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "\\") {
			continue
		}
		lineNo++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read diff: %w", err)
	}

	return lines, nil
}

func RunGitDiff(base, head string) ([]Line, error) {
	args := []string{"diff", "--no-color", "-U0"}
	if base != "" {
		if head != "" {
			args = append(args, base+"..."+head)
		} else {
			args = append(args, base)
		}
	}

	cmd := exec.Command("git", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to run git diff: %w", err)
	}

	lines, parseErr := Parse(stdout)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	return lines, parseErr
}
