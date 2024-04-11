package mergedlog_test

import (
	"bufio"
	"merge-logs/mergedlog"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, 0)
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
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, 0)
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
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, 0)
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
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, 0)
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
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, nilRegex, 0)
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
			processor := mergedlog.NewProcessor(1447951959505000000, 1447951959506000000, nilRegex, nilRegex, 0)
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

			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, regex, nilRegex, 0)
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

	Context("when grepping for text across multiple files", func() {
		It("returns log entry from all files", func() {
			regex := mergedlog.MakeGrepRegex("SomeException")
			testPalette := make([]mergedlog.ColorFn, 1)
			f1 := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted(s) }
			f2 := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted("#" + s + "#") }
			testPalette[0] = mergedlog.ColorFn{f1, f2, f1}

			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, regex, nilRegex, 0)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(testPalette)

			file1 := `[info 2015/11/19 08:52:39.504 PST  line1 also has SomeException in it
SomeException
  at foo.com

[info 2015/11/19 08:52:40.774 PST  line2`

			file2 := `[info 2015/11/19 08:42:39.504 PST  line3 may have SomeException in it

[info 2015/11/19 08:52:40.774 PST  line4 does not

[info 2015/11/19 08:57:40.774 PST  line5 has SomeException`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.AddLog("", false, strings.NewReader(file2), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[] [info 2015/11/19 08:42:39.504 PST  line3 may have #SomeException# in it",
				"[] ",
				"[] [info 2015/11/19 08:52:39.504 PST  line1 also has #SomeException# in it",
				"[] #SomeException#",
				"[]   at foo.com",
				"[] ",
				"[] [info 2015/11/19 08:57:40.774 PST  line5 has #SomeException#",
			}))
		})
	})

	Context("when grepping across multiple files with wide ranging timestamps", func() {
		It("returns log entry from all files", func() {
			regex := mergedlog.MakeGrepRegex("Exception")
			testPalette := make([]mergedlog.ColorFn, 1)
			f1 := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted(s) }
			f2 := func(s string) mergedlog.Highlighted { return mergedlog.Highlighted("#" + s + "#") }
			testPalette[0] = mergedlog.ColorFn{f1, f2, f1}

			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, regex, nilRegex, 0)
			result := &strings.Builder{}
			processor.SetWriter(result)
			processor.SetPalette(testPalette)
			mergedlog.FLUSH_BATCH_SIZE = 0

			file1 := `[info 2015/11/19 08:00:00.000 PST  line1-1 Exception
[info 2015/11/19 08:01:00.000 PST  line1-2
[info 2015/11/19 08:02:00.000 PST  line1-3
[info 2015/11/19 08:03:00.000 PST  line1-4
[info 2015/11/19 08:04:00.000 PST  line1-5 Exception`

			file2 := `[info 2015/11/19 08:00:01.000 PST  line2-1 Exception
[info 2015/11/19 08:01:01.000 PST  line2-2
[info 2015/11/19 08:02:01.000 PST  line2-3`

			file3 := `[info 2015/11/19 08:00:01.000 PST  line3-1
[info 2015/11/19 08:01:01.000 PST  line3-2
[info 2015/11/19 08:02:01.000 PST  line3-3
[info 2015/11/19 08:05:01.000 PST  line3-4 Exception`

			processor.AddLog("", false, strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.AddLog("", false, strings.NewReader(file2), bufio.MaxScanTokenSize)
			processor.AddLog("", false, strings.NewReader(file3), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(strings.Split(strings.TrimSpace(result.String()), "\n")).To(Equal([]string{
				"[] [info 2015/11/19 08:00:00.000 PST  line1-1 #Exception#",
				"[] [info 2015/11/19 08:00:01.000 PST  line2-1 #Exception#",
				"[] [info 2015/11/19 08:04:00.000 PST  line1-5 #Exception#",
				"[] [info 2015/11/19 08:05:01.000 PST  line3-4 #Exception#",
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

			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT, nilRegex, regex, 0)
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
