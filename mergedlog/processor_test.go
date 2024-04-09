package mergedlog_test

import (
	"bufio"
	"merge-logs/mergedlog"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Getting a pointer to a literal int(0)
var debug = func(i int) *int { return &i }(0)

// regex is nil
var nilRegex *regexp.Regexp
var noopPalette []mergedlog.ColorFn

func init() {
	noopPalette = make([]mergedlog.ColorFn, 1)
	f := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted(s) }
	noopPalette[0] = mergedlog.ColorFn{f, f, f}
}

var _ = Describe("processor integration test", func() {
	Context("when processing a single file", func() {
		It("returns the same content", func() {
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, debug)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(noopPalette)

			file1 := "[fine 2015/11/19 08:52:39.504 PST foo\nbar\nbaz"

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(result.String()).To(Equal("[] " + strings.Join(strings.Split(file1, "\n"), "\n[] ") + "\n"))
		})
	})

	Context("when processing a single file", func() {
		It("returns correctly ordered content", func() {
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, debug)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(noopPalette)

			file1 := `[fine 2015/11/19 08:52:39.504 PST  line1

[fine 2015/11/19 08:52:39.505 PST  line2

[fine 2015/11/19 08:52:39.506 PST  line3`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[] [fine 2015/11/19 08:52:39.504 PST  line1",
				"[] ",
				"[] [fine 2015/11/19 08:52:39.505 PST  line2",
				"[] ",
				"[] [fine 2015/11/19 08:52:39.506 PST  line3",
			}))
		})
	})

	Context("when processing a single file with incorrectly ordered lines", func() {
		It("returns correctly ordered content", func() {
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, debug)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(noopPalette)

			file1 := `[fine 2015/11/19 08:52:39.506 GMT  line2

[fine 2015/11/19 08:52:39.506 GMT  line3

[fine 2015/11/19 08:52:39.504 GMT  line1

[fine 2015/11/19 08:52:39.504 GMT  line1.5

[fine 2015/11/19 08:52:39.507 GMT  line4
`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[] [fine 2015/11/19 08:52:39.504 GMT  line1",
				"[] ",
				"[] [fine 2015/11/19 08:52:39.504 GMT  line1.5",
				"[] ",
				"[] [fine 2015/11/19 08:52:39.506 GMT  line2",
				"[] ",
				"[] [fine 2015/11/19 08:52:39.506 GMT  line3",
				"[] ",
				"[] [fine 2015/11/19 08:52:39.507 GMT  line4",
			}))
		})
	})

	Context("when processing multiple files with dates", func() {
		It("returns correctly ordered content", func() {
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, debug)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(noopPalette)

			file1 := `[fine 2015/11/19 08:52:39.504 PST  line1

[fine 2015/11/19 08:52:39.506 PST  line2`
			file2 := `[fine 2015/11/19 08:52:39.505 PST  line3`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.AddLog("", false, strings.NewReader(file2), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[] [fine 2015/11/19 08:52:39.504 PST  line1",
				"[] ",
				"[] [fine 2015/11/19 08:52:39.505 PST  line3",
				"[] [fine 2015/11/19 08:52:39.506 PST  line2",
			}))
		})
	})

	Context("when processing multiple files with undated lines", func() {
		It("returns correctly ordered content", func() {
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, debug)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(noopPalette)

			file1 := `[fine 2015/11/19 08:52:39.504 PST  line1
SomeException
  at foo.com
Caused by: AnotherException
  at bar.com
[fine 2015/11/19 08:52:39.506 PST  line2`
			file2 := `[fine 2015/11/19 08:52:39.505 PST  line3
AnotherException
  at acme.com`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.AddLog("", false, strings.NewReader(file2), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[] [fine 2015/11/19 08:52:39.504 PST  line1",
				"[] SomeException",
				"[]   at foo.com",
				"[] Caused by: AnotherException",
				"[]   at bar.com",
				"[] [fine 2015/11/19 08:52:39.505 PST  line3",
				"[] AnotherException",
				"[]   at acme.com",
				"[] [fine 2015/11/19 08:52:39.506 PST  line2",
			}))
		})
	})

	Context("when limiting output by timestamp", func() {
		It("returns correctly ordered content", func() {
			processor := mergedlog.NewProcessor(1447951959505000000, 1447951959506000000, nilRegex, nilRegex, debug)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(noopPalette)

			file1 := `[fine 2015/11/19 08:52:39.504 PST  line1

[fine 2015/11/19 08:52:39.506 PST  line2`
			file2 := `[fine 2015/11/19 08:52:39.505 PST  line3

[fine 2015/11/19 08:52:39.507 PST  line4`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.AddLog("", false, strings.NewReader(file2), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[] [fine 2015/11/19 08:52:39.505 PST  line3",
				"[] ",
				"[] [fine 2015/11/19 08:52:39.506 PST  line2",
			}))
		})
	})

	Context("when grepping for text", func() {
		It("returns complete log entry", func() {
			regex := mergedlog.MakeGrepRegex("SomeException")
			testPalette := make([]mergedlog.ColorFn, 1)
			f1 := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted(s) }
			f2 := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted("#" + s + "#") }
			testPalette[0] = mergedlog.ColorFn{f1, f2, f1}

			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, regex, nilRegex, debug)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(testPalette)

			file1 := `[info 2015/11/19 08:52:39.504 PST  line1 also has SomeException in it
SomeException
  at foo.com
[info 2015/11/19 08:52:40.774 PST  line2`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[] [info 2015/11/19 08:52:39.504 PST  line1 also has #SomeException# in it",
				"[] #SomeException#",
				"[]   at foo.com",
			}))
		})
	})

	Context("when only highlighting text", func() {
		It("returns highlighted log entry", func() {
			regex := mergedlog.MakeGrepRegex("SomeException")
			testPalette := make([]mergedlog.ColorFn, 1)
			f1 := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted(">" + s + "<") }
			f2 := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted("#" + s + "#") }
			testPalette[0] = mergedlog.ColorFn{f1, f1, f2}

			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, regex, debug)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(testPalette)

			file1 := `[info 2015/11/19 08:52:39.504 PST  line1 also has SomeException in it
SomeException and again SomeException
  at foo.com
[info 2015/11/19 08:52:40.774 PST  line2`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[><] >[info 2015/11/19 08:52:39.504 PST  line1 also has <#SomeException#> in it<",
				"[><] ><#SomeException#> and again <#SomeException#><",
				"[><] >  at foo.com<",
				"[><] >[info 2015/11/19 08:52:40.774 PST  line2<",
			}))
		})
	})
})
