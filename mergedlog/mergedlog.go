package mergedlog

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LogFile struct {
	Alias          string
	Scanner        *bufio.Scanner
	Color          ColorFn
	RangeStart     int64
	RangeStop      int64
	grepRegex      *regexp.Regexp
	highlightRegex *regexp.Regexp
	index          int
	logChannel     chan *LogLine
	peek           *LogLine
	Format         string
}

type LogLine struct {
	Alias string
	UTime int64
	Text  LogEntry
	Color ColorFn
}

const MAX_INT = int64(^uint64(0) >> 1)

func (lf *LogFile) Peek() *LogLine {
	if lf.peek == nil {
		lf.Take()
	}
	return lf.peek
}

func (lf *LogFile) Take() *LogLine {
	taken := lf.peek
	lf.peek = <-lf.logChannel
	return taken
}

func (lf *LogFile) SetFormat(maxNameSize int) {
	lf.Format = "%" + strconv.Itoa(len(lf.Alias)-maxNameSize) + "s[%s] "
}

func (lf *LogFile) Process() {
	lineCount := 0
	var grepMatch []string
	var logChunk string

	for {
		if lf.Scanner.Scan() {
			logChunk = lf.Scanner.Text()

			matches := gfeLogLineRE.FindStringSubmatch(logChunk)
			if matches == nil {
				if lineCount == 0 || logChunk == "" {
					continue
				}
				// This should not happen since the [ScanLogEntries] function should bring us
				// a whole log entry chunk of text.
				panic(fmt.Sprintf("No match for pattern on logChunk: '%s' file: %s line: %d",
					logChunk, lf.Alias, lineCount))
			}

			logEntry := LogEntry{}

			foundGrep := false
			for _, line := range strings.Split(logChunk, "\n") {
				span := Span{}
				if lf.grepRegex != nil {
					grepMatch = lf.grepRegex.FindStringSubmatch(line)
					if grepMatch != nil {
						foundGrep = true
						span = append(span, grepMatch[1],
							lf.Color.Grep(grepMatch[2]),
							grepMatch[len(grepMatch)-1])
					} else {
						span = append(span, line)
					}
				} else {
					span = append(span, line)
				}

				logEntry = append(logEntry, span)
			}

			// If we're grepping but didn't find anything in the whole log entry then move on
			if lf.grepRegex != nil && !foundGrep {
				continue
			}

			lineCount++

			if lf.highlightRegex != nil {
				for j := 0; j < len(logEntry); j++ {
					for k := 0; k < len(logEntry[j]); k++ {
						span := logEntry[j]
						if s, ok := span[k].(string); ok {
							m := lf.highlightRegex.FindStringSubmatch(s)
							if m != nil {
								newSpan := Span{}
								for n := 0; n < k; n++ {
									newSpan = append(newSpan, span[n])
								}
								newSpan = append(newSpan, m[1], lf.Color.Highlight(m[2]), m[3])
								for n := k + 1; n < len(span); n++ {
									newSpan = append(newSpan, span[n])
								}
								logEntry[j] = newSpan
							}
						}
					}
				}
			}

			stamp := strings.TrimSpace(matches[1])
			t, err := time.Parse(STAMP_FORMAT, stamp)
			if err != nil {
				log.Printf("Unable to parse date stamp '%s': %s", stamp, err)
				continue
			}
			if t.UnixNano() < lf.RangeStart || lf.RangeStop < t.UnixNano() {
				continue
			}

			l := &LogLine{
				Alias: lf.Alias,
				UTime: t.UnixNano(),
				Text:  logEntry,
				Color: lf.Color,
			}

			lf.logChannel <- l

		} else if err := lf.Scanner.Err(); err != nil {
			log.Fatalf("error reading '%s': %s", lf.Alias, err.Error())
		} else {
			break
		}
	}

	endToken := &LogLine{
		UTime: MAX_INT,
	}
	lf.logChannel <- endToken
}

var endOfLine = regexp.MustCompile(`\n\[\w`)

func ScanLogEntries(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if loc := endOfLine.FindIndex(data); loc != nil {
		// We have a full newline-terminated line.
		return loc[0] + 1, bytes.TrimRight(data[0:loc[0]], "\r"), nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it,
	// dropping the trailing newline.
	if atEOF {
		return len(data), bytes.TrimRight(data[0:], "\r\n"), nil
	}

	// Request more data.
	return 0, nil, nil
}
