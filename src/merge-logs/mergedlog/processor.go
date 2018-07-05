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

	"github.com/mgutz/ansi"
)

type Processor struct {
	logs             LogCollection
	aggLog           *list.List
	rangeStart       int64
	rangeStop        int64
	maxLogNameLength int
	colorIndex       int
	palette          []string
	writer           io.Writer
}

var FLUSH_BATCH_SIZE = 1000
var resetColor string
var gfeLogLineRE = regexp.MustCompile(`^\[\w+ (([^ ]* ){3}).*`)
var StampFormat = "2006/01/02 15:04:05.000 MST"

func init() {
	resetColor = ansi.ColorCode("reset")
}

func NewProcessor(rangeStart, rangeStop int64) *Processor {
	processor := &Processor{}
	processor.aggLog = list.New()
	processor.rangeStart = rangeStart
	processor.rangeStop = rangeStop
	processor.logs = NewLogCollection()
	processor.palette = make([]string, 1)
	processor.palette[0] = ""
	processor.writer = os.Stdout

	return processor
}

func (this *Processor) AddLog(alias string, reader io.Reader, maxBuffer int) {
	logFile := LogFile{
		Alias:      alias,
		Scanner:    bufio.NewScanner(reader),
		AggLog:     this.aggLog,
		RangeStart: this.rangeStart,
		RangeStop:  this.rangeStop,
		Color:      this.palette[this.colorIndex],
	}

	logFile.Scanner.Split(ScanLogEntries)
	logFile.Scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), maxBuffer)

	this.logs = append(this.logs, logFile)
	this.colorIndex = (this.colorIndex + 1) % len(this.palette)

	if len(logFile.Alias) > this.maxLogNameLength {
		this.maxLogNameLength = len(logFile.Alias)
	}
}

func (this *Processor) Crank() {
	var oldestStampSeen = MAX_INT
	var lastTimeRead int64
	lineCount := 0

	for len(this.logs) > 0 {
		// Process log list backwards so that we can delete entries as necessary
		for i := len(this.logs) - 1; i >= 0; i-- {
			if this.logs[i].Scanner.Scan() {
				lineCount++
				line := this.logs[i].Scanner.Text()
				matches := gfeLogLineRE.FindStringSubmatch(line)

				if matches != nil {
					stamp := strings.TrimSpace(matches[1])
					t, err := time.Parse(StampFormat, stamp)
					if err != nil {
						log.Printf("Unable to parse date stamp '%s': %s", stamp, err)
						continue
					}
					lastTimeRead = t.UnixNano()

					l := &LogLine{Alias: this.logs[i].Alias, UTime: t.UnixNano(), Text: line, Color: this.logs[i].Color}
					this.logs[i].Insert(l)
				} else {
					if x := this.logs[i].InsertTimeless(line); x != nil {
						v, _ := x.Value.(*LogLine)
						lastTimeRead = v.UTime
					}
				}

				if lastTimeRead < oldestStampSeen {
					oldestStampSeen = lastTimeRead
				}

				//fmt.Printf("---- dumping ===> %s\n", line)
				//this.Dump()

			} else if err := this.logs[i].Scanner.Err(); err != nil {
				log.Fatalf("error reading '%s': %s", this.logs[i].Alias, err.Error())
			} else {
				this.logs = append(this.logs[:i], this.logs[i+1:]...)
			}
		}

		if lineCount%FLUSH_BATCH_SIZE == 0 {
			this.flushLogs(oldestStampSeen, this.aggLog, this.maxLogNameLength)
			oldestStampSeen = MAX_INT
		}
	}

	this.flushLogs(MAX_INT, this.aggLog, this.maxLogNameLength)
}

func (this *Processor) SetPalette(palette []string) {
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
	var reset string
	for e := aggLog.Front(); e != nil; e = aggLog.Front() {
		entry, _ := e.Value.(*LogLine)
		if entry.UTime < highestStamp {
			format := "%s%" + strconv.Itoa(len(entry.Alias)-maxLogNameLength) + "s[%s] %s%s\n"

			if entry.Color != "" {
				reset = resetColor
			} else {
				reset = ""
			}

			for _, line := range strings.Split(entry.Text, "\n") {
				fmt.Fprintf(this.writer, format, entry.Color, "", entry.Alias, line, reset)
			}
			aggLog.Remove(e)
		} else {
			break
		}
	}
}
