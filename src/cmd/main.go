package main

import (
	"flag"
	"mergedlog"
	"container/list"
	"regexp"
	"log"
	"strings"
	"os"
	"bufio"
	"time"
	"fmt"
)

var MAX_INT int64 = int64(^uint64(0) >> 1)
var FLUSH_BATCH_SIZE int = 1000

var stampFormat = "2006/01/02 15:04:05.000 MST"
var logs mergedlog.LogCollection
var aggLog *list.List
var lineCount int = 0

func main() {
	flag.Parse()

	gfeLogLineRE, err := regexp.Compile(`^\[\w+ (([^ ]* ){3}).*`)
	if err != nil {
		log.Fatalf("Invalid regex: %s", err)
	}

	aggLog = list.New()

	var logName, alias string
	// Gather our files and set up a Scanner for each of them
	for _, logTagName := range (flag.Args()) {
		parts := strings.Split(logTagName, ":")
		alias = parts[0]

		// See if we have an alias for the log
		if len(parts) == 1 {
			logName = parts[0]
			bits := strings.Split(logName, "/")
			alias = bits[len(bits) - 1]
		} else {
			logName = parts[1]
		}

		f, err := os.Open(logName)
		if err != nil {
			log.Fatalf("Error opening file: %s", err)
		}
		defer f.Close()

		logs = append(logs, mergedlog.LogFile{Alias: alias, Scanner: bufio.NewScanner(f), AggLog: aggLog})
	}

	var oldestStampSeen int64 = MAX_INT
	var lastTimeRead int64

	for len(logs) > 0 {
		// Process log list backwards so that we can delete entries as necessary
		for i := len(logs) - 1; i >= 0; i-- {
			if logs[i].Scanner.Scan() {
				lineCount++
				line := logs[i].Scanner.Text()
				matches := gfeLogLineRE.FindStringSubmatch(line)

				if matches != nil {
					stamp := strings.TrimSpace(matches[1])
					t, err := time.Parse(stampFormat, stamp)
					if err != nil {
						log.Printf("Unable to parse date stamp '%s': %s", stamp, err)
						continue
					}
					lastTimeRead = t.UnixNano()

					l := &mergedlog.LogLine{Alias: logs[i].Alias, UTime: t.UnixNano(), Text: line}
					logs[i].Insert(l)
				} else {
					x := logs[i].InsertTimeless(line)
					v, _ := x.Value.(*mergedlog.LogLine)
					lastTimeRead = v.UTime
				}

				if lastTimeRead < oldestStampSeen {
					oldestStampSeen = lastTimeRead
				}

			} else {
				logs = append(logs[:i], logs[i + 1:]...)
			}
		}

		if lineCount % FLUSH_BATCH_SIZE == 0 {
			flushLogs(oldestStampSeen, aggLog)
			oldestStampSeen = MAX_INT
		}
	}

	flushLogs(MAX_INT, aggLog)
}

func flushLogs(highestStamp int64, aggLog *list.List) {
	for e := aggLog.Front(); e != nil; e = aggLog.Front() {
		line, _ := e.Value.(*mergedlog.LogLine)
		if line.UTime < highestStamp {
			fmt.Printf("[%s] %s\n", line.Alias, line.Text)
			aggLog.Remove(e)
		} else {
			break
		}
	}
}
