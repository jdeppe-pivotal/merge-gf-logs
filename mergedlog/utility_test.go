package mergedlog_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"merge-logs/mergedlog"
)

var _ = Describe("MergeGfLogs", func() {
	Context("processing file names", func() {
		It("returns full name", func() {
			name, shorter, tag := mergedlog.ProcessFilename("foo/bar/server-1.log", true)
			Expect(name).To(Equal("foo/bar/server-1.log"))
			Expect(shorter).To(Equal("foo/bar/server-1.log"))
			Expect(tag).To(BeNil())
		})
		It("returns base name", func() {
			name, shorter, tag := mergedlog.ProcessFilename("foo/bar/server-1.log", false)
			Expect(name).To(Equal("foo/bar/server-1.log"))
			Expect(shorter).To(Equal("server-1.log"))
			Expect(tag).To(BeNil())
		})
		It("returns tag with full name", func() {
			name, shorter, tag := mergedlog.ProcessFilename("tag:foo/bar/server-1.log", true)
			Expect(name).To(Equal("foo/bar/server-1.log"))
			Expect(shorter).To(Equal("foo/bar/server-1.log"))
			Expect(*tag).To(Equal("tag"))
		})
		It("returns tag with base name", func() {
			name, shorter, tag := mergedlog.ProcessFilename("tag:foo/bar/server-1.log", false)
			Expect(name).To(Equal("foo/bar/server-1.log"))
			Expect(shorter).To(Equal("server-1.log"))
			Expect(*tag).To(Equal("tag"))
		})
		It("returns shorter name with full name", func() {
			name, shorter, tag := mergedlog.ProcessFilename("foo/bar/server-01-02.log", true)
			Expect(name).To(Equal("foo/bar/server-01-02.log"))
			Expect(shorter).To(Equal("foo/bar/server.log"))
			Expect(tag).To(BeNil())
		})
		It("returns shorter name with base name", func() {
			name, shorter, tag := mergedlog.ProcessFilename("foo/bar/server-01-02.log", false)
			Expect(name).To(Equal("foo/bar/server-01-02.log"))
			Expect(shorter).To(Equal("server.log"))
			Expect(tag).To(BeNil())
		})
		It("returns shorter name and tag with full name", func() {
			name, shorter, tag := mergedlog.ProcessFilename("tag:foo/bar/server-01-02.log", true)
			Expect(name).To(Equal("foo/bar/server-01-02.log"))
			Expect(shorter).To(Equal("foo/bar/server.log"))
			Expect(*tag).To(Equal("tag"))
		})
		It("returns shorter name and tag with base name", func() {
			name, shorter, tag := mergedlog.ProcessFilename("foo/bar/server-01-02.log", false)
			Expect(name).To(Equal("foo/bar/server-01-02.log"))
			Expect(shorter).To(Equal("server.log"))
			Expect(tag).To(BeNil())
		})
	})
})
