package mergedlog

import (
	"bufio"
	"bytes"
	"container/list"
)

type LogFile struct {
	Alias      string
	Scanner    *bufio.Scanner
	AggLog     *list.List
	Color      string
	lastLine   *list.Element
	RangeStart int64
	RangeStop  int64
}

type LogLine struct {
	Alias string
	UTime int64
	Text  string
	Color string
}

type LogCollection []LogFile

const MAX_INT = int64(^uint64(0) >> 1)

func NewLogCollection() LogCollection {
	return make([]LogFile, 0)
}

func (this *LogCollection) AddLogs(file *LogFile) {
	*this = append(*this, *file)
}

func (this *LogFile) Insert(line *LogLine) *list.Element {
	line.Alias = this.Alias

	if line.UTime < this.RangeStart || this.RangeStop < line.UTime {
		return nil
	}

	if this.lastLine == nil {
		if this.AggLog.Len() == 0 {
			this.lastLine = this.AggLog.PushFront(line)
			return this.lastLine
		}
		this.lastLine = this.AggLog.Front()
	}

	for ; this.lastLine != nil; this.lastLine = this.lastLine.Next() {
		x := this.lastLine.Value.(*LogLine)
		if line.UTime < x.UTime {
			break
		}
	}

	if this.lastLine == nil {
		this.lastLine = this.AggLog.PushBack(line)
	} else {
		this.lastLine = this.AggLog.InsertBefore(line, this.lastLine)
	}

	return this.lastLine
}

func (this *LogFile) InsertTimeless(line string) *list.Element {
	if this.lastLine != nil {
		// this should not happen with the custom scan function
		// TODO return element and error
		last, _ := this.lastLine.Value.(*LogLine)
		l := &LogLine{Alias: last.Alias, UTime: last.UTime, Text: line, Color: last.Color}
		this.lastLine = this.AggLog.InsertAfter(l, this.lastLine)
		return this.lastLine
	} else if this.RangeStart == 0 {
		l := &LogLine{Alias: this.Alias, UTime: 0, Text: line, Color: this.Color}
		this.lastLine = this.AggLog.PushFront(l)

		return this.lastLine
	}

	return nil
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
		return len(data), bytes.TrimRight(data[0:len(data)], "\r\n"), nil
	}

	// Request more data.
	return 0, nil, nil
}
