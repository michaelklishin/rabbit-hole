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
})
