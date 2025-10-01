// Package main implements a small tool that processes files containing frames
// separated by a fixed separator line (80 box-drawing '─' characters). It reads
// a file, splits it into frames, filters empty frames and merges frames to leave only
// key frames - frames which were captured with partial content are dropped, and
// writes the resulting content back to the original file. This makes the diff tool
// stable when comparing textual output from vhs.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	sepLine = strings.Repeat("─", 80)
	emptyRe = regexp.MustCompile(`^\s*>?\s*$`)
)

func frameSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}
	sep := []byte(sepLine)

	searchStart := 0
	for {
		idx := bytes.Index(data[searchStart:], sep)
		if idx < 0 {
			break
		}
		idx += searchStart
		end := idx + len(sep)

		// Ensure separator is on its own line
		if (idx == 0 || data[idx-1] == '\n') && (end == len(data) || data[end] == '\n') {
			adv := end
			if end < len(data) && data[end] == '\n' {
				adv++
			}
			return adv, data[:idx], nil
		}
		searchStart = idx + 1
	}

	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func countNonEmptyLines(s string) int {
	if len(s) == 0 {
		return 0
	}
	sc := bufio.NewScanner(strings.NewReader(s))
	count := 0
	for sc.Scan() {
		if emptyRe.Match(sc.Bytes()) {
			continue
		}
		count++
	}
	return count
}

func isIgnorableFrame(frame string) bool {
	return countNonEmptyLines(frame) == 0
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s `path/to/file`", os.Args[0])
	}
	path := os.Args[1]

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("error reading file: %s", err)
	}

	reader := bytes.NewReader(data)
	sc := bufio.NewScanner(reader)
	sc.Split(frameSplit)

	var frames []string
	for sc.Scan() {
		s := sc.Text()
		s = strings.TrimSuffix(s, "\n")
		if len(s) > 0 {
			frames = append(frames, s)
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var out []string
	for _, f := range frames {
		if isIgnorableFrame(f) {
			continue
		}
		if len(out) > 0 && strings.HasPrefix(f, strings.TrimSpace(out[len(out)-1])) {
			out[len(out)-1] = f
			continue
		}
		out = append(out, f)
	}

	err = os.WriteFile(path, []byte(strings.Join(out, "\n"+sepLine+"\n")), 0644)
	if err != nil {
		log.Fatalf("error writing file: %s", err)
	}
}
