package mergedlog

import (
	"bufio"
	"container/list"
    "fmt"
)

type LogFile struct {
	Alias    string
	Scanner  *bufio.Scanner
	AggLog   *list.List
	lastLine *list.Element
}

type LogLine struct {
	Alias string
	UTime int64
	Text  string
}

type LogCollection []LogFile

func NewLogCollection(size int) LogCollection {
	return make([]LogFile, size)
}

func (this *LogCollection) AddLog(file *LogFile) {
	*this = append(*this, *file)
}

func (this *LogFile) Insert(line *LogLine) *list.Element {
	line.Alias = this.Alias

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
		last, _ := this.lastLine.Value.(*LogLine)
		l := &LogLine{Alias: last.Alias, UTime: last.UTime, Text: line}
		this.lastLine = this.AggLog.InsertAfter(l, this.lastLine)
	} else {
		l := &LogLine{Alias: this.Alias, UTime: 0, Text: line}
		this.lastLine = this.AggLog.PushFront(l)
	}

	return this.lastLine
}

func Dump(agg *list.List) {
	for v := agg.Front(); v != nil; v = v.Next() {
		x, _ := v.Value.(*LogLine)
		fmt.Printf(">>> %+v\n", x)
	}
}
