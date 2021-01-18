package rabbithole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streadway/amqp"
)

func FindQueueByName(sl []QueueInfo, name string) (q QueueInfo) {
	for _, i := range sl {
		if name == i.Name {
			q = i
			break
		}
	}
	return q
}

func FindUserByName(sl []UserInfo, name string) (u UserInfo) {
	for _, i := range sl {
		if name == i.Name {
			u = i
			break
		}
	}
	return u
}

func openConnection(vhost string) *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/" + url.QueryEscape(vhost))
	Ω(err).Should(BeNil())

	if err != nil {
		panic("failed to connect")
	}

	return conn
}

func ensureNonZeroMessageRate(ch *amqp.Channel) {
	for i := 0; i < 2000; i++ {
		q, _ := ch.QueueDeclare(
			"",    // name
			false, // durable
			false, // auto delete
			true,  // exclusive
			false,
			nil)
		_ = ch.Publish("", q.Name, false, false, amqp.Publishing{Body: []byte("")})
	}
}

// Wait for the list of connections to reach the expected length
func listConnectionsUntil(c *Client, i int) {
	xs, _ := c.ListConnections()
	// Avoid infinity loops by breaking it after 30s
	breakLoop := 0
	for i != len(xs) {
		if breakLoop == 300 {
			fmt.Printf("Stopping listConnectionsUntil loop: expected %v obtained %v", i, len(xs))
			break
		}
		breakLoop++
		// Wait between calls
		time.Sleep(100 * time.Millisecond)
		xs, _ = c.ListConnections()
	}
}

func awaitEventPropagation() {
	time.Sleep(1150 * time.Millisecond)
}

type portTestStruct struct {
	Port Port `json:"port"`
}

