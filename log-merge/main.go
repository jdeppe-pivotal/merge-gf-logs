package main

import (
	"flag"
	"os"
	"log"
    "fmt"
	"bufio"
    "regexp"
	"strings"
	"time"
	"container/list"
	"runtime/pprof"
)

type LogFile struct {
	alias    string
	scanner  *bufio.Scanner
	lastLine *list.Element
}

type LogLine struct {
	alias string
	uTime int64
	text  string
}

var stampFormat = "2006/01/02 15:04:05.000 MST"
var cpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")
var logs []LogFile
var aggLog *list.List

func main() {
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	gfeLogLineRE, err := regexp.Compile(`^\[\w+ (([^ ]* ){3}).*`)
	if err != nil {
		log.Fatalf("Invalid regex: %s", err)
	}

	aggLog = list.New()

	var logName, alias string
	// Gather our files and set up a Scanner for each of them
	for _, logTagName := range(flag.Args()) {
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

		logs = append(logs, LogFile{alias: alias, scanner: bufio.NewScanner(f)})
	}

	for len(logs) > 0 {
		// Process log list backwards so that we can delete entries as necessary
		for i := len(logs) - 1; i >= 0; i-- {
			if logs[i].scanner.Scan() {
				line := logs[i].scanner.Text()
				matches := gfeLogLineRE.FindStringSubmatch(line)

				if matches != nil {
					stamp := strings.TrimSpace(matches[1])
					t, err := time.Parse(stampFormat, stamp)
					if err != nil {
						log.Printf("Unable to parse date stamp '%s': %s", stamp, err)
						continue
					}

					l := &LogLine{alias: logs[i].alias, uTime: t.UnixNano(), text: line}
					if aggLog.Len() == 0 {
						aggLog.PushFront(l)
					} else {
						e := aggLog.Back()
						for ; e != nil; e = e.Prev() {
							if x, ok := e.Value.(*LogLine); ok {
								if l.uTime >= x.uTime {
									logs[i].lastLine = aggLog.InsertAfter(l, e)
									break
								}
							}
						}
						if e == nil {
							logs[i].lastLine = aggLog.PushFront(l)
						}
					}
				} else {
					if logs[i].lastLine != nil {
						prev, _ := logs[i].lastLine.Value.(*LogLine)
						l := &LogLine{alias: logs[i].alias, uTime: prev.uTime, text: line}
						logs[i].lastLine = aggLog.InsertAfter(l, logs[i].lastLine)
					} else {
						l := &LogLine{alias: logs[i].alias, uTime: 0, text: line}
						logs[i].lastLine = aggLog.PushFront(l)
					}
				}
			} else {
				logs = append(logs[:i], logs[i + 1:]...)
			}
		}
	}
	dump(aggLog)
}

func dump(aggLog *list.List) {
	for e := aggLog.Front(); e != nil; e = e.Next() {
		line, _ := e.Value.(*LogLine)
		fmt.Printf("[%s] %s\n", line.alias, line.text)
	}
}


