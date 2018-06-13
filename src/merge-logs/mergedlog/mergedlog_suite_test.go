package mergedlog

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMergedlog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mergedlog Suite")
}
