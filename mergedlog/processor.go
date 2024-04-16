package mergedlog

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Processor struct {
	logs             LogCollection
	aggLog           *list.List
	rangeStart       int64
	rangeStop        int64
	maxLogNameLength int
	colorIndex       int
	aliasColorMap    map[string]int
	palette          []ColorFn
	writer           io.Writer
	debugLevel       int
	grepRegex        *regexp.Regexp
	highlightRegex   *regexp.Regexp
}

type ColorFn struct {
	Normal    func(string) Highlighted
	Grep      func(string) Highlighted
	Highlight func(string) Highlighted
}

// Highlighted indicates a string that has been marked up with ANSI codes
type Highlighted string

// Span is a slice of either *string or [Highlighted] values and represents a line of text
type Span []any

type LogEntry []Span

var FLUSH_BATCH_SIZE = 20
var gfeLogLineRE = regexp.MustCompile(`^\[\w+ (([^ ]* ){3}).*`)
var StampFormat = "2006/01/02 15:04:05.000 MST"

func NewProcessor(rangeStart, rangeStop int64, grepRegex, highlightRegex *regexp.Regexp, debugLevel int) *Processor {
	processor := &Processor{}
	processor.logs = NewLogCollection()
	processor.aggLog = list.New()
	processor.rangeStart = rangeStart
	processor.rangeStop = rangeStop
	processor.aliasColorMap = make(map[string]int)
	processor.palette = make([]ColorFn, 1)
	processor.palette[0] = MakePaletteEntry("234")
	processor.debugLevel = debugLevel
	processor.grepRegex = grepRegex
	processor.highlightRegex = highlightRegex
	processor.writer = os.Stdout

	return processor
}

func (this *Processor) AddLog(alias string, rolled bool, reader io.Reader, maxBuffer int) {
	if _, ok := this.aliasColorMap[alias]; !ok {
		this.aliasColorMap[alias] = this.colorIndex
		this.colorIndex = (this.colorIndex + 1) % len(this.palette)
	}

	aliasAndRoll := alias
	if rolled {
		aliasAndRoll += "*"
	}

	logFile := LogFile{
		Alias:      aliasAndRoll,
		Scanner:    bufio.NewScanner(reader),
		AggLog:     this.aggLog,
		RangeStart: this.rangeStart,
		RangeStop:  this.rangeStop,
		Color:      this.palette[this.aliasColorMap[alias]],
	}

	logFile.Scanner.Split(ScanLogEntries)
	logFile.Scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), maxBuffer)

	this.logs = append(this.logs, logFile)

	if len(logFile.Alias) > this.maxLogNameLength {
		this.maxLogNameLength = len(logFile.Alias)
	}
}

func (this *Processor) Crank() {
	var oldestStampSeen = MAX_INT
	var lastTimeRead int64
	var grepMatch []string
	lineCount := 0
	var linesSeenPerLoop int

	for len(this.logs) > 0 {
		linesSeenPerLoop = 0
		// Process log list backwards so that we can delete entries as necessary
		for i := len(this.logs) - 1; i >= 0; i-- {
			if this.logs[i].Scanner.Scan() {
				logChunk := this.logs[i].Scanner.Text()

				matches := gfeLogLineRE.FindStringSubmatch(logChunk)
				if matches == nil {
					if lineCount == 0 || logChunk == "" {
						continue
					}
					// This should not happen since the [ScanLogEntries] function should bring us
					// a whole log entry chunk of text.
					panic(fmt.Sprintf("No match for pattern on logChunk: '%s' file: %s line: %d",
						logChunk, this.logs[i].Alias, lineCount))
				}

				logEntry := LogEntry{}

				foundGrep := false
				for _, line := range strings.Split(logChunk, "\n") {
					span := Span{}
					if this.grepRegex != nil {
						grepMatch = this.grepRegex.FindStringSubmatch(line)
						if grepMatch != nil {
							foundGrep = true
							span = append(span, grepMatch[1],
								this.logs[i].Color.Grep(grepMatch[2]),
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
				if this.grepRegex != nil && !foundGrep {
					continue
				}

				lineCount++

				if this.highlightRegex != nil {
					for j := 0; j < len(logEntry); j++ {
						for k := 0; k < len(logEntry[j]); k++ {
							span := logEntry[j]
							if s, ok := span[k].(string); ok {
								m := this.highlightRegex.FindStringSubmatch(s)
								if m != nil {
									newSpan := Span{}
									for n := 0; n < k; n++ {
										newSpan = append(newSpan, span[n])
									}
									newSpan = append(newSpan, m[1], this.logs[i].Color.Highlight(m[2]), m[3])
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
				t, err := time.Parse(StampFormat, stamp)
				if err != nil {
					log.Printf("Unable to parse date stamp '%s': %s", stamp, err)
					continue
				}
				lastTimeRead = t.UnixNano()

				l := &LogLine{Alias: this.logs[i].Alias, UTime: t.UnixNano(), Text: logEntry, Color: this.logs[i].Color}
				this.logs[i].Insert(l)
				linesSeenPerLoop++

				if lastTimeRead < oldestStampSeen {
					oldestStampSeen = lastTimeRead
				}

				if this.debugLevel > 0 {
					if this.debugLevel >= 1 {
						fmt.Printf("---- DEBUG oldestStampSeen: %d\n", oldestStampSeen)
						fmt.Printf("---- DEBUG ===> %v\n", l)
					}
					if this.debugLevel > 1 {
						this.Dump()
					}
				}

			} else if err := this.logs[i].Scanner.Err(); err != nil {
				log.Fatalf("error reading '%s': %s", this.logs[i].Alias, err.Error())
			} else {
				this.logs = append(this.logs[:i], this.logs[i+1:]...)
			}
		}

		if lineCount > FLUSH_BATCH_SIZE && linesSeenPerLoop == len(this.logs) {
			if this.debugLevel > 0 {
				fmt.Printf("---- DEBUG flushing logs\n")
			}
			this.flushLogs(oldestStampSeen, this.aggLog, this.maxLogNameLength)
			oldestStampSeen = MAX_INT
			lineCount = 0
		}
	}

	this.flushLogs(MAX_INT, this.aggLog, this.maxLogNameLength)
}

func (this *Processor) SetPalette(palette []ColorFn) {
	this.palette = palette
}

func (this *Processor) SetWriter(writer io.Writer) {
	this.writer = writer
}

func (this *Processor) Dump() {
	for v := this.aggLog.Front(); v != nil; v = v.Next() {
		x, _ := v.Value.(*LogLine)
		fmt.Printf(">>> %+v\n", x)
	}
}

func (this *Processor) flushLogs(highestStamp int64, aggLog *list.List, maxLogNameLength int) {
	for e := aggLog.Front(); e != nil; e = aggLog.Front() {
		entry, _ := e.Value.(*LogLine)
		if entry.UTime < highestStamp {
			format := "%" + strconv.Itoa(len(entry.Alias)-maxLogNameLength) + "s[%s] "

			for _, logEntry := range entry.Text {
				fmt.Fprintf(this.writer, format, "", entry.Color.Normal(entry.Alias))
				for _, span := range logEntry {
					switch s := span.(type) {
					case Highlighted:
						fmt.Fprint(this.writer, span)
					case string:
						fmt.Fprint(this.writer, entry.Color.Normal(s))
					}

				}
				fmt.Fprintln(this.writer)
			}

			aggLog.Remove(e)
		} else {
			break
		}
	}
}
