package mergedlog_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMergedlog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mergedlog Suite")
}
