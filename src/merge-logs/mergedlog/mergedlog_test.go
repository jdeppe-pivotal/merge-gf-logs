package mergedlog_test

import (
	"container/list"
	"merge-logs/mergedlog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("adding lines", func() {
	Context("using a single file", func() {
		It("adding one line", func() {
			aggLog := list.New()
			log := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}

			line := &mergedlog.LogLine{UTime: 0, Text: "line 1"}
			log.Insert(line)

			Expect(aggLog.Front().Value).To(Equal(line))
		})

		It("adding 2 lines in order", func() {
			aggLog := list.New()
			log := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}

			line1 := &mergedlog.LogLine{UTime: 0, Text: "line 1"}
			line2 := &mergedlog.LogLine{UTime: 1, Text: "line 2"}
			log.Insert(line1)
			log.Insert(line2)

			Expect(aggLog.Front().Value).To(Equal(line1))
			Expect(aggLog.Back().Value).To(Equal(line2))
		})

		It("adding 2 lines out of order", func() {
			aggLog := list.New()
			log := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}

			line1 := &mergedlog.LogLine{UTime: 0, Text: "line 1"}
			line2 := &mergedlog.LogLine{UTime: 1, Text: "line 2"}
			log.Insert(line2)
			log.Insert(line1)

			Expect(aggLog.Front().Value).To(Equal(line1))
			Expect(aggLog.Back().Value).To(Equal(line2))
		})

		It("adding 3 lines out of order", func() {
			aggLog := list.New()
			log := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}

			line1 := &mergedlog.LogLine{UTime: 0, Text: "line 1"}
			line2 := &mergedlog.LogLine{UTime: 1, Text: "line 2"}
			line3 := &mergedlog.LogLine{UTime: 2, Text: "line 3"}
			log.Insert(line3)
			log.Insert(line1)
			log.Insert(line2)

			line := aggLog.Front()
			Expect(line.Value).To(Equal(line1))
			line = line.Next()
			Expect(line.Value).To(Equal(line2))
			line = line.Next()
			Expect(line.Value).To(Equal(line3))
		})

		It("adding lines with the same time", func() {
			aggLog := list.New()
			log := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}

			line1 := &mergedlog.LogLine{UTime: 0, Text: "line 1"}
			line2 := &mergedlog.LogLine{UTime: 1, Text: "line 2"}
			line3 := &mergedlog.LogLine{UTime: 1, Text: "line 3"}
			log.Insert(line2)
			log.Insert(line1)
			log.Insert(line3)

			line := aggLog.Front()
			Expect(line.Value).To(Equal(line1))
			line = line.Next()
			Expect(line.Value).To(Equal(line2))
			line = line.Next()
			Expect(line.Value).To(Equal(line3))
		})

		It("only retains lines within the range", func() {
			aggLog := list.New()
			log := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  1,
			}

			line1 := &mergedlog.LogLine{UTime: 0, Text: "line 1"}
			line2 := &mergedlog.LogLine{UTime: 1, Text: "line 2"}
			line3 := &mergedlog.LogLine{UTime: 2, Text: "line 3"}

			log.Insert(line2)
			log.Insert(line1)
			log.Insert(line3)

			line := aggLog.Front()
			Expect(line.Value).To(Equal(line1))
			line = line.Next()
			Expect(line.Value).To(Equal(line2))
			Expect(line.Next()).To(BeNil())
		})

		It("adding timeless as the first line", func() {
			aggLog := list.New()
			log := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}

			log.InsertTimeless("line")
			line := aggLog.Front()
			Expect(line.Value).To(Equal(&mergedlog.LogLine{UTime: 0, Text: "line"}))
		})

		It("adding timeless as the first line with non-zero start", func() {
			aggLog := list.New()
			log := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 1,
				RangeStop:  mergedlog.MAX_INT,
			}

			log.InsertTimeless("line")
			line := aggLog.Front()
			Expect(line).Should(BeNil())
		})
	})

	Context("Using 2 files", func() {
		It("works add one line from each file", func() {
			aggLog := list.New()
			log1 := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}
			log2 := &mergedlog.LogFile{
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}

			line1 := &mergedlog.LogLine{UTime: 0, Text: "line 1"}
			line2 := &mergedlog.LogLine{UTime: 1, Text: "line 2"}
			line3 := &mergedlog.LogLine{UTime: 2, Text: "line 3"}
			line4 := &mergedlog.LogLine{UTime: 2, Text: "line 4"}
			log1.Insert(line3)
			log2.Insert(line1)
			log1.Insert(line2)
			log2.Insert(line4)

			line := aggLog.Front()
			Expect(line.Value).To(Equal(line1))
			line = line.Next()
			Expect(line.Value).To(Equal(line2))
			line = line.Next()
			Expect(line.Value).To(Equal(line3))
			line = line.Next()
			Expect(line.Value).To(Equal(line4))
		})
	})

	Context("Using 2 files", func() {
		It("works adding a timeless line from each file", func() {
			aggLog := list.New()
			log1 := &mergedlog.LogFile{
				Alias:      "vm1",
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}
			log2 := &mergedlog.LogFile{
				Alias:      "vm2",
				AggLog:     aggLog,
				RangeStart: 0,
				RangeStop:  mergedlog.MAX_INT,
			}

			line1 := &mergedlog.LogLine{UTime: 0, Text: "line 1"}
			line2 := &mergedlog.LogLine{UTime: 1, Text: "line 2"}
			line3 := &mergedlog.LogLine{UTime: 2, Text: "line 3"}

			r := log1.Insert(line2)
			Expect(r.Value).To(Equal(line2))
			r = log2.Insert(line1)
			Expect(r.Value).To(Equal(line1))
			r = log1.Insert(line3)
			Expect(r.Value).To(Equal(line3))

			line4 := &mergedlog.LogLine{Alias: "vm1", UTime: 2, Text: "line 4"}
			r = log1.InsertTimeless("line 4")
			Expect(r.Value).To(Equal(line4))

			line5 := &mergedlog.LogLine{Alias: "vm2", UTime: 0, Text: "line 5"}
			r = log2.InsertTimeless("line 5")
			Expect(r.Value).To(Equal(line5))

			//mergedlog.Dump(aggLog)

			line := aggLog.Front()
			Expect(line.Value).To(Equal(line1))
			line = line.Next()
			Expect(line.Value).To(Equal(line5))
			line = line.Next()
			Expect(line.Value).To(Equal(line2))
			line = line.Next()
			Expect(line.Value).To(Equal(line3))
			line = line.Next()
			Expect(line.Value).To(Equal(line4))
		})
	})

	Context("Custom scan function", func() {
		It("works at the end of a reader with no data", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte{}, true)
			Expect(advance).Should(Equal(0))
			Expect(token).Should(Equal([]byte{}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("works at the end of a reader with data", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line\n"), true)
			Expect(advance).Should(Equal(5))
			Expect(token).Should(Equal([]byte("line")))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("requests more data when sep is no present", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line"), false)
			Expect(advance).Should(Equal(0))
			Expect(token).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("requests more data when sep is no present", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line\n["), false)
			Expect(advance).Should(Equal(5))
			Expect(token).Should(Equal([]byte("line")))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
