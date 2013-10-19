package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "rabbitmq/http/client"
)

var _ = Describe("Client", func() {
	var (
		rmqc *Client
	)

	BeforeEach(func() {
		rmqc = NewClient("http://127.0.0.1:15672", "guest", "guest")
	})

	Context("GET /overview", func() {
		It("returns decoded response", func() {
			res, err := rmqc.Overview()

			Ω(err).Should(BeNil())

			Ω(res.Node).ShouldNot(BeNil())
			Ω(res.StatisticsDBNode).ShouldNot(BeNil())

			fanoutExchange := ExchangeType{Name: "fanout", Description: "AMQP fanout exchange, as per the AMQP specification", Enabled: true}
			Ω(res.ExchangeTypes).Should(ContainElement(fanoutExchange))

		})
	})

	Context("GET /nodes", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListNodes()
			res := xs[0]

			Ω(err).Should(BeNil())

			Ω(res.Name).ShouldNot(BeNil())
			Ω(res.NodeType).Should(Equal("disc"))

			Ω(res.FdUsed).Should(BeNumerically(">=", 0))
			Ω(res.FdTotal).Should(BeNumerically(">", 64))

			Ω(res.MemUsed).Should(BeNumerically(">", 10 * 1024 * 1024))
			Ω(res.MemLimit).Should(BeNumerically(">", 64 * 1024 * 1024))
			Ω(res.MemAlarm).Should(Equal(false))

			Ω(res.IsRunning).Should(Equal(true))

			Ω(res.SocketsUsed).Should(BeNumerically(">=", 0))
			Ω(res.SocketsTotal).Should(BeNumerically(">=", 1))

		})
	})
})
