package mergedlog

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
)

type Processor struct {
	writer           io.Writer
	logFiles         []*LogFile
	rangeStart       int64
	rangeStop        int64
	maxLogNameLength int
	colorIndex       int
	aliasColorMap    map[string]int
	palette          []ColorFn
	debugLevel       int
	grepRegex        *regexp.Regexp
	highlightRegex   *regexp.Regexp
	FileCount        int
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

var gfeLogLineRE = regexp.MustCompile(`^\[\w+ (([^ ]* ){3}).*`)

const STAMP_FORMAT = "2006/01/02 15:04:05.000 MST"

func NewProcessor(rangeStart, rangeStop int64, grepRegex, highlightRegex *regexp.Regexp, debugLevel int) *Processor {
	processor := &Processor{}
	processor.logFiles = make([]*LogFile, 0)
	processor.rangeStart = rangeStart
	processor.rangeStop = rangeStop
	processor.aliasColorMap = make(map[string]int)
	processor.palette = make([]ColorFn, 1)
	processor.palette[0] = MakePaletteEntry("234")
	processor.debugLevel = debugLevel
	processor.grepRegex = grepRegex
	processor.highlightRegex = highlightRegex

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
		Alias:          aliasAndRoll,
		Scanner:        bufio.NewScanner(reader),
		RangeStart:     this.rangeStart,
		RangeStop:      this.rangeStop,
		Color:          this.palette[this.aliasColorMap[alias]],
		grepRegex:      this.grepRegex,
		highlightRegex: this.highlightRegex,
		index:          this.FileCount,
		logChannel:     make(chan *LogLine, 100),
	}
	this.FileCount++

	logFile.Scanner.Split(ScanLogEntries)
	logFile.Scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), maxBuffer)

	if len(logFile.Alias) > this.maxLogNameLength {
		this.maxLogNameLength = len(logFile.Alias)
	}

	this.logFiles = append(this.logFiles, &logFile)
	go logFile.Process()
}

func (p *Processor) SetFormat(maxNameLen int) {
	for _, logFile := range p.logFiles {
		logFile.SetFormat(maxNameLen)
	}
}

func (this *Processor) Crank() {
	var oldestTimestamp int64
	var idx int
	var line *LogLine

	for {
		idx = -1
		oldestTimestamp = MAX_INT
		for i, logFile := range this.logFiles {
			if logFile.Peek().UTime < oldestTimestamp {
				idx = i
				oldestTimestamp = logFile.Peek().UTime
			}
		}

		if idx < 0 {
			break
		}

		line = this.logFiles[idx].Take()

		for _, logEntry := range line.Text {
			fmt.Fprintf(this.writer, this.logFiles[idx].Format, "", line.Color.Normal(line.Alias))
			for _, span := range logEntry {
				switch s := span.(type) {
				case Highlighted:
					fmt.Fprint(this.writer, span)
				case string:
					fmt.Fprint(this.writer, line.Color.Normal(s))
				}
			}
			fmt.Fprintln(this.writer)
		}
	}

	if w, ok := this.writer.(*bufio.Writer); ok {
		w.Flush()
	}
}

func (this *Processor) SetPalette(palette []ColorFn) {
	this.palette = palette
}

func (this *Processor) SetWriter(writer io.Writer) {
	this.writer = writer
}