var _ = Describe("Rabbithole", func() {
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

			res, err := rmqc.HealthCheckAlarms()
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/local-alarms", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			res, err := rmqc.HealthCheckLocalAlarms()
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/certificate-expiration/1/days", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			res, err := rmqc.HealthCheckCertificateExpiration(1, DAYS)
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/port-listener/5672", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			res, err := rmqc.HealthCheckPortListenerListener(5672)
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/protocol-listener/amqp091", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			res, err := rmqc.HealthCheckProtocolListener(AMQP091)
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/virtual-hosts", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			res, err := rmqc.HealthCheckVirtualHosts()
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/node-is-mirror-sync-critical", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			res, err := rmqc.HealthCheckNodeIsMirrorSyncCritical()
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /health/checks/node-is-quorum-critical", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			res, err := rmqc.HealthCheckNodeIsMirrorSyncCritical()
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))
		})
	})

	Context("GET /aliveness-test/%2F", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())

			ensureNonZeroMessageRate(ch)

			res, err := rmqc.Aliveness("%2F")
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))

			ch.Close()
		})
	})

	Context("GET /healthchecks/nodes", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())

			ensureNonZeroMessageRate(ch)

			res, err := rmqc.HealthCheck()
			Ω(err).Should(BeNil())

			Ω(res.Status).Should(Equal("ok"))

			ch.Close()
		})
	})

	Context("GET /overview", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())

			ensureNonZeroMessageRate(ch)

			res, err := rmqc.Overview()
			Ω(err).Should(BeNil())

			Ω(res.Node).ShouldNot(BeNil())
			Ω(res.StatisticsDBNode).ShouldNot(BeNil())
			Ω(res.MessageStats).ShouldNot(BeNil())
			Ω(res.MessageStats.DeliverDetails).ShouldNot(BeNil())
			Ω(res.MessageStats.DeliverDetails.Rate).Should(BeNumerically(">=", 0))
			Ω(res.MessageStats.DeliverNoAckDetails).ShouldNot(BeNil())
			Ω(res.MessageStats.DeliverNoAckDetails.Rate).Should(BeNumerically(">=", 0))
			Ω(res.MessageStats.DeliverGetDetails).ShouldNot(BeNil())
			Ω(res.MessageStats.DeliverGetDetails.Rate).Should(BeNumerically(">=", 0))
			Ω(res.MessageStats.ReturnUnroutable).ShouldNot(BeNil())
			Ω(res.MessageStats.ReturnUnroutableDetails.Rate).Should(BeNumerically(">=", 0))

			// there are at least 4 exchange types, the built-in ones
			Ω(len(res.ExchangeTypes)).Should(BeNumerically(">=", 4))

			ch.Close()
		})
	})

	Context("DELETE /api/connections/{name}", func() {
		It("closes the connection", func() {
			listConnectionsUntil(rmqc, 0)
			conn := openConnection("/")

			awaitEventPropagation()
			xs, err := rmqc.ListConnections()
			Ω(err).Should(BeNil())

			closeEvents := make(chan *amqp.Error)
			conn.NotifyClose(closeEvents)

			n := xs[0].Name
			_, err = rmqc.CloseConnection(n)
			Ω(err).Should(BeNil())

			evt := <-closeEvents
			Ω(evt).ShouldNot(BeNil())
			Ω(evt.Code).Should(Equal(320))
			Ω(evt.Reason).Should(Equal("CONNECTION_FORCED - Closed via management plugin"))
			// server-initiated
			Ω(evt.Server).Should(Equal(true))
		})
	})

	Context("EnabledProtocols", func() {
		It("returns a list of enabled protocols", func() {
			xs, err := rmqc.EnabledProtocols()

			Ω(err).Should(BeNil())
			Ω(xs).ShouldNot(BeEmpty())
			Ω(xs).Should(ContainElement("amqp"))
		})
	})

	Context("ProtocolPorts", func() {
		It("returns a map of enabled protocols => ports", func() {
			m, err := rmqc.ProtocolPorts()

			Ω(err).Should(BeNil())
			Ω(m["amqp"]).Should(BeEquivalentTo(5672))
		})
	})

	Context("Ports", func() {
		It("parses an int from json", func() {
			testCase := []byte("{\"port\": 1}")
			parsed := portTestStruct{}
			err := json.Unmarshal(testCase, &parsed)

			Ω(err).Should(BeNil())
			Ω(parsed.Port).Should(BeEquivalentTo(1))
		})

		It("parses an int encoded as a string from json", func() {
			testCase := []byte("{\"port\": \"1\"}")
			parsed := portTestStruct{}
			err := json.Unmarshal(testCase, &parsed)

			Ω(err).Should(BeNil())
			Ω(parsed.Port).Should(BeEquivalentTo(1))
		})
	})

	Context("GET /whoami", func() {
		It("returns decoded response", func() {
			res, err := rmqc.Whoami()

			Ω(err).Should(BeNil())

			Ω(res.Name).ShouldNot(BeNil())
			Ω(res.Name).Should(Equal("guest"))

			tags := UserTags([]string{"administrator"})
			Ω(res.Tags).Should(Equal(tags))
			Ω(res.AuthBackend).ShouldNot(BeNil())
		})
	})

	Context("GET /nodes", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListNodes()
			res := xs[0]

			Ω(err).Should(BeNil())

			Ω(res.Name).ShouldNot(BeNil())
			Ω(res.IsRunning).Should(Equal(true))
			Ω(res.Partitions).ShouldNot(BeNil())

		})
	})

	Context("GET /nodes/{name}", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListNodes()
			Ω(err).Should(BeNil())
			n := xs[0]
			res, err := rmqc.GetNode(n.Name)

			Ω(err).Should(BeNil())

			Ω(res.Name).ShouldNot(BeNil())
			Ω(res.IsRunning).Should(Equal(true))
			Ω(res.Partitions).ShouldNot(BeNil())

		})
	})

	Context("PUT /cluster-name", func() {
		It("Set cluster name", func() {
			previousClusterName, err := rmqc.GetClusterName()
			Ω(err).Should(BeNil())

			cnStr := "rabbitmq@rabbit-hole-test"
			cn := ClusterName{Name: cnStr}
			resp, err := rmqc.SetClusterName(cn)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(Equal("204 No Content"))
			awaitEventPropagation()

			cn2, err := rmqc.GetClusterName()
			Ω(err).Should(BeNil())
			Ω(cn2.Name).Should(Equal(cnStr))
			// Restore cluster name
			resp, err = rmqc.SetClusterName(*previousClusterName)
			Ω(resp.Status).Should(Equal("204 No Content"))
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /connections when there are active connections", func() {
		It("returns decoded response", func() {
			// this really should be tested with > 1 connection and channel. MK.
			conn := openConnection("/")
			defer conn.Close()

			conn2 := openConnection("/")
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

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			xs, err := rmqc.ListConnections()
			Ω(err).Should(BeNil())

			info := xs[0]
			Ω(info.Name).ShouldNot(BeNil())
			Ω(info.Host).ShouldNot(BeEmpty())
			Ω(info.UsesTLS).Should(Equal(false))
		})
	})

	Context("GET /channels when there are active connections with open channels", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
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

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

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
			conn := openConnection("/")
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

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

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
			conn := openConnection("/")
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

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			xs, err := rmqc.ListChannels()
			Ω(err).Should(BeNil())

			x := xs[0]
			info, err := rmqc.GetChannel(x.Name)
			Ω(err).Should(BeNil())

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
			Ω(x.AutoDelete).Should(Equal(false))
			Ω(x.Internal).Should(Equal(false))
			Ω(x.Type).Should(Equal("fanout"))
			Ω(x.Vhost).Should(Equal("rabbit/hole"))
			Ω(x.Incoming).Should(BeEmpty())
			Ω(x.Outgoing).Should(BeEmpty())
			Ω(x.Arguments).Should(BeEmpty())
		})
	})

	Context("GET /queues", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			_, err = ch.QueueDeclare(
				"",    // name
				false, // durable
				false, // auto delete
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			qs, err := rmqc.ListQueues()
			Ω(err).Should(BeNil())

			q := qs[0]
			Ω(q.Name).ShouldNot(Equal(""))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Durable).ShouldNot(BeNil())
			Ω(q.Status).ShouldNot(BeEmpty())
		})
	})

	Context("GET /queues with arguments", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			_, err = ch.QueueDeclare(
				"",    // name
				false, // durable
				false, // auto delete
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			params := url.Values{}
			params.Add("lengths_age", "1800")
			params.Add("lengths_incr", "30")

			qs, err := rmqc.ListQueuesWithParameters(params)
			Ω(err).Should(BeNil())

			q := qs[0]
			Ω(q.Name).ShouldNot(Equal(""))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Durable).ShouldNot(BeNil())
			Ω(q.Status).ShouldNot(BeEmpty())
			Ω(q.MessagesDetails.Samples[0]).ShouldNot(BeNil())
		})
	})

	Context("GET /queues paged with arguments", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			_, err = ch.QueueDeclare(
				"",    // name
				false, // durable
				false, // auto delete
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			params := url.Values{}
			params.Add("page", "1")

			qs, err := rmqc.PagedListQueuesWithParameters(params)
			Ω(err).Should(BeNil())

			q := qs.Items[0]
			Ω(q.Name).ShouldNot(Equal(""))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Durable).ShouldNot(BeNil())
			Ω(q.Status).ShouldNot(BeEmpty())
			Ω(qs.Page).Should(Equal(1))
			Ω(qs.PageCount).Should(Equal(1))
			Ω(qs.ItemCount).ShouldNot(BeNil())
			Ω(qs.PageSize).Should(Equal(100))
			Ω(qs.TotalCount).ShouldNot(BeNil())
			Ω(qs.FilteredCount).ShouldNot(BeNil())
		})
	})

	Context("GET /queues/{vhost}", func() {
		It("returns decoded response", func() {
			conn := openConnection("rabbit/hole")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			_, err = ch.QueueDeclare(
				"q2",  // name
				false, // durable
				false, // auto delete
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			qs, err := rmqc.ListQueuesIn("rabbit/hole")
			Ω(err).Should(BeNil())

			q := FindQueueByName(qs, "q2")
			Ω(q.Name).Should(Equal("q2"))
			Ω(q.Vhost).Should(Equal("rabbit/hole"))
			Ω(q.Durable).Should(Equal(false))
			Ω(q.Status).ShouldNot(BeEmpty())
		})
	})

	Context("GET /queues/{vhost}/{name}", func() {
		It("returns decoded response", func() {
			conn := openConnection("rabbit/hole")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			_, err = ch.QueueDeclare(
				"q3",  // name
				false, // durable
				false, // auto delete
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			q, err := rmqc.GetQueue("rabbit/hole", "q3")
			Ω(err).Should(BeNil())

			Ω(q.Vhost).Should(Equal("rabbit/hole"))
			Ω(q.Durable).Should(Equal(false))
			Ω(q.Status).ShouldNot(BeEmpty())
		})
	})

	Context("DELETE /queues/{vhost}/{name}", func() {
		It("deletes a queue", func() {
			conn := openConnection("rabbit/hole")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			q, err := ch.QueueDeclare(
				"",    // name
				false, // durable
				false, // auto delete
				false, // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			_, err = rmqc.GetQueue("rabbit/hole", q.Name)
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteQueue("rabbit/hole", q.Name)
			Ω(err).Should(BeNil())

			qi2, err := rmqc.GetQueue("rabbit/hole", q.Name)
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
			Ω(qi2).Should(BeNil())
		})
	})

	Context("GET /consumers", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			_, err = ch.QueueDeclare(
				"",    // name
				false, // durable
				false, // auto delete
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			_, err = ch.Consume(
				"",    // queue
				"",    // consumer
				false, // auto ack
				false, // exclusive
				false, // no local
				false, // no wait
				amqp.Table{})
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			cs, err := rmqc.ListConsumers()
			Ω(err).Should(BeNil())

			Ω(len(cs)).Should(Equal(1))
			c := cs[0]
			Ω(c.Queue.Name).ShouldNot(Equal(""))
			Ω(c.ConsumerTag).ShouldNot(Equal(""))
			Ω(c.Exclusive).ShouldNot(BeNil())
			Ω(c.AcknowledgementMode).Should(Equal(ManualAcknowledgement))
		})
	})

	Context("GET /consumers/{vhost}", func() {
		It("returns decoded response", func() {
			conn := openConnection("rabbit/hole")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			_, err = ch.QueueDeclare(
				"",    // name
				false, // durable
				false, // auto delete
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			_, err = ch.Consume(
				"",    // queue
				"",    // consumer
				true,  // auto ack
				false, // exclusive
				false, // no local
				false, // no wait
				amqp.Table{})
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			cs, err := rmqc.ListConsumers()
			Ω(err).Should(BeNil())

			Ω(len(cs)).Should(Equal(1))
			c := cs[0]
			Ω(c.Queue.Name).ShouldNot(Equal(""))
			Ω(c.ConsumerTag).ShouldNot(Equal(""))
			Ω(c.Exclusive).ShouldNot(BeNil())
			Ω(c.AcknowledgementMode).Should(Equal(AutomaticAcknowledgment))
		})
	})

	Context("GET /users", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListUsers()
			Ω(err).Should(BeNil())

			u := FindUserByName(xs, "guest")
			Ω(u.Name).Should(BeEquivalentTo("guest"))
			Ω(u.PasswordHash).ShouldNot(BeNil())

			tags := UserTags([]string{"administrator"})
			Ω(u.Tags).Should(Equal(tags))
		})
	})

	Context("GET /users/{name} when user exists", func() {
		It("returns decoded response", func() {
			u, err := rmqc.GetUser("guest")
			Ω(err).Should(BeNil())

			Ω(u.Name).Should(BeEquivalentTo("guest"))
			Ω(u.PasswordHash).ShouldNot(BeNil())

			tags := UserTags([]string{"administrator"})
			Ω(u.Tags).Should(Equal(tags))
		})
	})

	Context("PUT /users/{name}", func() {
		It("updates the user", func() {
			username := "rabbithole"
			_, err := rmqc.DeleteUser(username)
			Ω(err).Should(BeNil())

			info := UserSettings{Password: "s3krE7", Tags: UserTags{"policymaker", "management"}}
			resp, err := rmqc.PutUser(username, info)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			u, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(BeNil())

			Ω(u.PasswordHash).ShouldNot(BeNil())
			tags := UserTags([]string{"policymaker", "management"})
			Ω(u.Tags).Should(Equal(tags))

			// cleanup
			_, err = rmqc.DeleteUser(username)
			Ω(err).Should(BeNil())
		})

		It("updates the user with a password hash and hashing function", func() {
			username := "rabbithole_with_hashed_password1"
			_, err := rmqc.DeleteUser(username)
			Ω(err).Should(BeNil())

			tags := UserTags{"policymaker", "management"}
			password := "s3krE7-s3t-v!a-4A$h"
			info := UserSettings{PasswordHash: Base64EncodedSaltedPasswordHashSHA256(password),
				HashingAlgorithm: HashingAlgorithmSHA256,
				Tags:             tags}
			resp, err := rmqc.PutUser(username, info)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			permissions := Permissions{Configure: "log.*", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdatePermissionsIn("/", username, permissions)
			Ω(err).Should(BeNil())

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			u, err := rmqc.GetUser(username)
			Ω(err).Should(BeNil())

			Ω(u.PasswordHash).ShouldNot(BeNil())
			Ω(u.PasswordHash).ShouldNot(BeEquivalentTo(""))

			Ω(u.Tags).Should(Equal(tags))

			// make sure the user can successfully connect
			conn, err := amqp.Dial("amqp://" + username + ":" + password + "@localhost:5672/%2f")
			Ω(err).Should(BeNil())
			defer conn.Close()

			// cleanup
			_, err = rmqc.DeleteUser(username)
			Ω(err).Should(BeNil())
		})

		It("fails to update the user with incorrectly encoded password hash", func() {
			username := "rabbithole_with_hashed_password2"
			_, err := rmqc.DeleteUser(username)
			Ω(err).Should(BeNil())

			tags := UserTags{"policymaker", "management"}
			password := "s3krE7-s3t-v!a-4A$h"
			info := UserSettings{PasswordHash: password,
				HashingAlgorithm: HashingAlgorithmSHA256,
				Tags:             tags}
			_, err = rmqc.PutUser(username, info)
			Ω(err).Should(HaveOccurred())
			Ω(err.(ErrorResponse).StatusCode).Should(Equal(400))

			// cleanup
			_, err = rmqc.DeleteUser(username)
			Ω(err).Should(BeNil())
		})

		It("updates the user with no password", func() {
			tags := UserTags{"policymaker", "management"}
			info := UserSettings{Tags: tags}
			uname := "rabbithole-passwordless"
			rmqc.DeleteUser(uname)
			resp, err := rmqc.PutUserWithoutPassword(uname, info)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			// give internal events a moment to be
			// handled
			awaitEventPropagation()

			u, err := rmqc.GetUser(uname)
			Ω(err).Should(BeNil())

			Ω(u.PasswordHash).Should(BeEquivalentTo(""))
			Ω(u.Tags).Should(Equal(tags))

			// cleanup
			_, err = rmqc.DeleteUser(uname)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /users/{name}", func() {
		It("deletes the user", func() {
			info := UserSettings{Password: "s3krE7", Tags: UserTags{"management", "policymaker"}}
			_, err := rmqc.PutUser("rabbithole", info)
			Ω(err).Should(BeNil())

			u, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(BeNil())
			Ω(u).ShouldNot(BeNil())

			resp, err := rmqc.DeleteUser("rabbithole")
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			awaitEventPropagation()

			u2, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
			Ω(u2).Should(BeNil())
		})
	})

	Context("GET /vhosts", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListVhosts()
			Ω(err).Should(BeNil())

			x := xs[0]
			Ω(x.Name).ShouldNot(BeNil())
			Ω(x.Tracing).ShouldNot(BeNil())
			Ω(x.ClusterState).ShouldNot(BeNil())
		})
	})

	Context("GET /vhosts/{name} when vhost exists", func() {
		It("returns decoded response", func() {
			x, err := rmqc.GetVhost("rabbit/hole")
			Ω(err).Should(BeNil())

			Ω(x.Name).ShouldNot(BeNil())
			Ω(x.Tracing).ShouldNot(BeNil())
			Ω(x.ClusterState).ShouldNot(BeNil())
		})
	})

	Context("PUT /vhosts/{name}", func() {
		It("creates a vhost", func() {
			vs := VhostSettings{Tracing: false}
			_, err := rmqc.PutVhost("rabbit/hole2", vs)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			x, err := rmqc.GetVhost("rabbit/hole2")
			Ω(err).Should(BeNil())

			Ω(x.Name).Should(BeEquivalentTo("rabbit/hole2"))
			Ω(x.Tracing).Should(Equal(false))

			_, err = rmqc.DeleteVhost("rabbit/hole2")
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /vhosts/{name}", func() {
		It("deletes a vhost", func() {
			vs := VhostSettings{Tracing: false}
			_, err := rmqc.PutVhost("rabbit/hole2", vs)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			x, err := rmqc.GetVhost("rabbit/hole2")
			Ω(err).Should(BeNil())
			Ω(x).ShouldNot(BeNil())

			_, err = rmqc.DeleteVhost("rabbit/hole2")
			Ω(err).Should(BeNil())

			x2, err := rmqc.GetVhost("rabbit/hole2")
			Ω(x2).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("GET /bindings", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			q, _ := ch.QueueDeclare(
				"",
				false,
				false,
				true,
				false,
				nil)
			err = ch.QueueBind(q.Name, "", "amq.topic", false, nil)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			bs, err := rmqc.ListBindings()
			Ω(err).Should(BeNil())
			Ω(bs).ShouldNot(BeEmpty())

			b := bs[0]

			Ω(b.Source).ShouldNot(BeNil())
			Ω(b.Destination).ShouldNot(BeNil())
			Ω(b.Vhost).ShouldNot(BeNil())
			Ω([]string{"queue", "exchange"}).Should(ContainElement(b.DestinationType))
		})
	})

	Context("GET /bindings/{vhost}", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			q, _ := ch.QueueDeclare(
				"",
				false,
				false,
				true,
				false,
				nil)
			err = ch.QueueBind(q.Name, "", "amq.topic", false, nil)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			bs, err := rmqc.ListBindingsIn("/")
			Ω(err).Should(BeNil())
			Ω(bs).ShouldNot(BeEmpty())

			b := bs[0]
			Ω(b.Source).ShouldNot(BeNil())
			Ω(b.Destination).ShouldNot(BeNil())
			Ω(b.Vhost).Should(Equal("/"))
		})
	})

	Context("GET /queues/{vhost}/{queue}/bindings", func() {
		It("returns decoded response", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())

			q, err := ch.QueueDeclare(
				"",    // name
				false, // durable
				false, // auto delete
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			bs, err := rmqc.ListQueueBindings("/", q.Name)
			Ω(err).Should(BeNil())

			b := bs[0]
			Ω(b.Source).Should(Equal(""))
			Ω(b.Destination).Should(Equal(q.Name))
			Ω(b.RoutingKey).Should(Equal(q.Name))
			Ω(b.Vhost).Should(Equal("/"))

			ch.Close()
		})
	})

	Context("POST /bindings/{vhost}/e/{source}/q/{destination}", func() {
		It("adds a binding to a queue", func() {
			vh := "rabbit/hole"
			qn := "test.bindings.post.queue"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{})
			Ω(err).Should(BeNil())

			info := BindingInfo{
				Source:          "amq.topic",
				Destination:     qn,
				DestinationType: "queue",
				RoutingKey:      "#",
				Arguments: map[string]interface{}{
					"one": "two",
				},
			}

			res, err := rmqc.DeclareBinding(vh, info)
			Ω(err).Should(BeNil())

			// Grab the Location data from the POST response {destination}/{propertiesKey}
			propertiesKey, _ := url.QueryUnescape(strings.Split(res.Header.Get("Location"), "/")[1])

			awaitEventPropagation()
			bs, err := rmqc.ListQueueBindings(vh, qn)
			Ω(err).Should(BeNil())
			Ω(bs).ShouldNot(BeEmpty())

			b := bs[1]
			Ω(b.Source).Should(Equal(info.Source))
			Ω(b.Vhost).Should(Equal(vh))
			Ω(b.Destination).Should(Equal(info.Destination))
			Ω(b.DestinationType).Should(Equal(info.DestinationType))
			Ω(b.RoutingKey).Should(Equal(info.RoutingKey))
			Ω(b.PropertiesKey).Should(Equal(propertiesKey))

			_, err = rmqc.DeleteBinding(vh, b)
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /bindings/{vhost}/e/{exchange}/q/{queue}", func() {
		It("returns a list of bindings between the exchange and the queue", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())

			vh := conn.Config.Vhost
			x := "test.bindings.x30"
			q := "test.bindings.q%7C30"

			err = ch.ExchangeDeclare(x, "topic", false, false, false, false, nil)
			Ω(err).Should(BeNil())
			_, err = ch.QueueDeclare(q, false, false, false, false, nil)
			Ω(err).Should(BeNil())

			err = ch.QueueBind(q, "#", x, false, nil)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			bindings, err := rmqc.ListQueueBindingsBetween(vh, x, q)
			Ω(err).Should(BeNil())

			b := bindings[0]
			Ω(b.Source).Should(Equal(x))
			Ω(b.Destination).Should(Equal(q))
			Ω(b.RoutingKey).Should(Equal("#"))
			Ω(b.Vhost).Should(Equal("/"))

			err = ch.ExchangeDelete(x, false, false)
			Ω(err).Should(BeNil())
			_, err = ch.QueueDelete(q, false, false, false)
			Ω(err).Should(BeNil())
			ch.Close()
		})
	})

	Context("GET /bindings/{vhost}/e/{source}/e/{destination}", func() {
		It("returns a list of E2E bindings between the exchanges", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())

			vh := conn.Config.Vhost
			sx := "test.bindings.exchanges.source"
			dx := "test.bindings.exchanges.destination"

			err = ch.ExchangeDeclare(sx, "topic", false, false, false, false, nil)
			Ω(err).Should(BeNil())
			err = ch.ExchangeDeclare(dx, "topic", false, false, false, false, nil)
			Ω(err).Should(BeNil())

			err = ch.ExchangeBind(dx, "#", sx, false, nil)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			bindings, err := rmqc.ListExchangeBindingsBetween(vh, sx, dx)
			Ω(err).Should(BeNil())

			b := bindings[0]
			Ω(b.Source).Should(Equal(sx))
			Ω(b.Destination).Should(Equal(dx))
			Ω(b.RoutingKey).Should(Equal("#"))
			Ω(b.Vhost).Should(Equal("/"))

			err = ch.ExchangeDelete(sx, false, false)
			Ω(err).Should(BeNil())
			err = ch.ExchangeDelete(dx, false, false)
			Ω(err).Should(BeNil())
			ch.Close()
		})
	})

	Context("GET /exchanges/{vhost}/{exchange}/bindings/{vertex}", func() {
		It("returns a list of E2E bindings where the exchange is the source", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())

			vh := conn.Config.Vhost
			sx := "test.bindings.exchanges.source"
			dx := "test.bindings.exchanges.destination"

			err = ch.ExchangeDeclare(sx, "topic", false, false, false, false, nil)
			Ω(err).Should(BeNil())
			err = ch.ExchangeDeclare(dx, "topic", false, false, false, false, nil)
			Ω(err).Should(BeNil())

			err = ch.ExchangeBind(dx, "#", sx, false, nil)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			bindings, err := rmqc.ListExchangeBindingsWithSource(vh, sx)
			Ω(err).Should(BeNil())

			b := bindings[0]
			Ω(b.Source).Should(Equal(sx))
			Ω(b.Destination).Should(Equal(dx))
			Ω(b.RoutingKey).Should(Equal("#"))
			Ω(b.Vhost).Should(Equal("/"))

			err = ch.ExchangeDelete(sx, false, false)
			Ω(err).Should(BeNil())
			err = ch.ExchangeDelete(dx, false, false)
			Ω(err).Should(BeNil())
			ch.Close()
		})

		It("returns a list of E2E bindings where the exchange is the destination", func() {
			conn := openConnection("/")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())

			vh := conn.Config.Vhost
			sx := "test.bindings.exchanges.source"
			dx := "test.bindings.exchanges.destination"

			err = ch.ExchangeDeclare(sx, "topic", false, false, false, false, nil)
			Ω(err).Should(BeNil())
			err = ch.ExchangeDeclare(dx, "topic", false, false, false, false, nil)
			Ω(err).Should(BeNil())

			err = ch.ExchangeBind(dx, "#", sx, false, nil)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			bindings, err := rmqc.ListExchangeBindingsWithDestination(vh, dx)
			Ω(err).Should(BeNil())

			b := bindings[0]
			Ω(b.Source).Should(Equal(sx))
			Ω(b.Destination).Should(Equal(dx))
			Ω(b.RoutingKey).Should(Equal("#"))
			Ω(b.Vhost).Should(Equal("/"))

			err = ch.ExchangeDelete(sx, false, false)
			Ω(err).Should(BeNil())
			err = ch.ExchangeDelete(dx, false, false)
			Ω(err).Should(BeNil())
			ch.Close()
		})
	})

	Context("POST /bindings/{vhost}/e/{source}/e/{destination}", func() {
		It("adds a binding to an exchange", func() {
			vh := "rabbit/hole"
			xn := "test.bindings.post.exchange"

			_, err := rmqc.DeclareExchange(vh, xn, ExchangeSettings{Type: "topic"})
			Ω(err).Should(BeNil())

			info := BindingInfo{
				Source:          "amq.topic",
				Destination:     xn,
				DestinationType: "exchange",
				RoutingKey:      "#",
				Arguments: map[string]interface{}{
					"one": "two",
				},
			}

			res, err := rmqc.DeclareBinding(vh, info)
			Ω(err).Should(BeNil())

			// Grab the Location data from the POST response {destination}/{propertiesKey}
			propertiesKey, _ := url.QueryUnescape(strings.Split(res.Header.Get("Location"), "/")[1])

			awaitEventPropagation()
			bs, err := rmqc.ListExchangeBindingsBetween(vh, "amq.topic", xn)
			Ω(err).Should(BeNil())
			Ω(bs).ShouldNot(BeEmpty())

			b := bs[0]
			Ω(b.Source).Should(Equal(info.Source))
			Ω(b.Vhost).Should(Equal(vh))
			Ω(b.Destination).Should(Equal(info.Destination))
			Ω(b.DestinationType).Should(Equal(info.DestinationType))
			Ω(b.RoutingKey).Should(Equal(info.RoutingKey))
			Ω(b.PropertiesKey).Should(Equal(propertiesKey))

			// cleanup
			_, err = rmqc.DeleteBinding(vh, b)
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteExchange(vh, xn)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /bindings/{vhost}/e/{source}/e/{destination}/{propertiesKey}", func() {
		It("deletes an individual exchange binding", func() {
			vh := "rabbit/hole"
			xn := "test.bindings.post.exchange"

			_, err := rmqc.DeclareExchange(vh, xn, ExchangeSettings{Type: "topic"})
			Ω(err).Should(BeNil())

			info := BindingInfo{
				Source:          "amq.topic",
				Destination:     xn,
				DestinationType: "exchange",
				RoutingKey:      "#",
				Arguments: map[string]interface{}{
					"one": "two",
				},
			}

			res, err := rmqc.DeclareBinding(vh, info)
			Ω(err).Should(BeNil())

			awaitEventPropagation()

			// Grab the Location data from the POST response {destination}/{propertiesKey}
			propertiesKey, _ := url.QueryUnescape(strings.Split(res.Header.Get("Location"), "/")[1])
			info.PropertiesKey = propertiesKey

			_, err = rmqc.DeleteBinding(vh, info)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			bs, err := rmqc.ListExchangeBindingsWithDestination(vh, xn)
			Ω(bs).Should(BeEmpty())
			Ω(err).Should(BeNil())

			// cleanup
			_, err = rmqc.DeleteExchange(vh, xn)
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /permissions", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListPermissions()
			Ω(err).Should(BeNil())

			x := xs[0]
			Ω(x.User).ShouldNot(BeNil())
			Ω(x.Vhost).ShouldNot(BeNil())

			Ω(x.Configure).ShouldNot(BeNil())
			Ω(x.Write).ShouldNot(BeNil())
			Ω(x.Read).ShouldNot(BeNil())
		})
	})

	Context("GET /users/{name}/permissions", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListPermissionsOf("guest")
			Ω(err).Should(BeNil())

			x := xs[0]
			Ω(x.User).ShouldNot(BeNil())
			Ω(x.Vhost).ShouldNot(BeNil())

			Ω(x.Configure).ShouldNot(BeNil())
			Ω(x.Write).ShouldNot(BeNil())
			Ω(x.Read).ShouldNot(BeNil())
		})
	})

	Context("GET /permissions/{vhost}/{user}", func() {
		It("returns decoded response", func() {
			x, err := rmqc.GetPermissionsIn("/", "guest")
			Ω(err).Should(BeNil())

			Ω(x.User).Should(Equal("guest"))
			Ω(x.Vhost).Should(Equal("/"))

			Ω(x.Configure).Should(Equal(".*"))
			Ω(x.Write).Should(Equal(".*"))
			Ω(x.Read).Should(Equal(".*"))
		})
	})

	Context("PUT /permissions/{vhost}/{user}", func() {
		It("updates permissions", func() {
			u := "temporary"

			_, err := rmqc.PutUser(u, UserSettings{Password: "s3krE7"})
			Ω(err).Should(BeNil())

			permissions := Permissions{Configure: "log.*", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdatePermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			fetched, err := rmqc.GetPermissionsIn("/", u)
			Ω(err).Should(BeNil())
			Ω(fetched.Configure).Should(Equal(permissions.Configure))
			Ω(fetched.Write).Should(Equal(permissions.Write))
			Ω(fetched.Read).Should(Equal(permissions.Read))

			_, err = rmqc.DeleteUser(u)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /permissions/{vhost}/{user}", func() {
		It("clears permissions", func() {
			u := "temporary"

			_, err := rmqc.PutUser(u, UserSettings{Password: "s3krE7"})
			Ω(err).Should(BeNil())

			permissions := Permissions{Configure: "log.*", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdatePermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())

			awaitEventPropagation()

			_, err = rmqc.ClearPermissionsIn("/", u)
			Ω(err).Should(BeNil())
			awaitEventPropagation()
			_, err = rmqc.GetPermissionsIn("/", u)
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))

			_, err = rmqc.DeleteUser(u)
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /topic-permissions", func() {
		It("returns decoded response", func() {
			u := "temporary"

			_, err := rmqc.PutUser(u, UserSettings{Password: "s3krE7"})
			Ω(err).Should(BeNil())

			permissions := TopicPermissions{Exchange: "amq.topic", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdateTopicPermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())

			awaitEventPropagation()

			xs, err := rmqc.ListTopicPermissions()
			Ω(err).Should(BeNil())

			x := xs[0]
			Ω(x.User).ShouldNot(BeNil())
			Ω(x.Vhost).ShouldNot(BeNil())

			Ω(x.Exchange).ShouldNot(BeNil())
			Ω(x.Write).ShouldNot(BeNil())
			Ω(x.Read).ShouldNot(BeNil())

			_, err = rmqc.DeleteUser(u)
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /users/{name}/topic-permissions", func() {
		It("returns decoded response", func() {
			u := "temporary"

			_, err := rmqc.PutUser(u, UserSettings{Password: "s3krE7"})
			Ω(err).Should(BeNil())

			permissions := TopicPermissions{Exchange: "amq.topic", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdateTopicPermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())

			awaitEventPropagation()

			xs, err := rmqc.ListTopicPermissionsOf("temporary")
			Ω(err).Should(BeNil())

			x := xs[0]
			Ω(x.User).ShouldNot(BeNil())
			Ω(x.Vhost).ShouldNot(BeNil())

			Ω(x.Exchange).ShouldNot(BeNil())
			Ω(x.Write).ShouldNot(BeNil())
			Ω(x.Read).ShouldNot(BeNil())

			_, err = rmqc.DeleteUser(u)
			Ω(err).Should(BeNil())
		})
	})

	Context("PUT /topic-permissions/{vhost}/{user}", func() {
		It("updates topic permissions", func() {
			u := "temporary"

			_, err := rmqc.PutUser(u, UserSettings{Password: "s3krE7"})
			Ω(err).Should(BeNil())

			permissions := TopicPermissions{Exchange: "amq.topic", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdateTopicPermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			fetched, err := rmqc.GetTopicPermissionsIn("/", u)
			Ω(err).Should(BeNil())
			x := fetched[0]
			Ω(x.Exchange).Should(Equal(permissions.Exchange))
			Ω(x.Write).Should(Equal(permissions.Write))
			Ω(x.Read).Should(Equal(permissions.Read))

			_, err = rmqc.DeleteUser(u)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /topic-permissions/{vhost}/{user}", func() {
		It("clears topic permissions", func() {
			u := "temporary"

			_, err := rmqc.PutUser(u, UserSettings{Password: "s3krE7"})
			Ω(err).Should(BeNil())

			permissions := TopicPermissions{Exchange: "amq.topic", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdateTopicPermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())

			awaitEventPropagation()

			_, err = rmqc.ClearTopicPermissionsIn("/", u)
			Ω(err).Should(BeNil())
			awaitEventPropagation()
			_, err = rmqc.GetTopicPermissionsIn("/", u)
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))

			_, err = rmqc.DeleteUser(u)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /topic-permissions/{vhost}/{user}/{exchange}", func() {
		It("deletes one topic permissions", func() {
			u := "temporary"

			_, err := rmqc.PutUser(u, UserSettings{Password: "s3krE7"})
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			permissions := TopicPermissions{Exchange: "amq.topic", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdateTopicPermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())
			permissions = TopicPermissions{Exchange: "foobar", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdateTopicPermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			_, err = rmqc.DeleteTopicPermissionsIn("/", u, "foobar")
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			xs, err := rmqc.ListTopicPermissionsOf(u)
			Ω(err).Should(BeNil())

			Ω(len(xs)).Should(BeEquivalentTo(1))

			_, err = rmqc.DeleteUser(u)
			Ω(err).Should(BeNil())
		})
	})

	Context("PUT /exchanges/{vhost}/{exchange}", func() {
		It("declares an exchange", func() {
			vh := "rabbit/hole"
			xn := "temporary"

			_, err := rmqc.DeclareExchange(vh, xn, ExchangeSettings{Type: "fanout", Durable: false})
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			x, err := rmqc.GetExchange(vh, xn)
			Ω(err).Should(BeNil())
			Ω(x.Name).Should(Equal(xn))
			Ω(x.Durable).Should(Equal(false))
			Ω(x.AutoDelete).Should(Equal(false))
			Ω(x.Type).Should(Equal("fanout"))
			Ω(x.Vhost).Should(Equal(vh))

			_, err = rmqc.DeleteExchange(vh, xn)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /exchanges/{vhost}/{exchange}", func() {
		It("deletes an exchange", func() {
			vh := "rabbit/hole"
			xn := "temporary"

			_, err := rmqc.DeclareExchange(vh, xn, ExchangeSettings{Type: "fanout", Durable: false})
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteExchange(vh, xn)
			Ω(err).Should(BeNil())

			awaitEventPropagation()

			x, err := rmqc.GetExchange(vh, xn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("PUT /queues/{vhost}/{queue}", func() {
		It("declares a queue", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			x, err := rmqc.GetQueue(vh, qn)
			Ω(err).Should(BeNil())
			Ω(x.Name).Should(Equal(qn))
			Ω(x.Durable).Should(Equal(false))
			Ω(x.AutoDelete).Should(Equal(false))
			Ω(x.Vhost).Should(Equal(vh))

			_, err = rmqc.DeleteQueue(vh, qn)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /queues/{vhost}/{queue}", func() {
		It("deletes a queue", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			_, err = rmqc.DeleteQueue(vh, qn)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			x, err := rmqc.GetQueue(vh, qn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})

		It("accepts IfEmpty option", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			_, err = rmqc.DeleteQueue(vh, qn, QueueDeleteOptions{IfEmpty: true})
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			x, err := rmqc.GetQueue(vh, qn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})

		It("accepts IfUnused option", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			_, err = rmqc.DeleteQueue(vh, qn, QueueDeleteOptions{IfUnused: true})
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			x, err := rmqc.GetQueue(vh, qn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("DELETE /queues/{vhost}/{queue}/contents", func() {
		It("purges a queue", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			_, err = rmqc.PurgeQueue(vh, qn)
			awaitEventPropagation()
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteQueue(vh, qn)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			x, err := rmqc.GetQueue(vh, qn)
			awaitEventPropagation()
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("POST /queues/{vhost}/{queue}/actions", func() {
		It("synchronises queue", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			// it would be better to test this in a cluster configuration
			x, err := rmqc.SyncQueue(vh, qn)
			Ω(err).Should(BeNil())
			Ω(x.StatusCode).Should(Equal(204))
			_, err = rmqc.DeleteQueue(vh, qn)
			Ω(err).Should(BeNil())
		})

		It("cancels queue synchronisation", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			// it would be better to test this in a cluster configuration
			x, err := rmqc.CancelSyncQueue(vh, qn)
			Ω(err).Should(BeNil())
			Ω(x.StatusCode).Should(Equal(204))
			_, err = rmqc.DeleteQueue(vh, qn)
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /policies", func() {
		Context("when policy exists", func() {
			It("returns decoded response", func() {
				policy1 := Policy{
					Pattern:    "abc",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100, "ha-mode": "all"},
					Priority:   0,
				}

				policy2 := Policy{
					Pattern:    ".*",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100, "ha-mode": "all"},
					Priority:   0,
				}

				// prepare policies
				_, err := rmqc.PutPolicy("rabbit/hole", "woot1", policy1)
				Ω(err).Should(BeNil())

				_, err = rmqc.PutPolicy("rabbit/hole", "woot2", policy2)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				// test
				pols, err := rmqc.ListPolicies()
				Ω(err).Should(BeNil())
				Ω(pols).ShouldNot(BeEmpty())
				Ω(len(pols)).Should(BeNumerically(">=", 2))
				Ω(pols[0].Name).ShouldNot(BeNil())
				Ω(pols[1].Name).ShouldNot(BeNil())

				// cleanup
				_, err = rmqc.DeletePolicy("rabbit/hole", "woot1")
				Ω(err).Should(BeNil())

				_, err = rmqc.DeletePolicy("rabbit/hole", "woot2")
				Ω(err).Should(BeNil())
			})
		})
	})

	Context("GET /polices/{vhost}", func() {
		Context("when policy exists", func() {
			It("returns decoded response", func() {
				policy1 := Policy{
					Pattern:    "abc",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100, "ha-mode": "all"},
					Priority:   0,
				}

				policy2 := Policy{
					Pattern:    ".*",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100, "ha-mode": "all"},
					Priority:   0,
				}

				// prepare policies
				_, err := rmqc.PutPolicy("rabbit/hole", "woot1", policy1)
				Ω(err).Should(BeNil())

				_, err = rmqc.PutPolicy("/", "woot2", policy2)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				// test
				pols, err := rmqc.ListPoliciesIn("rabbit/hole")
				Ω(err).Should(BeNil())
				Ω(pols).ShouldNot(BeEmpty())
				Ω(len(pols)).Should(Equal(1))
				Ω(pols[0].Name).Should(Equal("woot1"))

				// cleanup
				_, err = rmqc.DeletePolicy("rabbit/hole", "woot1")
				Ω(err).Should(BeNil())

				_, err = rmqc.DeletePolicy("/", "woot2")
				Ω(err).Should(BeNil())
			})
		})

		Context("when no policies exist", func() {
			It("returns decoded response", func() {
				pols, err := rmqc.ListPoliciesIn("rabbit/hole")
				Ω(err).Should(BeNil())
				Ω(pols).Should(BeEmpty())
			})
		})
	})

	Context("GET /policies/{vhost}/{name}", func() {
		Context("when policy exists", func() {
			It("returns decoded response", func() {
				policy := Policy{
					Pattern:    ".*",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100, "ha-mode": "all"},
					Priority:   0,
				}

				_, err := rmqc.PutPolicy("rabbit/hole", "woot", policy)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				pol, err := rmqc.GetPolicy("rabbit/hole", "woot")
				Ω(err).Should(BeNil())
				Ω(pol.Vhost).Should(Equal("rabbit/hole"))
				Ω(pol.Name).Should(Equal("woot"))
				Ω(pol.ApplyTo).Should(Equal("all"))
				Ω(pol.Pattern).Should(Equal(".*"))
				Ω(pol.Priority).Should(BeEquivalentTo(0))
				Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
				Ω(pol.Definition["expires"]).Should(BeEquivalentTo(100))
				Ω(pol.Definition["ha-mode"]).Should(Equal("all"))

				_, err = rmqc.DeletePolicy("rabbit/hole", "woot")
				Ω(err).Should(BeNil())
			})
		})

		Context("when policy not found", func() {
			It("returns decoded response", func() {
				pol, err := rmqc.GetPolicy("rabbit/hole", "woot")
				Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
				Ω(pol).Should(BeNil())
			})
		})
	})

	Context("DELETE /polices/{vhost}/{name}", func() {
		It("deletes the policy", func() {
			policy := Policy{
				Pattern:    ".*",
				ApplyTo:    "all",
				Definition: PolicyDefinition{"expires": 100, "ha-mode": "all"},
				Priority:   0,
			}

			_, err := rmqc.PutPolicy("rabbit/hole", "woot", policy)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			resp, err := rmqc.DeletePolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))
		})
	})

	Context("PUT /policies/{vhost}/{name}", func() {
		It("creates the policy", func() {
			policy := Policy{
				Pattern: ".*",
				ApplyTo: "all",
				Definition: PolicyDefinition{
					"expires":   100,
					"ha-mode":   "nodes",
					"ha-params": NodeNames{"a", "b", "c"},
				},
				Priority: 0,
			}

			resp, err := rmqc.PutPolicy("rabbit/hole", "woot", policy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			awaitEventPropagation()
			_, err = rmqc.GetPolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())

			_, err = rmqc.DeletePolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())
		})

		It("updates the policy", func() {
			policy := Policy{
				Pattern:    ".*",
				ApplyTo:    "exchanges",
				Definition: PolicyDefinition{"expires": 100, "ha-mode": "all"},
			}

			// create policy
			resp, err := rmqc.PutPolicy("rabbit/hole", "woot", policy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			awaitEventPropagation()

			pol, err := rmqc.GetPolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())
			Ω(pol.Vhost).Should(Equal("rabbit/hole"))
			Ω(pol.Name).Should(Equal("woot"))
			Ω(pol.Pattern).Should(Equal(".*"))
			Ω(pol.ApplyTo).Should(Equal("exchanges"))
			Ω(pol.Priority).Should(BeEquivalentTo(0))
			Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
			Ω(pol.Definition["ha-mode"]).Should(Equal("all"))
			Ω(pol.Definition["expires"]).Should(BeEquivalentTo(100))

			// update the policy
			newPolicy := Policy{
				Pattern: "\\d+",
				ApplyTo: "all",
				Definition: PolicyDefinition{
					"max-length": 100,
					"ha-mode":    "nodes",
					"ha-params":  NodeNames{"a", "b", "c"},
				},
				Priority: 1,
			}

			resp, err = rmqc.PutPolicy("rabbit/hole", "woot", newPolicy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			awaitEventPropagation()

			pol, err = rmqc.GetPolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())
			Ω(pol.Vhost).Should(Equal("rabbit/hole"))
			Ω(pol.Name).Should(Equal("woot"))
			Ω(pol.Pattern).Should(Equal("\\d+"))
			Ω(pol.ApplyTo).Should(Equal("all"))
			Ω(pol.Priority).Should(BeEquivalentTo(1))
			Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
			Ω(pol.Definition["max-length"]).Should(BeEquivalentTo(100))
			Ω(pol.Definition["ha-mode"]).Should(Equal("nodes"))
			Ω(pol.Definition["ha-params"]).Should(HaveLen(3))
			Ω(pol.Definition["ha-params"]).Should(ContainElement("a"))
			Ω(pol.Definition["ha-params"]).Should(ContainElement("c"))
			Ω(pol.Definition["expires"]).Should(BeNil())

			// cleanup
			_, err = rmqc.DeletePolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /api/parameters/federation-upstream", func() {
		Context("when there are no upstreams", func() {
			It("returns an empty response", func() {
				list, err := rmqc.ListFederationUpstreams()
				Ω(err).Should(BeNil())
				Ω(list).Should(BeEmpty())
			})
		})

		Context("when there are upstreams", func() {
			It("returns the list of upstreams", func() {
				def1 := FederationDefinition{
					Uri: "amqp://server-name/%2f",
				}
				_, err := rmqc.PutFederationUpstream("rabbit/hole", "upstream1", def1)
				Ω(err).Should(BeNil())

				def2 := FederationDefinition{
					Uri: "amqp://example.com/%2f",
				}
				_, err = rmqc.PutFederationUpstream("/", "upstream2", def2)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				list, err := rmqc.ListFederationUpstreams()
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(2))

				_, err = rmqc.DeleteFederationUpstream("rabbit/hole", "upstream1")
				Ω(err).Should(BeNil())

				_, err = rmqc.DeleteFederationUpstream("/", "upstream2")
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				list, err = rmqc.ListFederationUpstreams()
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(0))
			})
		})
	})

	Context("GET /api/parameters/federation-upstream/{vhost}", func() {
		Context("when there are no upstreams", func() {
			It("returns an empty response", func() {
				list, err := rmqc.ListFederationUpstreamsIn("rabbit/hole")
				Ω(err).Should(BeNil())
				Ω(list).Should(BeEmpty())
			})
		})

		Context("when there are upstreams", func() {
			It("returns the list of upstreams", func() {
				vh := "rabbit/hole"

				def1 := FederationDefinition{
					Uri: "amqp://server-name/%2f",
				}

				_, err := rmqc.PutFederationUpstream(vh, "upstream1", def1)
				Ω(err).Should(BeNil())

				def2 := FederationDefinition{
					Uri: "amqp://example.com/%2f",
				}

				_, err = rmqc.PutFederationUpstream(vh, "upstream2", def2)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				list, err := rmqc.ListFederationUpstreamsIn(vh)
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(2))

				// delete upstream1
				_, err = rmqc.DeleteFederationUpstream(vh, "upstream1")
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				list, err = rmqc.ListFederationUpstreamsIn(vh)
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(1))

				// delete upstream2
				_, err = rmqc.DeleteFederationUpstream(vh, "upstream2")
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				list, err = rmqc.ListFederationUpstreamsIn(vh)
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(0))
			})
		})
	})

	Context("GET /api/parameters/federation-upstream/{vhost}/{upstream}", func() {
		Context("when the upstream does not exist", func() {
			It("returns a 404 error", func() {
				vh := "rabbit/hole"
				name := "temporary"

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(up).Should(BeNil())
				Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
			})
		})

		Context("when the upstream exists", func() {
			It("returns the upstream", func() {
				vh := "rabbit/hole"
				name := "temporary"

				def := FederationDefinition{
					Uri:            "amqp://127.0.0.1/%2f",
					PrefetchCount:  1000,
					ReconnectDelay: 1,
					AckMode:        "on-confirm",
					TrustUserId:    false,
				}

				_, err := rmqc.PutFederationUpstream(vh, name, def)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))
				Ω(up.Component).Should(Equal(FederationUpstreamComponent))
				Ω(up.Definition.Uri).Should(Equal(def.Uri))
				Ω(up.Definition.PrefetchCount).Should(Equal(def.PrefetchCount))
				Ω(up.Definition.ReconnectDelay).Should(Equal(def.ReconnectDelay))
				Ω(up.Definition.AckMode).Should(Equal(def.AckMode))
				Ω(up.Definition.TrustUserId).Should(Equal(def.TrustUserId))

				_, err = rmqc.DeleteFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
			})
		})
	})

	Context("PUT /api/parameters/federation-upstream/{vhost}/{upstream}", func() {
		Context("when the upstream does not exist", func() {
			It("creates the upstream", func() {
				vh := "rabbit/hole"
				name := "temporary"

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(up).Should(BeNil())
				Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))

				def := FederationDefinition{
					Uri:            "amqp://127.0.0.1/%2f",
					Expires:        1800000,
					MessageTTL:     360000,
					MaxHops:        1,
					PrefetchCount:  500,
					ReconnectDelay: 5,
					AckMode:        "on-publish",
					TrustUserId:    false,
					Exchange:       "",
					Queue:          "",
				}

				_, err = rmqc.PutFederationUpstream(vh, name, def)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				up, err = rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))
				Ω(up.Component).Should(Equal(FederationUpstreamComponent))
				Ω(up.Definition.Uri).Should(Equal(def.Uri))
				Ω(up.Definition.Expires).Should(Equal(def.Expires))
				Ω(up.Definition.MessageTTL).Should(Equal(def.MessageTTL))
				Ω(up.Definition.MaxHops).Should(Equal(def.MaxHops))
				Ω(up.Definition.PrefetchCount).Should(Equal(def.PrefetchCount))
				Ω(up.Definition.ReconnectDelay).Should(Equal(def.ReconnectDelay))
				Ω(up.Definition.AckMode).Should(Equal(def.AckMode))
				Ω(up.Definition.TrustUserId).Should(Equal(def.TrustUserId))
				Ω(up.Definition.Exchange).Should(Equal(def.Exchange))
				Ω(up.Definition.Queue).Should(Equal(def.Queue))

				_, err = rmqc.DeleteFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
			})
		})

		Context("when the upstream exists", func() {
			It("updates the upstream", func() {
				vh := "rabbit/hole"
				name := "temporary"

				// create the upstream
				def := FederationDefinition{
					Uri:            "amqp://127.0.0.1/%2f",
					PrefetchCount:  1000,
					ReconnectDelay: 1,
					AckMode:        "on-confirm",
					TrustUserId:    false,
				}

				_, err := rmqc.PutFederationUpstream(vh, name, def)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))
				Ω(up.Component).Should(Equal(FederationUpstreamComponent))
				Ω(up.Definition.Uri).Should(Equal(def.Uri))
				Ω(up.Definition.PrefetchCount).Should(Equal(def.PrefetchCount))
				Ω(up.Definition.ReconnectDelay).Should(Equal(def.ReconnectDelay))
				Ω(up.Definition.AckMode).Should(Equal(def.AckMode))
				Ω(up.Definition.TrustUserId).Should(Equal(def.TrustUserId))

				// update the upstream
				def2 := FederationDefinition{
					Uri:            "amqp://127.0.0.1/%2f",
					PrefetchCount:  500,
					ReconnectDelay: 10,
					AckMode:        "no-ack",
					TrustUserId:    true,
				}

				_, err = rmqc.PutFederationUpstream(vh, name, def2)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				up, err = rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))
				Ω(up.Component).Should(Equal(FederationUpstreamComponent))
				Ω(up.Definition.Uri).Should(Equal(def2.Uri))
				Ω(up.Definition.PrefetchCount).Should(Equal(def2.PrefetchCount))
				Ω(up.Definition.ReconnectDelay).Should(Equal(def2.ReconnectDelay))
				Ω(up.Definition.AckMode).Should(Equal(def2.AckMode))
				Ω(up.Definition.TrustUserId).Should(Equal(def2.TrustUserId))

				_, err = rmqc.DeleteFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
			})
		})

		Context("when the upstream definition is bad", func() {
			It("returns a 400 error response", func() {
				resp, err := rmqc.PutFederationUpstream("rabbit/hole", "error", FederationDefinition{})
				Ω(resp).Should(BeNil())
				Ω(err).Should(HaveOccurred())

				e, ok := err.(ErrorResponse)
				Ω(ok).Should(BeTrue())
				Ω(e.StatusCode).Should(Equal(400))
				Ω(e.Message).Should(Equal("bad_request"))
				Ω(e.Reason).Should(MatchRegexp(`^Validation failed`))
			})
		})
	})

	Context("DELETE /api/parameters/federation-upstream/{vhost}/{name}", func() {
		Context("when the upstream does not exist", func() {
			It("returns a 404 error response", func() {
				vh := "rabbit/hole"
				name := "temporary"

				// an error is not returned by design
				resp, err := rmqc.DeleteFederationUpstream(vh, name)
				Ω(resp.Status).Should(Equal("404 Not Found"))
				Ω(err).Should(BeNil())
			})
		})

		Context("when the upstream exists", func() {
			It("deletes the upstream", func() {
				vh := "rabbit/hole"
				name := "temporary"

				def := FederationDefinition{
					Uri: "amqp://127.0.0.1/%2f",
				}

				_, err := rmqc.PutFederationUpstream(vh, name, def)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))

				_, err = rmqc.DeleteFederationUpstream(vh, name)
				Ω(err).Should(BeNil())

				up, err = rmqc.GetFederationUpstream(vh, name)
				Ω(up).Should(BeNil())
				Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
			})
		})
	})

	Context("GET /api/federation-links", func() {
		Context("when there are no links", func() {
			It("returns an empty response", func() {
				list, err := rmqc.ListFederationLinks()
				Ω(list).Should(BeEmpty())
				Ω(err).Should(BeNil())
			})
		})
	})

	Context("GET /api/federation-links/{vhost}", func() {
		Context("when there are no links", func() {
			It("returns an empty response", func() {
				list, err := rmqc.ListFederationLinksIn("rabbit/hole")
				Ω(list).Should(BeEmpty())
				Ω(err).Should(BeNil())
			})
		})

		Context("when there are links", func() {
			It("returns the list of links", func() {
				// upstream vhost: '/'
				// upstream exchange: amq.topic
				// downstream vhost: 'rabbit-hole'
				// downstream exchange: amq.topic
				vhost := "rabbit/hole"

				// create upstream
				upstreamName := "myUpsteam"
				def := FederationDefinition{
					Uri:     "amqp://localhost/%2f",
					Expires: 1800000,
				}

				_, err := rmqc.PutFederationUpstream(vhost, upstreamName, def)
				Ω(err).Should(BeNil())

				// create policy to match upstream
				policyName := "myUpsteamPolicy"
				policy := Policy{
					Pattern: "(^amq.topic$)",
					ApplyTo: "exchanges",
					Definition: PolicyDefinition{
						"federation-upstream": upstreamName,
					},
				}

				_, err = rmqc.PutPolicy(vhost, policyName, policy)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				// assertions
				list, err := rmqc.ListFederationLinksIn(vhost)
				Ω(len(list)).Should(Equal(1))
				Ω(err).Should(BeNil())

				var link map[string]interface{} = list[0]
				Ω(link["vhost"]).Should(Equal(vhost))
				Ω(link["upstream"]).Should(Equal(upstreamName))
				Ω(link["type"]).Should(Equal("exchange"))
				Ω(link["exchange"]).Should(Equal("amq.topic"))
				Ω(link["uri"]).Should(Equal(def.Uri))
				Ω(link["status"]).Should(Equal("running"))

				// cleanup
				_, err = rmqc.DeletePolicy(vhost, policyName)
				Ω(err).Should(BeNil())

				_, err = rmqc.DeleteFederationUpstream(vhost, upstreamName)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				list, err = rmqc.ListFederationLinksIn(vhost)
				Ω(len(list)).Should(Equal(0))
				Ω(err).Should(BeNil())
			})
		})
	})

	Context("PUT /parameters/shovel/{vhost}/{name}", func() {
		It("declares a shovel using AMQP 1.0 protocol", func() {
			vh := "rabbit/hole"
			sn := "temporary"

			ssu := "amqp://127.0.0.1/%2f"
			sdu := "amqp://127.0.0.1/%2f"

			shovelDefinition := ShovelDefinition{
				SourceURI:                     ssu,
				SourceAddress:                 "mySourceQueue",
				SourceProtocol:                "amqp10",
				DestinationURI:                sdu,
				DestinationProtocol:           "amqp10",
				DestinationAddress:            "myDestQueue",
				DestinationAddForwardHeaders:  true,
				DestinationAddTimestampHeader: true,
				AckMode:                       "on-confirm",
				SourcePrefetchCount:           42,
				SourceDeleteAfter:             "never"}

			_, err := rmqc.DeclareShovel(vh, sn, shovelDefinition)
			Ω(err).Should(BeNil())

			awaitEventPropagation()
			x, err := rmqc.GetShovel(vh, sn)
			Ω(err).Should(BeNil())
			Ω(x.Name).Should(Equal(sn))
			Ω(x.Vhost).Should(Equal(vh))
			Ω(x.Component).Should(Equal("shovel"))
			Ω(x.Definition.SourceAddress).Should(Equal("mySourceQueue"))
			Ω(x.Definition.SourceURI).Should(Equal(ssu))
			Ω(x.Definition.SourcePrefetchCount).Should(Equal(42))
			Ω(x.Definition.SourceProtocol).Should(Equal("amqp10"))
			Ω(x.Definition.DestinationAddress).Should(Equal("myDestQueue"))
			Ω(x.Definition.DestinationURI).Should(Equal(sdu))
			Ω(x.Definition.DestinationProtocol).Should(Equal("amqp10"))
			Ω(x.Definition.DestinationAddForwardHeaders).Should(Equal(true))
			Ω(x.Definition.DestinationAddTimestampHeader).Should(Equal(true))
			Ω(x.Definition.AckMode).Should(Equal("on-confirm"))
			Ω(x.Definition.SourceDeleteAfter).Should(Equal("never"))

			_, err = rmqc.DeleteShovel(vh, sn)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			_, err = rmqc.DeleteQueue("/", "mySourceQueue")
			Ω(err).Should(BeNil())
			_, err = rmqc.DeleteQueue("/", "myDestQueue")
			Ω(err).Should(BeNil())

			x, err = rmqc.GetShovel(vh, sn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("PUT /parameters/shovel/{vhost}/{name}", func() {
		It("declares a shovel", func() {
			vh := "rabbit/hole"
			sn := "temporary"

			ssu := "amqp://127.0.0.1/%2f"
			sdu := "amqp://127.0.0.1/%2f"

			shovelDefinition := ShovelDefinition{
				SourceURI:         ssu,
				SourceQueue:       "mySourceQueue",
				DestinationURI:    sdu,
				DestinationQueue:  "myDestQueue",
				AddForwardHeaders: true,
				AckMode:           "on-confirm",
				DeleteAfter:       "never"}

			_, err := rmqc.DeclareShovel(vh, sn, shovelDefinition)
			Ω(err).Should(BeNil(), "Error declaring shovel")

			awaitEventPropagation()
			x, err := rmqc.GetShovel(vh, sn)
			Ω(err).Should(BeNil(), "Error getting shovel")
			Ω(x.Name).Should(Equal(sn))
			Ω(x.Vhost).Should(Equal(vh))
			Ω(x.Component).Should(Equal("shovel"))
			Ω(x.Definition.SourceURI).Should(Equal(ssu))
			Ω(x.Definition.SourceQueue).Should(Equal("mySourceQueue"))
			Ω(x.Definition.DestinationURI).Should(Equal(sdu))
			Ω(x.Definition.DestinationQueue).Should(Equal("myDestQueue"))
			Ω(x.Definition.AddForwardHeaders).Should(Equal(true))
			Ω(x.Definition.AckMode).Should(Equal("on-confirm"))
			Ω(string(x.Definition.DeleteAfter)).Should(Equal("never"))

			_, err = rmqc.DeleteShovel(vh, sn)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			_, err = rmqc.DeleteQueue("/", "mySourceQueue")
			Ω(err).Should(BeNil())
			_, err = rmqc.DeleteQueue("/", "myDestQueue")
			Ω(err).Should(BeNil())

			x, err = rmqc.GetShovel(vh, sn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
		It("declares a shovel with a numeric delete-after value", func() {
			vh := "rabbit/hole"
			sn := "temporary"

			ssu := "amqp://127.0.0.1/%2f"
			sdu := "amqp://127.0.0.1/%2f"

			shovelDefinition := ShovelDefinition{
				SourceURI:         ssu,
				SourceQueue:       "mySourceQueue",
				DestinationURI:    sdu,
				DestinationQueue:  "myDestQueue",
				AddForwardHeaders: true,
				AckMode:           "on-confirm",
				DeleteAfter:       "42"}

			_, err := rmqc.DeclareShovel(vh, sn, shovelDefinition)
			Ω(err).Should(BeNil(), "Error declaring shovel")

			awaitEventPropagation()
			x, err := rmqc.GetShovel(vh, sn)
			Ω(err).Should(BeNil(), "Error getting shovel")
			Ω(x.Name).Should(Equal(sn))
			Ω(x.Vhost).Should(Equal(vh))
			Ω(x.Component).Should(Equal("shovel"))
			Ω(x.Definition.SourceURI).Should(Equal(ssu))
			Ω(x.Definition.SourceQueue).Should(Equal("mySourceQueue"))
			Ω(x.Definition.DestinationURI).Should(Equal(sdu))
			Ω(x.Definition.DestinationQueue).Should(Equal("myDestQueue"))
			Ω(x.Definition.AddForwardHeaders).Should(Equal(true))
			Ω(x.Definition.AckMode).Should(Equal("on-confirm"))
			Ω(string(x.Definition.DeleteAfter)).Should(Equal("42"))

			_, err = rmqc.DeleteShovel(vh, sn)
			Ω(err).Should(BeNil())
			awaitEventPropagation()

			_, err = rmqc.DeleteQueue("/", "mySourceQueue")
			Ω(err).Should(BeNil())
			_, err = rmqc.DeleteQueue("/", "myDestQueue")
			Ω(err).Should(BeNil())

			x, err = rmqc.GetShovel(vh, sn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("PUT /api/parameters/{component}/{vhost}/{name}", func() {
		Context("when the parameter does not exist", func() {
			It("creates the parameter", func() {
				component := FederationUpstreamComponent
				vhost := "rabbit/hole"
				name := "temporary"

				pv := RuntimeParameterValue{
					"uri":             "amqp://server-name",
					"prefetch-count":  500,
					"reconnect-delay": 5,
					"ack-mode":        "on-confirm",
					"trust-user-id":   false,
				}

				_, err := rmqc.PutRuntimeParameter(component, vhost, name, pv)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				p, err := rmqc.GetRuntimeParameter(component, vhost, name)

				Ω(err).Should(BeNil())
				Ω(p.Component).Should(Equal(FederationUpstreamComponent))
				Ω(p.Vhost).Should(Equal(vhost))
				Ω(p.Name).Should(Equal(name))

				// we need to convert from interface{}
				v, ok := p.Value.(map[string]interface{})
				Ω(ok).Should(BeTrue())
				Ω(v["uri"]).Should(Equal(pv["uri"]))
				Ω(v["prefetch-count"]).Should(BeNumerically("==", pv["prefetch-count"]))
				Ω(v["reconnect-delay"]).Should(BeNumerically("==", pv["reconnect-delay"]))

				_, err = rmqc.DeleteRuntimeParameter(component, vhost, name)
				Ω(err).Should(BeNil())
			})
		})
	})

	Context("GET /api/parameters", func() {
		Context("when there are no runtime parameters", func() {
			It("returns an empty response", func() {
				err := rmqc.DeleteAllRuntimeParameters()
				Ω(err).Should(BeNil())

				list, err := rmqc.ListRuntimeParameters()
				Ω(err).Should(BeNil())
				Ω(list).Should(BeEmpty())
			})
		})

		Context("when there are runtime parameters", func() {
			It("returns the list of parameters", func() {
				err := rmqc.DeleteAllRuntimeParameters()
				Ω(err).Should(BeNil())

				fDef := FederationDefinition{
					Uri: "amqp://server-name/%2f",
				}
				_, err = rmqc.PutFederationUpstream("rabbit/hole", "upstream1", fDef)
				Ω(err).Should(BeNil())

				sDef := ShovelDefinition{
					SourceURI:         "amqp://127.0.0.1/%2f",
					SourceQueue:       "mySourceQueue",
					DestinationURI:    "amqp://127.0.0.1/%2f",
					DestinationQueue:  "myDestQueue",
					AddForwardHeaders: true,
					AckMode:           "on-confirm",
					DeleteAfter:       "never",
				}

				_, err = rmqc.DeclareShovel("/", "shovel1", sDef)
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				list, err := rmqc.ListRuntimeParameters()
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(2))

				_, err = rmqc.DeleteFederationUpstream("rabbit/hole", "upstream1")
				Ω(err).Should(BeNil())

				_, err = rmqc.DeleteShovel("/", "shovel1")
				Ω(err).Should(BeNil())

				awaitEventPropagation()

				list, err = rmqc.ListRuntimeParameters()
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(0))

				// cleanup
				_, err = rmqc.DeleteQueue("/", "mySourceQueue")
				Ω(err).Should(BeNil())
				_, err = rmqc.DeleteQueue("/", "myDestQueue")
				Ω(err).Should(BeNil())
			})
		})
	})

	Context("paramToUpstream", func() {
		Context("when the parameter value is not initialized", func() {
			It("returns an empty FederationUpstream", func() {
				p := RuntimeParameter{} // p.Value is interface{}
				up := paramToUpstream(&p)
				Ω(up.Name).Should(BeEmpty())
				Ω(up.Vhost).Should(BeEmpty())
				Ω(up.Component).Should(BeEmpty())
				Ω(up.Definition).Should(Equal(FederationDefinition{}))
			})
		})
	})

	Context("feature flags", func() {
		It("lists and enables feature flags", func() {
			By("GET /feature-flags")
			featureFlags, err := rmqc.ListFeatureFlags()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(featureFlags).ShouldNot(BeNil())

			By("PUT /feature-flags/{name}/enable")
			for _, f := range featureFlags {
				resp, err := rmqc.EnableFeatureFlag(f.Name)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(resp).Should(HaveHTTPStatus(http.StatusNoContent))
			}

			featureFlags, err = rmqc.ListFeatureFlags()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(featureFlags).ShouldNot(BeNil())
			for _, f := range featureFlags {
				Ω(f.State).Should(Equal(StateEnabled))
			}
		})
	})

	Context("definition export", func() {
		It("returns exported definitions", func() {
			By("GET /definitions")
			defs, err := rmqc.ListDefinitions()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(defs).ShouldNot(BeNil())

			Ω(defs.RabbitMQVersion).ShouldNot(BeNil())

			Ω(defs.Vhosts).ShouldNot(BeEmpty())
			Ω(defs.Users).ShouldNot(BeEmpty())

			Ω(defs.Queues).ShouldNot(BeNil())
			Ω(defs.Parameters).ShouldNot(BeNil())
			Ω(defs.Policies).ShouldNot(BeNil())
		})

		It("returns exported global parameters", func() {
			By("GET /definitions")
			defs, err := rmqc.ListDefinitions()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(defs).ShouldNot(BeNil())

			foundClusterName := false
			for _, param := range defs.GlobalParameters {
				if param.Name == "cluster_name" {
					foundClusterName = true
				}
			}
			Ω(foundClusterName).Should(Equal(true))
		})
	})
})
