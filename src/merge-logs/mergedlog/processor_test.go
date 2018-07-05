package mergedlog_test

import (
	"bufio"
	"merge-logs/mergedlog"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("processor integration test", func() {
	Context("when processing a single file", func() {
		It("returns the same content", func() {
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT)
			result := &strings.Builder{}
			processor.SetWriter(result)

			file1 := "foo\nbar\nbaz"

			processor.AddLog("", strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.Crank()

			Expect(result.String()).To(Equal("[] " + strings.Join(strings.Split(file1, "\n"), "\n[] ") + "\n"))
		})
	})

	Context("when processing a single file", func() {
		It("returns correctly ordered content", func() {
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT)
			result := &strings.Builder{}
			processor.SetWriter(result)

			file1 := `[fine 2015/11/19 08:52:39.504 PST  line1

[fine 2015/11/19 08:52:39.505 PST  line2

[fine 2015/11/19 08:52:39.506 PST  line3`

			processor.AddLog("", strings.NewReader(file1), bufio.MaxScanTokenSize)
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

	Context("when processing multiple files with dates", func() {
		It("returns correctly ordered content", func() {
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT)
			result := &strings.Builder{}
			processor.SetWriter(result)

			file1 := `[fine 2015/11/19 08:52:39.504 PST  line1

[fine 2015/11/19 08:52:39.506 PST  line2`
			file2 := `[fine 2015/11/19 08:52:39.505 PST  line3`

			processor.AddLog("", strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.AddLog("", strings.NewReader(file2), bufio.MaxScanTokenSize)
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
			processor := mergedlog.NewProcessor(0, mergedlog.MAX_INT)
			result := &strings.Builder{}
			processor.SetWriter(result)

			file1 := `[fine 2015/11/19 08:52:39.504 PST  line1
SomeException
  at foo.com
Caused by: AnotherException
  at bar.com
[fine 2015/11/19 08:52:39.506 PST  line2`
			file2 := `[fine 2015/11/19 08:52:39.505 PST  line3
AnotherException
  at acme.com`

			processor.AddLog("", strings.NewReader(file1), bufio.MaxScanTokenSize)
			processor.AddLog("", strings.NewReader(file2), bufio.MaxScanTokenSize)
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
})
