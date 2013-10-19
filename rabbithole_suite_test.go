package rabbithole_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRabbitHole(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client")
}
