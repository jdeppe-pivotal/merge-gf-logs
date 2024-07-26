package mergedlog

import (
	"bufio"
	"io"
	"regexp"
	"sync"
)

type Processor struct {
	aggLog           *MergedLog
	rangeStart       int64
	rangeStop        int64
	maxLogNameLength int
	colorIndex       int
	aliasColorMap    map[string]int
	palette          []ColorFn
	debugLevel       int
	grepRegex        *regexp.Regexp
	highlightRegex   *regexp.Regexp
	logChannel       chan *LogLine
	waitGroup        *sync.WaitGroup
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

var FLUSH_BATCH_SIZE = 20
var gfeLogLineRE = regexp.MustCompile(`^\[\w+ (([^ ]* ){3}).*`)

const STAMP_FORMAT = "2006/01/02 15:04:05.000 MST"

func NewProcessor(logChannnel chan *LogLine, rangeStart, rangeStop int64, grepRegex, highlightRegex *regexp.Regexp, debugLevel int) *Processor {
	processor := &Processor{}
	processor.aggLog = NewMergedLog()
	processor.rangeStart = rangeStart
	processor.rangeStop = rangeStop
	processor.aliasColorMap = make(map[string]int)
	processor.palette = make([]ColorFn, 1)
	processor.palette[0] = MakePaletteEntry("234")
	processor.debugLevel = debugLevel
	processor.grepRegex = grepRegex
	processor.highlightRegex = highlightRegex
	processor.logChannel = logChannnel
	processor.waitGroup = &sync.WaitGroup{}

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
		waitGroup:      this.waitGroup,
		index:          this.FileCount,
	}
	this.FileCount++

	logFile.Scanner.Split(ScanLogEntries)
	logFile.Scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), maxBuffer)

	if len(logFile.Alias) > this.maxLogNameLength {
		this.maxLogNameLength = len(logFile.Alias)
	}

	this.waitGroup.Add(1)
	go logFile.Process(this.logChannel)
}

func (this *Processor) Crank() {
	var lineCount = 0
	var filesCompleted = 0
	logsSeen := make([]int64, this.FileCount)

	go func() {
		this.waitGroup.Wait()
		close(this.logChannel)
	}()

	for line := range this.logChannel {
		if line.FileIndex == -1 {
			filesCompleted += 1
			continue
		}

		lineCount++
		this.aggLog.Insert(line)
		logsSeen[line.FileIndex] = line.UTime

		if lineCount > FLUSH_BATCH_SIZE {
			seenAllLogs := 0
			oldestStampSeen := MAX_INT
			for i, _ := range logsSeen {
				if logsSeen[i] > 0 {
					seenAllLogs++
					if logsSeen[i] < oldestStampSeen {
						oldestStampSeen = logsSeen[i]
					}
				}
			}
			if seenAllLogs+filesCompleted < this.FileCount {
				continue
			}

			this.aggLog.FlushLogs(oldestStampSeen, this.maxLogNameLength)
			oldestStampSeen = MAX_INT
			lineCount = 0
			for i, _ := range logsSeen {
				logsSeen[i] = 0
			}
		}
	}

	this.aggLog.FlushLogs(MAX_INT, this.maxLogNameLength)
}

func (this *Processor) SetPalette(palette []ColorFn) {
	this.palette = palette
}

func (this *Processor) SetWriter(writer io.Writer) {
	this.aggLog.SetWriter(writer)
}
