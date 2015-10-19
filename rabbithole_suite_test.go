package rabbithole

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRabbitHole(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rabbithole Suite")
}
