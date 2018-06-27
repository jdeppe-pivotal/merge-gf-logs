package main

import (
	"flag"
	"log"
	"merge-logs/mergedlog"
	"os"
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
	duration := flag.Int64("duration", mergedlog.MAX_INT, "start of range of logs")
	rangeStopStr := flag.String("stop", "", "start of range of logs")
	flag.Parse()

	rangeStop := mergedlog.MAX_INT
	if *rangeStopStr != "" {
		t, err := time.Parse(mergedlog.StampFormat, *rangeStopStr)
		if err != nil {
			log.Fatalf("Unable to parse '%s' as timestamp", rangeStopStr)
		}
		rangeStop = t.UnixNano()
		// if duration is larger than the stop time, adjust it so that start
		// time is positive
		if *duration > t.Unix() {
			*duration = t.Unix()
		}
	}

	rangeStart := rangeStop - int64(time.Duration(*duration)*time.Second)
	processor := mergedlog.NewProcessor(rangeStart, rangeStop)

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
			bits := strings.Split(logName, "/")
			alias = bits[len(bits)-1]
		} else {
			logName = parts[1]
		}

		f, err := os.Open(logName)
		if err != nil {
			log.Fatalf("Error opening file: %s", err)
		}
		defer f.Close()

		processor.AddLog(alias, f)
	}

	processor.Crank()
}
