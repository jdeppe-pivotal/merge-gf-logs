package mergedlog

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
	waitGroup      *sync.WaitGroup
	index          int
}

type LogLine struct {
	Alias     string
	UTime     int64
	Text      LogEntry
	Color     ColorFn
	FileIndex int
}

type MergedLog struct {
	AggLog   *list.List
	lastLine *list.Element
	writer   io.Writer
}

const MAX_INT = int64(^uint64(0) >> 1)

func (lf *LogFile) Process(logChannel chan<- *LogLine) {
	lineCount := 0
	var grepMatch []string

	for {
		if lf.Scanner.Scan() {
			logChunk := lf.Scanner.Text()

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
							grepMatch[3])
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
				Alias:     lf.Alias,
				UTime:     t.UnixNano(),
				Text:      logEntry,
				Color:     lf.Color,
				FileIndex: lf.index,
			}

			logChannel <- l

		} else if err := lf.Scanner.Err(); err != nil {
			log.Fatalf("error reading '%s': %s", lf.Alias, err.Error())
		} else {
			break
		}
	}

	lf.waitGroup.Done()
}

func NewMergedLog() *MergedLog {
	return &MergedLog{
		AggLog: list.New(),
		writer: os.Stdout,
	}
}

func (ml *MergedLog) SetWriter(writer io.Writer) {
	ml.writer = writer
}

func (ml *MergedLog) Insert(line *LogLine) {
	if ml.lastLine == nil {
		if ml.AggLog.Len() == 0 {
			ml.lastLine = ml.AggLog.PushFront(line)
			return
		}
		ml.lastLine = ml.AggLog.Front()
	}

	var insertBefore *list.Element
	// Skip back if necessary
	for ml.lastLine != nil {
		x := ml.lastLine.Value.(*LogLine)
		if line.UTime >= x.UTime {
			break
		}
		insertBefore = ml.lastLine
		ml.lastLine = ml.lastLine.Prev()
	}

	for ; ml.lastLine != nil; ml.lastLine = ml.lastLine.Next() {
		x := ml.lastLine.Value.(*LogLine)
		if line.UTime < x.UTime {
			break
		}
	}

	if ml.lastLine == nil {
		if insertBefore != nil {
			ml.lastLine = ml.AggLog.InsertBefore(line, insertBefore)
		} else {
			ml.lastLine = ml.AggLog.PushBack(line)
		}
	} else {
		ml.lastLine = ml.AggLog.InsertBefore(line, ml.lastLine)
	}

	return
}

func (ml *MergedLog) FlushLogs(highestStamp int64, maxLogNameLength int) {
	for e := ml.AggLog.Front(); e != nil; e = ml.AggLog.Front() {
		entry, _ := e.Value.(*LogLine)
		if entry.UTime < highestStamp {
			format := "%" + strconv.Itoa(len(entry.Alias)-maxLogNameLength) + "s[%s] "

			for _, logEntry := range entry.Text {
				fmt.Fprintf(ml.writer, format, "", entry.Color.Normal(entry.Alias))
				for _, span := range logEntry {
					switch s := span.(type) {
					case Highlighted:
						fmt.Fprint(ml.writer, span)
					case string:
						fmt.Fprint(ml.writer, entry.Color.Normal(s))
					}
				}
				fmt.Fprintln(ml.writer)
			}

			ml.AggLog.Remove(e)
		} else {
			break
		}
	}
}

func ScanLogEntries(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, []byte("\n[")); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, bytes.TrimRight(data[0:i], "\r"), nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it,
	// dropping the trailing newline.
	if atEOF {
		return len(data), bytes.TrimRight(data[0:], "\r\n"), nil
	}

	// Request more data.
	return 0, nil, nil
}
