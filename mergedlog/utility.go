package mergedlog

import (
	"github.com/mgutz/ansi"
	"path/filepath"
	"regexp"
	"strings"
)

// ProcessFilename processes a [tag:]file input into its constituent parts. In addition, a filename
// will be parsed to try and determine if it is a rolled filename. Specifically if it ends in a
// format like "-01-23.log". The returned values will be: original name, shorter name, tag
func ProcessFilename(name string, useFullName bool) (string, string, *string) {
	var fullName string
	var shorterName string
	var tag *string

	parts := strings.Split(name, ":")

	if len(parts) == 2 {
		tag = &parts[0]
		fullName = parts[1]
	} else {
		fullName = parts[0]
	}

	var r = regexp.MustCompile("(.*)-\\d+-\\d+.log")
	var m = r.FindStringSubmatch(fullName)
	if m != nil {
		shorterName = m[1] + ".log"
	} else {
		shorterName = fullName
	}

	if !useFullName {
		shorterName = filepath.Base(shorterName)
	}

	return fullName, shorterName, tag
}

func MakeGrepRegex(regex string) *regexp.Regexp {
	return regexp.MustCompile("(.*?)(" + regex + ")(.*)")
}

// MakeColorFn takes a color string and returns a wrapped [ansi.ColorFunc] that, when called,
// produces a [Highlighted] string.
var MakeColorFn = func(s string) func(string) Highlighted {
	cf := ansi.ColorFunc(s)
	return func(x string) Highlighted {
		return Highlighted(cf(x))
	}
}

var MakePaletteEntry = func(s string) ColorFn {
	return ColorFn{
		Normal:    MakeColorFn(s),
		Grep:      MakeColorFn(s + "+i"),
		Highlight: MakeColorFn(s + "+i"),
	}
}
