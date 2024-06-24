package rabbithole

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Health checks", func() {
	var (
		rmqc *Client
	)

	BeforeEach(func() {
		rmqc, _ = NewClient("http://127.0.0.1:15672", "guest", "guest")
	})

	Context("GET /health/checks/alarms", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			res, err := rmqc.HealthCheckAlarms()
			Ω(err).Should(BeNil())

			Ω(res.Ok()).Should(BeTrue())
			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/local-alarms", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			res, err := rmqc.HealthCheckLocalAlarms()
			Ω(err).Should(BeNil())

			Ω(res.Ok()).Should(BeTrue())
			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/certificate-expiration/1/days", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			res, err := rmqc.HealthCheckCertificateExpiration(1, DAYS)
			Ω(err).Should(BeNil())

			Ω(res.Ok()).Should(BeTrue())
			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/port-listener/5672", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			res, err := rmqc.HealthCheckPortListener(5672)
			Ω(err).Should(BeNil())

			res.Ok()
			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/protocol-listener/amqp091", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			res, err := rmqc.HealthCheckProtocolListener(AMQP091)
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/virtual-hosts", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			res, err := rmqc.HealthCheckVirtualHosts()
			Ω(err).Should(BeNil())

			Ω(res.Ok()).Should(BeTrue())
			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/node-is-quorum-critical", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			res, err := rmqc.HealthCheckNodeIsQuorumCritical()
			Ω(err).Should(BeNil())

			Ω(res.Ok()).Should(BeTrue())
			Ω(res.Status).Should(Equal("ok"))
		})
	})
})
