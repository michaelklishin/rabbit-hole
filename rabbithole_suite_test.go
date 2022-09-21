package rabbithole

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRabbitHole(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(5*time.Second)
	RunSpecs(t, "Rabbithole Suite")
}
