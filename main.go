package main

import (
	"flag"
	"fmt"
	"log"
	"merge-logs/mergedlog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mgutz/ansi"
)

var userColor string
var palette []string

func init() {
	// Taken from the Solarized color palette
	palette = make([]string, 8)
	palette[0] = ansi.ColorCode("253") // whiteish
	palette[1] = ansi.ColorCode("64")  // green
	palette[2] = ansi.ColorCode("37")  // cyan
	palette[3] = ansi.ColorCode("33")  // blue
	palette[4] = ansi.ColorCode("61")  // violet
	palette[5] = ansi.ColorCode("125") // magenta
	palette[6] = ansi.ColorCode("160") // red
	palette[7] = ansi.ColorCode("166") // orange
}

func main() {
	flag.StringVar(&userColor, "color", "dark", "Color scheme to use: light, dark or off")
	duration := flag.Int64("duration", mergedlog.MAX_INT, "duration (in seconds), relative to start or stop, to display")
	maxBuffer := flag.Int("max-buffer", 1024*1024, "maximum size of buffer to use when scanning")
	rangeStartStr := flag.String("start", "", "start timestamp of range of logs. Format: '2018/01/25 19:09:36.949 UTC'")
	rangeStopStr := flag.String("stop", "", "end timestamp of range of logs. Format: '2018/01/25 19:09:36.949 UTC'")
	debugLevel := flag.Int("debug", 0, "debug level - 0=off 1=verbose 2=very verbose")
	fullAlias := flag.Bool("full-alias", false, "use the full name as alias")

	flag.Parse()

	var rangeStart int64 = 0
	var rangeStop = mergedlog.MAX_INT

	if *rangeStartStr != "" {
		t, err := time.Parse(mergedlog.StampFormat, *rangeStartStr)
		if err != nil {
			log.Fatalf("Unable to parse '%s' as timestamp", rangeStartStr)
		}
		rangeStart = t.UnixNano()

		if *rangeStopStr == "" {
			if *duration == mergedlog.MAX_INT {
				rangeStop = mergedlog.MAX_INT
			} else {
				rangeStop = rangeStart + int64(time.Duration(*duration)*time.Second)
			}
		}
	}

	if *rangeStopStr != "" {
		t, err := time.Parse(mergedlog.StampFormat, *rangeStopStr)
		if err != nil {
			log.Fatalf("Unable to parse '%s' as timestamp", rangeStopStr)
		}
		rangeStop = t.UnixNano()

		if *rangeStartStr == "" {
			if *duration == mergedlog.MAX_INT {
				rangeStart = 0
			} else {
				rangeStart = rangeStop - int64(time.Duration(*duration)*time.Second)
			}
		}
	}

	if *debugLevel > 0 {
		fmt.Printf("---- DEBUG ===> rangeStart: %v\n", rangeStart)
		fmt.Printf("---- DEBUG ===> rangeStop: %v\n", rangeStop)
		fmt.Printf("---- DEBUG ===> calculated duration: %v\n", rangeStop-rangeStart)
		fmt.Printf("---- DEBUG ===> duration: %v\n", *duration)
		fmt.Printf("---- DEBUG ===> maxInt: %v\n", mergedlog.MAX_INT)
	}

	processor := mergedlog.NewProcessor(rangeStart, rangeStop, debugLevel)

	if userColor != "off" {
		if userColor == "light" {
			palette[0] = ansi.ColorCode("234") // blackish
		}
		processor.SetPalette(palette)
	}

	var logName, alias string
	// Gather our files and set up a Scanner for each of them
	for _, logTagName := range flag.Args() {
		parts := strings.Split(logTagName, ":")
		alias = parts[0]

		// See if we have an alias for the log
		if len(parts) == 1 {
			logName = parts[0]
			if !*fullAlias {
				alias = filepath.Base(logName)
			}
		} else {
			logName = parts[1]
		}

		f, err := os.Open(logName)
		if err != nil {
			log.Fatalf("Error opening file: %s", err)
		}
		defer f.Close()

		processor.AddLog(alias, f, *maxBuffer)
	}

	processor.Crank()
}
