package rabbithole

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streadway/amqp"
)

// TODO: extract duplication between these
func FindQueueByName(qs []QueueInfo, name string) QueueInfo {
	var q QueueInfo
	for _, i := range qs {
		if i.Name == name {
			q = i
		}
	}

	return q
}

func FindUserByName(us []UserInfo, name string) UserInfo {
	var u UserInfo
	for _, i := range us {
		if name == i.Name {
			u = i
		}
	}

	return u
}

var _ = Describe("Client", func() {
	var (
		rmqc *Client
	)

	BeforeEach(func() {
		rmqc, _ = NewClient("http://127.0.0.1:15672", "guest", "guest")
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

			Ω(res.MemUsed).Should(BeNumerically(">", 10*1024*1024))
			Ω(res.MemLimit).Should(BeNumerically(">", 64*1024*1024))
			Ω(res.MemAlarm).Should(Equal(false))

			Ω(res.IsRunning).Should(Equal(true))

			Ω(res.SocketsUsed).Should(BeNumerically(">=", 0))
			Ω(res.SocketsTotal).Should(BeNumerically(">=", 1))

		})
	})

	Context("GET /connections when there are active connections", func() {
		It("returns decoded response", func() {
			// this really should be tested with > 1 connection and channel. MK.
			conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
			Ω(err).Should(BeNil())
			defer conn.Close()

			conn2, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
			Ω(err).Should(BeNil())
			defer conn2.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			ch2, err := conn2.Channel()
			Ω(err).Should(BeNil())
			defer ch2.Close()

			ch3, err := conn2.Channel()
			Ω(err).Should(BeNil())
			defer ch3.Close()

			err = ch.Publish("",
				"",
				false,
				false,
				amqp.Publishing{Body: []byte("")})
			Ω(err).Should(BeNil())

			xs, err := rmqc.ListConnections()
			Ω(err).Should(BeNil())

			info := xs[0]
			Ω(info.Name).ShouldNot(BeNil())
			Ω(info.Host).Should(Equal("127.0.0.1"))
			Ω(info.UsesTLS).Should(Equal(false))
		})
	})

	Context("GET /channels when there are active connections with open channels", func() {
		It("returns decoded response", func() {
			conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
			Ω(err).Should(BeNil())
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			ch2, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch2.Close()

			ch3, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch3.Close()

			ch4, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch4.Close()

			err = ch.Publish("",
				"",
				false,
				false,
				amqp.Publishing{Body: []byte("")})
			Ω(err).Should(BeNil())

			err = ch2.Publish("",
				"",
				false,
				false,
				amqp.Publishing{Body: []byte("")})
			Ω(err).Should(BeNil())

			xs, err := rmqc.ListChannels()
			Ω(err).Should(BeNil())

			info := xs[0]
			Ω(info.Node).ShouldNot(BeNil())
			Ω(info.User).Should(Equal("guest"))
			Ω(info.Vhost).Should(Equal("/"))

			Ω(info.Transactional).Should(Equal(false))

			Ω(info.UnacknowledgedMessageCount).Should(Equal(0))
			Ω(info.UnconfirmedMessageCount).Should(Equal(0))
			Ω(info.UncommittedMessageCount).Should(Equal(0))
			Ω(info.UncommittedAckCount).Should(Equal(0))
		})
	})

	Context("GET /connections/{name] when connection exists", func() {
		It("returns decoded response", func() {
			// this really should be tested with > 1 connection and channel. MK.
			conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
			Ω(err).Should(BeNil())
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			err = ch.Publish("",
				"",
				false,
				false,
				amqp.Publishing{Body: []byte("")})
			Ω(err).Should(BeNil())

			xs, err := rmqc.ListConnections()
			Ω(err).Should(BeNil())

			c1 := xs[0]
			info, err := rmqc.GetConnection(c1.Name)
			Ω(err).Should(BeNil())
			Ω(info.Protocol).Should(Equal("AMQP 0-9-1"))
			Ω(info.User).Should(Equal("guest"))
		})
	})

	Context("GET /channels/{name} when channel exists", func() {
		It("returns decoded response", func() {
			conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
			Ω(err).Should(BeNil())
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			err = ch.Publish("",
				"",
				false,
				false,
				amqp.Publishing{Body: []byte("")})
			Ω(err).Should(BeNil())

			xs, err := rmqc.ListChannels()
			Ω(err).Should(BeNil())

			x := xs[0]
			info, err := rmqc.GetChannel(x.Name)

			Ω(info.Node).ShouldNot(BeNil())
			Ω(info.User).Should(Equal("guest"))
			Ω(info.Vhost).Should(Equal("/"))

			Ω(info.Transactional).Should(Equal(false))

			Ω(info.UnacknowledgedMessageCount).Should(Equal(0))
			Ω(info.UnconfirmedMessageCount).Should(Equal(0))
			Ω(info.UncommittedMessageCount).Should(Equal(0))
			Ω(info.UncommittedAckCount).Should(Equal(0))
		})
	})

	Context("GET /exchanges", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListExchanges()
			Ω(err).Should(BeNil())

			x := xs[0]
			Ω(x.Name).Should(Equal(""))
			Ω(x.Durable).Should(Equal(true))
		})
	})

	Context("GET /exchanges/{vhost}", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListExchangesIn("/")
			Ω(err).Should(BeNil())

			x := xs[0]
			Ω(x.Name).Should(Equal(""))
			Ω(x.Durable).Should(Equal(true))
		})
	})

	Context("GET /exchanges/{vhost}/{name}", func() {
		It("returns decoded response", func() {
			x, err := rmqc.GetExchange("rabbit/hole", "amq.fanout")
			Ω(err).Should(BeNil())
			Ω(x.Name).Should(Equal("amq.fanout"))
			Ω(x.Durable).Should(Equal(true))
			Ω(x.Type).Should(Equal("fanout"))
			Ω(x.Vhost).Should(Equal("rabbit/hole"))
		})
	})

	Context("GET /queues", func() {
		It("returns decoded response", func() {
			conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
			Ω(err).Should(BeNil())
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			ch.QueueDeclare(
				"q1",  // name
				false, // durable
				false, // delete when usused
				true,  // exclusive
				false,
				nil)

			qs, err := rmqc.ListQueues()
			Ω(err).Should(BeNil())

			q := qs[0]
			Ω(q.Name).ShouldNot(Equal(""))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Durable).ShouldNot(BeNil())
			Ω(q.Status).ShouldNot(BeNil())
		})
	})

	Context("GET /queues/{vhost}", func() {
		It("returns decoded response", func() {
			conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/rabbit%2Fhole")
			Ω(err).Should(BeNil())
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			ch.QueueDeclare(
				"q1",  // name
				false, // durable
				false, // delete when usused
				true,  // exclusive
				false,
				nil)

			qs, err := rmqc.ListQueuesIn("rabbit/hole")
			Ω(err).Should(BeNil())

			q := FindQueueByName(qs, "q1")
			Ω(q.Name).Should(Equal("q1"))
			Ω(q.Vhost).Should(Equal("rabbit/hole"))
			Ω(q.Durable).Should(Equal(false))
			Ω(q.Status).ShouldNot(BeNil())
		})
	})

	Context("GET /queues/{vhost}/{name}", func() {
		It("returns decoded response", func() {
			conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/rabbit%2Fhole")
			Ω(err).Should(BeNil())
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			ch.QueueDeclare(
				"q1",  // name
				false, // durable
				false, // delete when usused
				true,  // exclusive
				false,
				nil)

			q, err := rmqc.GetQueue("rabbit/hole", "q1")
			Ω(err).Should(BeNil())

			Ω(q.Name).Should(Equal("q1"))
			Ω(q.Vhost).Should(Equal("rabbit/hole"))
			Ω(q.Durable).Should(Equal(false))
			Ω(q.Status).ShouldNot(BeNil())
		})
	})

	Context("GET /users", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListUsers()
			Ω(err).Should(BeNil())

			u := FindUserByName(xs, "guest")
			Ω(u.Name).Should(BeEquivalentTo("guest"))
			Ω(u.PasswordHash).ShouldNot(BeNil())
			Ω(u.Tags).Should(Equal("administrator"))
		})
	})

	Context("GET /users/{name} when user exists", func() {
		It("returns decoded response", func() {
			u, err := rmqc.GetUser("guest")
			Ω(err).Should(BeNil())

			Ω(u.Name).Should(BeEquivalentTo("guest"))
			Ω(u.PasswordHash).ShouldNot(BeNil())
			Ω(u.Tags).Should(Equal("administrator"))
		})
	})

	Context("PUT /users/{name}", func() {
		It("updates the user", func() {
			info := UserInfo{Password: "s3krE7", Tags: "management policymaker"}
			resp, err := rmqc.PutUser("rabbithole", info)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(Equal("204 No Content"))

			u, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(BeNil())

			Ω(u.PasswordHash).ShouldNot(BeNil())
			Ω(u.Tags).Should(Equal("management policymaker"))
		})
	})

	Context("DELETE /users/{name}", func() {
		It("deletes the user", func() {
			info := UserInfo{Password: "s3krE7", Tags: "management policymaker"}
			_, err := rmqc.PutUser("rabbithole", info)
			Ω(err).Should(BeNil())

			u, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(BeNil())
			Ω(u).ShouldNot(BeNil())

			resp, err := rmqc.DeleteUser("rabbithole")
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(Equal("204 No Content"))

			u2, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(Equal(errors.New("user not found")))
			Ω(u2).Should(BeNil())
		})
	})
})
