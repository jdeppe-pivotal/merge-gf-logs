package mergedlog_test

import (
	"merge-logs/mergedlog"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func makeSpan(text string) mergedlog.LogEntry {
	s := mergedlog.Span{text}
	e := mergedlog.LogEntry{s}
	return e
}

var _ = Describe("adding lines", func() {

	Context("Custom scan function", func() {
		It("works at the end of a reader with no data", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte{}, true)
			Expect(advance).Should(Equal(0))
			Expect(token).Should(Equal([]byte{}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("works at the end of a reader with data with trailing newline", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line\n"), true)
			Expect(advance).Should(Equal(5))
			Expect(token).Should(Equal([]byte("line")))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("works at the end of a reader with data without trailing newline", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line"), true)
			Expect(advance).Should(Equal(4))
			Expect(token).Should(Equal([]byte("line")))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("requests more data when sep is not present", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line"), false)
			Expect(advance).Should(Equal(0))
			Expect(token).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("requests more data when sep is not present", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line\n["), false)
			Expect(advance).Should(Equal(0))
			Expect(token).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("requests more data when [ is not followed by a non-newline", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line\n[\n"), false)
			Expect(advance).Should(Equal(0))
			Expect(token).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("returns the line when [ is followed by a non-newline", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line\n[w"), false)
			Expect(advance).Should(Equal(5))
			Expect(token).Should(Equal([]byte("line")))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("returns the line containing only [", func() {
			advance, token, err := mergedlog.ScanLogEntries([]byte("line\n[\n[w"), false)
			Expect(advance).Should(Equal(7))
			Expect(token).Should(Equal([]byte("line\n[")))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
