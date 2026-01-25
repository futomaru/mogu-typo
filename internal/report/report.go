package report

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Finding struct {
	File        string
	Line        int
	Col         int
	Word        string
	Corrections []string
}

func WriteText(w io.Writer, findings []Finding) error {
	bw := bufio.NewWriter(w)
	for _, f := range findings {
		if _, err := fmt.Fprintf(bw, "%s:%d:%d  \"%s\" -> %s\n", f.File, f.Line, f.Col, f.Word, strings.Join(f.Corrections, ", ")); err != nil {
			return err
		}
	}
	return bw.Flush()
}

func WriteGitHub(w io.Writer, findings []Finding) error {
	bw := bufio.NewWriter(w)
	for _, f := range findings {
		if _, err := fmt.Fprintf(bw, "::error file=%s,line=%d,col=%d::Spellcheck: \"%s\" -> %s\n", f.File, f.Line, f.Col, f.Word, strings.Join(f.Corrections, ", ")); err != nil {
			return err
		}
	}
	return bw.Flush()
}
