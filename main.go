package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"log"
	"merge-logs/mergedlog"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var userColor string
var palette []mergedlog.ColorFn

func init() {
	// Taken from the Solarized color palette
	palette = make([]mergedlog.ColorFn, 8)
	palette[0] = mergedlog.MakePaletteEntry("253") // whiteish
	palette[1] = mergedlog.MakePaletteEntry("64")  // green
	palette[2] = mergedlog.MakePaletteEntry("37")  // cyan
	palette[3] = mergedlog.MakePaletteEntry("33")  // blue
	palette[4] = mergedlog.MakePaletteEntry("61")  // violet
	palette[5] = mergedlog.MakePaletteEntry("125") // magenta
	palette[6] = mergedlog.MakePaletteEntry("160") // red
	palette[7] = mergedlog.MakePaletteEntry("166") // orange
}

func main() {
	flag.StringVar(&userColor, "color", "dark", "ColorFn scheme to use: light, dark or off")
	duration := flag.Int64("duration", mergedlog.MAX_INT, "duration (in seconds), relative to start or stop, to display")
	maxBuffer := flag.Int("max-buffer", 1024*1024, "maximum size of buffer to use when scanning")
	rangeStartStr := flag.String("start", "", "start timestamp of range of logs. Format: '2018/01/25 19:09:36.949 UTC'")
	rangeStopStr := flag.String("stop", "", "end timestamp of range of logs. Format: '2018/01/25 19:09:36.949 UTC'")
	debugLevel := flag.Int("debug", 0, "debug level - 0=off 1=verbose 2=very verbose")
	fullAlias := flag.Bool("full-alias", false, "use the full name as alias")
	noLogRoll := flag.Bool("no-roll", false, "do not attempt to use the log rolling suffix numbers to associate different files with the same system (color)")
	grep := flag.StringP("grep", "g", "", "only process and display lines containing the regex")
	highlight := flag.StringP("highlight", "h", "", "highlight text that matches the regex")

	flag.Parse()

	var rangeStart int64 = 0
	var rangeStop = mergedlog.MAX_INT

	if *rangeStartStr != "" {
		t, err := time.Parse(mergedlog.STAMP_FORMAT, *rangeStartStr)
		if err != nil {
			log.Fatalf("Unable to parse '%s' as timestamp", *rangeStartStr)
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
		t, err := time.Parse(mergedlog.STAMP_FORMAT, *rangeStopStr)
		if err != nil {
			log.Fatalf("Unable to parse '%s' as timestamp", *rangeStopStr)
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

	var grepRegex *regexp.Regexp
	if *grep != "" {
		grepRegex = mergedlog.MakeGrepRegex(*grep)
	}

	var highlightRegex *regexp.Regexp
	if *highlight != "" {
		highlightRegex = regexp.MustCompile("(.*)(" + *highlight + ")(.*)")
	}

	logChannel := make(chan *mergedlog.LogLine)
	processor := mergedlog.NewProcessor(logChannel, rangeStart, rangeStop, grepRegex, highlightRegex, *debugLevel)

	if userColor != "off" {
		if userColor == "light" {
			// blackish
			palette[0] = mergedlog.MakePaletteEntry("234")
		}
		processor.SetPalette(palette)
	}

	// Use an array so that we get consistent ordering of the files and thus consistent coloring across runs
	filenameList := make([]string, len(flag.Args()))
	fullToShort := make(map[string]string)
	shortToTag := make(map[string]*string)
	// Process the log filenames and potentially group them
	for i, logTagName := range flag.Args() {
		fullName, shortName, tag := mergedlog.ProcessFilename(logTagName, *fullAlias)
		filenameList[i] = fullName
		var maybeShort string
		if *noLogRoll {
			maybeShort = fullName
		} else {
			maybeShort = shortName
		}
		fullToShort[fullName] = maybeShort
		if _, present := shortToTag[maybeShort]; !present && tag != nil {
			shortToTag[maybeShort] = tag
		}
	}

	var alias string
	// Gather our files and set up a Scanner for each of them
	for _, full := range filenameList {
		var rolled bool
		short := fullToShort[full]
		if tag, ok := shortToTag[short]; ok {
			alias = *tag
		} else {
			if *fullAlias {
				alias = full
			} else {
				alias = short
			}

			if filepath.Base(full) != filepath.Base(short) {
				rolled = true
			}
		}

		f, err := os.Open(full)
		if err != nil {
			log.Fatalf("Error opening file: %s", err)
		}
		defer f.Close()

		processor.AddLog(alias, rolled, f, *maxBuffer)
	}

	processor.Crank()
}
