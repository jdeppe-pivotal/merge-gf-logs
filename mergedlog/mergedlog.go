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

func (lf *LogFile) Insert(line *LogLine) *list.Element {
	line.Alias = lf.Alias

	if line.UTime < lf.RangeStart || lf.RangeStop < line.UTime {
		return nil
	}

	if lf.lastLine == nil {
		if lf.AggLog.Len() == 0 {
			lf.lastLine = lf.AggLog.PushFront(line)
			return lf.lastLine
		}
		lf.lastLine = lf.AggLog.Front()
	}

	var insertBefore *list.Element
	// Skip back if necessary
	for ; lf.lastLine != nil;  {
		x := lf.lastLine.Value.(*LogLine)
		if line.UTime >= x.UTime {
			break
		}
		insertBefore = lf.lastLine
		lf.lastLine = lf.lastLine.Prev()
	}

	for ; lf.lastLine != nil; lf.lastLine = lf.lastLine.Next() {
		x := lf.lastLine.Value.(*LogLine)
		if line.UTime < x.UTime {
			break
		}
	}

	if lf.lastLine == nil {
		if insertBefore != nil {
			lf.lastLine = lf.AggLog.InsertBefore(line, insertBefore)
		} else {
			lf.lastLine = lf.AggLog.PushBack(line)
		}
	} else {
		lf.lastLine = lf.AggLog.InsertBefore(line, lf.lastLine)
	}

	return lf.lastLine
}

func (lf *LogFile) InsertTimeless(line string) *list.Element {
	if lf.lastLine != nil {
		// lf should not happen with the custom scan function
		// TODO return element and error
		last, _ := lf.lastLine.Value.(*LogLine)
		l := &LogLine{Alias: last.Alias, UTime: last.UTime, Text: line, Color: last.Color}
		lf.lastLine = lf.AggLog.InsertAfter(l, lf.lastLine)
		return lf.lastLine
	} else if lf.RangeStart == 0 {
		l := &LogLine{Alias: lf.Alias, UTime: 0, Text: line, Color: lf.Color}
		lf.lastLine = lf.AggLog.PushFront(l)

		return lf.lastLine
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
