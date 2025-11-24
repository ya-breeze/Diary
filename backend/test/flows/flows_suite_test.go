package flows_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFlows(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flows Suite")
}
