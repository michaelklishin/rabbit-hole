package rabbithole

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	amqp "github.com/rabbitmq/amqp091-go"
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

func FindExchangeByName(sl []ExchangeInfo, name string) (x ExchangeInfo) {
	for _, i := range sl {
		if name == i.Name {
			x = i
			break
		}
	}
	return x
}

func FindBindingBySourceAndDestinationNames(sl []BindingInfo, sourceName string, destinationName string) (b BindingInfo) {
	for _, i := range sl {
		if sourceName == i.Source && destinationName == i.Destination {
			b = i
			break
		}
	}
	return b
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
	return openConnectionWithCredentials(vhost, "guest", "guest")
}

func openConnectionWithCredentials(vhost string, username string, password string) *amqp.Connection {
	uri := fmt.Sprintf("amqp://%s:%s@localhost:5672/%s", username, password, url.QueryEscape(vhost))
	conn, err := amqp.Dial(uri)
	Ω(err).Should(BeNil())

	if err != nil {
		panic("failed to connect")
	}

	return conn
}

func ensureNonZeroMessageRate(ch *amqp.Channel) {
	for i := 0; i < 1000; i++ {
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

func shortSleep() {
	time.Sleep(time.Duration(600) * time.Millisecond)
}

func mediumSleep() {
	time.Sleep(time.Duration(1100) * time.Millisecond)
}

type portTestStruct struct {
	Port Port `json:"port"`
}

var _ = BeforeSuite(func() {
	val, _ := os.LookupEnv("GOMEGA_DEFAULT_EVENTUALLY_TIMEOUT")
	log.Printf("Using Gomega Eventually matcher timeout of %s", val)
})

var _ = Describe("RabbitMQ HTTP API client", func() {
	var (
		rmqc *Client
	)

	BeforeEach(func() {
		rmqc, _ = NewClient("http://127.0.0.1:15672", "guest", "guest")
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
			xs, _ := rmqc.ListConnections()
			for _, c := range xs {
				rmqc.CloseConnection(c.Name)
			}

			Eventually(func(g Gomega) []ConnectionInfo {
				xs, err := rmqc.ListConnections()
				Ω(err).Should(BeNil())

				return xs
			}).Should(BeEmpty())

			conn := openConnection("/")
			defer conn.Close()

			Eventually(func(g Gomega) []ConnectionInfo {
				xs, err := rmqc.ListConnections()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			xs, err := rmqc.ListConnections()

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

	// rabbitmq/rabbitmq-server#8482, rabbitmq/rabbitmq-server#5319
	Context("DELETE /api/connections/username/{username} invoked by an administrator", func() {
		It("closes the connection", func() {
			// first close all connections
			xs, _ := rmqc.ListConnections()
			for _, c := range xs {
				rmqc.CloseConnection(c.Name)
			}
			u := "policymaker"

			Eventually(func(g Gomega) []UserConnectionInfo {
				xs, err := rmqc.ListConnectionsOfUser(u)
				Ω(err).Should(BeNil())

				return xs
			}).Should(BeEmpty())

			conn := openConnectionWithCredentials("/", u, u)
			defer conn.Close()

			Eventually(func(g Gomega) []UserConnectionInfo {
				xs, err := rmqc.ListConnectionsOfUser(u)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			closeEvents := make(chan *amqp.Error)
			conn.NotifyClose(closeEvents)

			_, err := rmqc.CloseAllConnectionsOfUser(u)
			Ω(err).Should(BeNil())

			evt := <-closeEvents
			Ω(evt).ShouldNot(BeNil())
			Ω(evt.Code).Should(Equal(320))
			Ω(evt.Reason).Should(Equal("CONNECTION_FORCED - Closed via management plugin"))
			// server-initiated
			Ω(evt.Server).Should(Equal(true))
		})
	})

	// rabbitmq/rabbitmq-server#8482, rabbitmq/rabbitmq-server#5319
	Context("DELETE /api/connections/username/{username} invoked by a non-privileged user, case 1", func() {
		It("closes the connection", func() {
			Skip("unskip when rabbitmq/rabbitmq-server#8483 ships in a GA release")

			// first close all connections as an administrative user
			xs, _ := rmqc.ListConnections()
			for _, c := range xs {
				rmqc.CloseConnection(c.Name)
			}
			u := "policymaker"

			// an HTTP API client that uses policymaker-level permissions
			alt_rmqc, _ := NewClient("http://127.0.0.1:15672", u, u)

			Eventually(func(g Gomega) []ConnectionInfo {
				xs, err := alt_rmqc.ListConnections()
				Ω(err).Should(BeNil())

				return xs
			}).Should(BeEmpty())

			conn := openConnectionWithCredentials("/", u, u)
			defer conn.Close()

			// the user can list their own connections
			Eventually(func(g Gomega) []UserConnectionInfo {
				xs, err := alt_rmqc.ListConnectionsOfUser(u)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			closeEvents := make(chan *amqp.Error)
			conn.NotifyClose(closeEvents)

			// the user can close their own connections
			_, err := alt_rmqc.CloseAllConnectionsOfUser(u)
			Ω(err).Should(BeNil())

			evt := <-closeEvents
			Ω(evt).ShouldNot(BeNil())
			Ω(evt.Code).Should(Equal(320))
			Ω(evt.Reason).Should(Equal("CONNECTION_FORCED - Closed via management plugin"))
			// server-initiated
			Ω(evt.Server).Should(Equal(true))
		})
	})

	// rabbitmq/rabbitmq-server#8482, rabbitmq/rabbitmq-server#5319
	Context("DELETE /api/connections/username/{username} invoked by a non-privileged user, case 2", func() {
		It("fails with insufficient permissions", func() {
			Skip("unskip when rabbitmq/rabbitmq-server#8483 ships in a GA release")

			u := "policymaker"

			// an HTTP API client that uses policymaker-level permissions
			alt_rmqc, _ := NewClient("http://127.0.0.1:15672", u, u)

			conn := openConnection("/")
			defer conn.Close()

			// the user cannot list connections of the default administrative user
			_, err := alt_rmqc.ListConnectionsOfUser("guest")
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Error: API responded with a 401 Unauthorized"))

			// the user cannot close connections of the default administrative user
			_, err = alt_rmqc.CloseAllConnectionsOfUser("guest")
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Error: API responded with a 401 Unauthorized"))
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

			Eventually(func(g Gomega) string {
				cn2, err := rmqc.GetClusterName()
				Ω(err).Should(BeNil())
				return cn2.Name
			}).Should(Equal(cnStr))

			// Restore cluster name
			resp, err = rmqc.SetClusterName(*previousClusterName)
			Ω(resp.Status).Should(Equal("204 No Content"))
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /connections when there are active connections", func() {
		It("returns decoded response", FlakeAttempts(3), func() {
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

			Eventually(func(g Gomega) []ConnectionInfo {
				xs, err := rmqc.ListConnections()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			xs, err := rmqc.ListConnections()
			Ω(err).Should(BeNil())

			info := xs[0]
			Ω(info.Name).ShouldNot(BeNil())
			Ω(info.Host).ShouldNot(BeEmpty())
			Ω(info.UsesTLS).Should(Equal(false))

		})
	})

	Context("GET /connections paged with arguments", func() {
		It("returns decoded response", func() {
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

			params := url.Values{}
			params.Add("page", "1")

			Eventually(func(g Gomega) []ConnectionInfo {
				page, err := rmqc.PagedListConnectionsWithParameters(params)
				Ω(err).Should(BeNil())

				return page.Items
			}).ShouldNot(BeEmpty())

			xs, err := rmqc.PagedListConnectionsWithParameters(params)
			Ω(err).Should(BeNil())
			Ω(xs.Page).Should(Equal(1))
			Ω(xs.PageCount).Should(Equal(1))
			Ω(xs.ItemCount).ShouldNot(BeNil())
			Ω(xs.PageSize).Should(Equal(100))
			Ω(xs.TotalCount).ShouldNot(BeNil())
			Ω(xs.FilteredCount).ShouldNot(BeNil())

			info := xs.Items[0]
			Ω(info.Name).ShouldNot(BeNil())
			Ω(info.Host).ShouldNot(BeEmpty())
			Ω(info.UsesTLS).Should(Equal(false))
		})
	})

	Context("GET /connections/username/{username} when there are active connections", func() {
		It("returns decoded response", FlakeAttempts(3), func() {
			conn := openConnectionWithCredentials("/", "policymaker", "policymaker")
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

			Eventually(func(g Gomega) []UserConnectionInfo {
				xs, err := rmqc.ListConnectionsOfUser("guest")
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			Eventually(func(g Gomega) []UserConnectionInfo {
				xs, err := rmqc.ListConnectionsOfUser("policymaker")
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			xs, err := rmqc.ListConnectionsOfUser("guest")
			Ω(err).Should(BeNil())

			info := xs[0]
			Ω(info.Name).ShouldNot(BeNil())
			Ω(info.User).ShouldNot(BeEmpty())
			Ω(info.Vhost).Should(Equal("/"))

			// administrative users can list connections of any user
			_, err = rmqc.ListConnectionsOfUser("policymaker")
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /channels when there are active connections with open channels", func() {
		It("returns decoded response", FlakeAttempts(3), func() {
			cs, _ := rmqc.ListConnections()
			for _, c := range cs {
				rmqc.CloseConnection(c.Name)
			}

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

			Eventually(func(g Gomega) []ChannelInfo {
				xs, err := rmqc.ListChannels()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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
		It("returns decoded response", FlakeAttempts(3), func() {
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

			Eventually(func(g Gomega) []ConnectionInfo {
				xs, err := rmqc.ListConnections()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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
		It("returns decoded response", FlakeAttempts(3), func() {
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

			Eventually(func(g Gomega) []ChannelInfo {
				xs, err := rmqc.ListChannels()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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
			Ω(bool(x.AutoDelete)).Should(Equal(false))
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
			qs, _ := rmqc.ListQueues()
			for _, q := range qs {
				rmqc.DeleteQueue(q.Vhost, q.Name, QueueDeleteOptions{IfEmpty: false})
			}

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

			shortSleep()
			Eventually(func(g Gomega) []QueueInfo {
				xs, err := rmqc.ListQueues()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			qs2, err := rmqc.ListQueues()
			Ω(err).Should(BeNil())

			q := qs2[0]
			Ω(q.Name).ShouldNot(Equal(""))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Durable).ShouldNot(BeNil())

			for _, q = range qs2 {
				rmqc.DeleteQueue(q.Vhost, q.Name, QueueDeleteOptions{IfEmpty: false})
			}
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

			params := url.Values{}
			params.Add("lengths_age", "1800")
			params.Add("lengths_incr", "30")

			mediumSleep()
			Eventually(func(g Gomega) []QueueInfo {
				xs, err := rmqc.ListQueuesWithParameters(params)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			qs, err := rmqc.ListQueuesWithParameters(params)
			Ω(err).Should(BeNil())

			q := qs[0]
			Ω(q.Name).ShouldNot(Equal(""))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Durable).ShouldNot(BeNil())
			Ω(q.MessagesDetails.Samples[0]).ShouldNot(BeNil())
		})
	})

	Context("GET /queues/{vhost} with arguments", func() {
		It("returns decoded response", func() {
			vh := "rabbit/hole"
			conn := openConnection(vh)
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			qn := "rabbit-hole.queues/in.vhost.with.arguments"
			rmqc.DeleteQueue(vh, qn)
			_, err = ch.QueueDeclare(
				qn,    // name
				false, // durable
				false, // auto delete
				false, // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			params := url.Values{}
			params.Add("lengths_age", "1800")
			params.Add("lengths_incr", "30")

			mediumSleep()
			Eventually(func(g Gomega) []QueueInfo {
				xs, err := rmqc.ListQueuesWithParametersIn(vh, params)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			qs, err := rmqc.ListQueuesWithParametersIn(vh, params)
			Ω(err).Should(BeNil())

			q := FindQueueByName(qs, qn)
			Ω(q.Name).ShouldNot(Equal(""))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Vhost).Should(Equal(vh))
			Ω(q.Durable).ShouldNot(BeNil())
			Ω(q.MessagesDetails.Samples[0]).ShouldNot(BeNil())

			rmqc.DeleteQueue(vh, qn)
		})
	})

	Context("GET /queues paged with arguments", func() {
		It("returns decoded response", func() {
			vh := "/"
			conn := openConnection(vh)
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			qn := "rabbit-hole.queues/paged.with.arguments"
			_, err = ch.QueueDeclare(
				qn,    // name
				false, // durable
				false, // auto delete
				false, // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			params := url.Values{}
			params.Add("page", "1")

			mediumSleep()
			Eventually(func(g Gomega) []QueueInfo {
				page, err := rmqc.PagedListQueuesWithParameters(params)
				Ω(err).Should(BeNil())

				return page.Items
			}).ShouldNot(BeEmpty())

			qs, err := rmqc.PagedListQueuesWithParameters(params)
			Ω(err).Should(BeNil())

			q := FindQueueByName(qs.Items, qn)

			Ω(q.Name).Should(Equal(qn))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Durable).ShouldNot(BeNil())
			Ω(qs.Page).Should(Equal(1))
			Ω(qs.PageCount).Should(Equal(1))
			Ω(qs.ItemCount).ShouldNot(BeNil())
			Ω(qs.PageSize).Should(Equal(100))
			Ω(qs.TotalCount).ShouldNot(BeNil())
			Ω(qs.FilteredCount).ShouldNot(BeNil())

			_, err = rmqc.DeleteQueue(vh, qn, QueueDeleteOptions{IfEmpty: true})
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /queues/{vhost} paged with arguments", func() {
		It("returns decoded response", func() {
			vh := "rabbit/hole"
			conn := openConnection(vh)
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			qn := "rabbit-hole.queues./.paged"
			_, err = ch.QueueDeclare(
				qn,    // name
				false, // durable
				false, // auto delete
				false, // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			params := url.Values{}
			params.Add("page", "1")

			shortSleep()
			Eventually(func(g Gomega) []QueueInfo {
				page, err := rmqc.PagedListQueuesWithParametersIn(vh, params)
				Ω(err).Should(BeNil())

				return page.Items
			}).ShouldNot(BeEmpty())

			qs, err := rmqc.PagedListQueuesWithParametersIn(vh, params)
			Ω(err).Should(BeNil())

			q := FindQueueByName(qs.Items, qn)

			Ω(q.Name).Should(Equal(qn))
			Ω(q.Node).ShouldNot(BeNil())
			Ω(q.Durable).ShouldNot(BeNil())
			Ω(qs.Page).Should(Equal(1))
			Ω(qs.PageCount).Should(Equal(1))
			Ω(qs.ItemCount).ShouldNot(BeNil())
			Ω(qs.PageSize).Should(Equal(100))
			Ω(qs.TotalCount).ShouldNot(BeNil())
			Ω(qs.FilteredCount).ShouldNot(BeNil())
			Ω(q.Vhost).Should(Equal(vh))

			_, err = rmqc.DeleteQueue(vh, qn, QueueDeleteOptions{IfEmpty: true})
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /queues/{vhost}", func() {
		It("returns decoded response", func() {
			vh := "rabbit/hole"
			conn := openConnection(vh)
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			qn := "rabbit-hole.queues.in.a.vhost"
			_, err = ch.QueueDeclare(
				qn,    // name
				false, // durable
				false, // auto delete
				false, // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			mediumSleep()
			Eventually(func(g Gomega) []QueueInfo {
				qs, err := rmqc.ListQueuesIn(vh)
				Ω(err).Should(BeNil())

				return qs
			}).ShouldNot(BeEmpty())

			qs, err := rmqc.ListQueuesIn(vh)
			Ω(err).Should(BeNil())

			q := FindQueueByName(qs, qn)
			Ω(q.Name).Should(Equal(qn))
			Ω(q.Vhost).Should(Equal(vh))
			Ω(q.Durable).Should(Equal(false))

			_, err = rmqc.DeleteQueue(vh, qn, QueueDeleteOptions{IfEmpty: true})
			Ω(err).Should(BeNil())
		})
	})

	Context("GET /queues/{vhost}/{name}", func() {
		It("returns decoded response", func() {
			vh := "rabbit/hole"
			conn := openConnection(vh)
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			qn := "rabbit-hole.queues.in.a.vhost/named"
			_, err = ch.QueueDeclare(
				qn,    // name
				false, // durable
				false, // auto delete
				false, // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			shortSleep()
			Eventually(func(g Gomega) string {
				q, err := rmqc.GetQueue(vh, qn)
				Ω(err).Should(BeNil())

				return q.Vhost
			}).Should(Equal(vh))

			q, err := rmqc.GetQueue(vh, qn)
			Ω(err).Should(BeNil())

			Ω(q.Vhost).Should(Equal(vh))
			Ω(q.Durable).Should(Equal(false))

			_, err = rmqc.DeleteQueue(vh, qn, QueueDeleteOptions{IfEmpty: true})
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /queues/{vhost}/{name}", func() {
		It("deletes a queue", func() {
			vh := "rabbit/hole"
			conn := openConnection(vh)
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			qn := "rabbit-hole.queues.in.a.vhost/to.be.deleted"
			q, err := ch.QueueDeclare(
				qn,    // name
				false, // durable
				false, // auto delete
				false, // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			shortSleep()
			Eventually(func(g Gomega) string {
				q, err := rmqc.GetQueue(vh, qn)
				Ω(err).Should(BeNil())

				return q.Name
			}).Should(Equal(qn))

			_, err = rmqc.GetQueue(vh, q.Name)
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteQueue(vh, q.Name)
			Ω(err).Should(BeNil())

			qi2, err := rmqc.GetQueue(vh, q.Name)
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

			mediumSleep()
			Eventually(func(g Gomega) []ConsumerInfo {
				xs, err := rmqc.ListConsumers()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			mediumSleep()
			Eventually(func(g Gomega) []ConsumerInfo {
				xs, err := rmqc.ListConsumers()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			shortSleep()
			Eventually(func(g Gomega) string {
				u, err := rmqc.GetUser(username)
				Ω(err).Should(BeNil())

				return u.Name
			}).Should(Equal(username))

			u, err := rmqc.GetUser(username)
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

			shortSleep()
			Eventually(func(g Gomega) string {
				u, err := rmqc.GetUser(username)
				Ω(err).Should(BeNil())

				return u.Name
			}).Should(Equal(username))

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
			username := "rabbithole-passwordless"
			rmqc.DeleteUser(username)
			resp, err := rmqc.PutUserWithoutPassword(username, info)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			shortSleep()
			Eventually(func(g Gomega) string {
				u, err := rmqc.GetUser(username)
				Ω(err).Should(BeNil())

				return u.Name
			}).Should(Equal(username))

			u, err := rmqc.GetUser(username)
			Ω(err).Should(BeNil())

			Ω(u.PasswordHash).Should(BeEquivalentTo(""))
			Ω(u.Tags).Should(Equal(tags))

			// cleanup
			_, err = rmqc.DeleteUser(username)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /users/{name}", func() {
		It("deletes the user", func() {
			username := "rabbithole"
			info := UserSettings{Password: "s3krE7", Tags: UserTags{"management", "policymaker"}}
			_, err := rmqc.PutUser("rabbithole", info)
			Ω(err).Should(BeNil())

			u, err := rmqc.GetUser(username)
			Ω(err).Should(BeNil())
			Ω(u).ShouldNot(BeNil())

			resp, err := rmqc.DeleteUser(username)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			shortSleep()
			Eventually(func(g Gomega) *UserInfo {
				u, _ := rmqc.GetUser(username)

				return u
			}).Should(BeNil())

			u2, err := rmqc.GetUser(username)
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
			Ω(x.Description).ShouldNot(BeNil())
			Ω(x.Tags).ShouldNot(BeNil())
			Ω(x.Tracing).ShouldNot(BeNil())
			Ω(x.ClusterState).ShouldNot(BeNil())
		})
	})

	Context("GET /vhosts/{name} when vhost exists", func() {
		It("returns decoded response", func() {
			x, err := rmqc.GetVhost("rabbit/hole")
			Ω(err).Should(BeNil())

			Ω(x.Name).ShouldNot(BeNil())
			Ω(x.Description).ShouldNot(BeNil())
			Ω(x.Tags).ShouldNot(BeNil())
			Ω(x.Tracing).ShouldNot(BeNil())
			Ω(x.ClusterState).ShouldNot(BeNil())
		})
	})

	Context("PUT /vhosts/{name}", func() {
		It("creates a vhost", func() {
			vh := "rabbit/hole2"
			tags := VhostTags{"production", "eu-west-1"}
			vs := VhostSettings{Description: "rabbit/hole2 vhost", Tags: tags, Tracing: false}
			_, err := rmqc.PutVhost(vh, vs)
			Ω(err).Should(BeNil())

			shortSleep()
			Eventually(func(g Gomega) string {
				v, _ := rmqc.GetVhost(vh)
				Ω(err).Should(BeNil())

				return v.Name
			}).Should(Equal(vh))

			x, err := rmqc.GetVhost(vh)
			Ω(err).Should(BeNil())

			Ω(x.Name).Should(BeEquivalentTo(vh))
			Ω(x.Description).Should(BeEquivalentTo("rabbit/hole2 vhost"))
			Ω(x.Tags).Should(Equal(tags))
			Ω(x.Tracing).Should(Equal(false))

			_, err = rmqc.DeleteVhost(vh)
			Ω(err).Should(BeNil())
		})
		When("A default queue type is set", func() {
			It("creates a vhost with a default queue type", func() {
				vh := "rabbit/hole3"
				tags := VhostTags{"production", "eu-west-1"}
				vs := VhostSettings{Description: "rabbit/hole3 vhost", DefaultQueueType: "quorum", Tags: tags, Tracing: false}
				_, err := rmqc.PutVhost(vh, vs)
				Ω(err).Should(BeNil())

				shortSleep()
				Eventually(func(g Gomega) string {
					v, err := rmqc.GetVhost(vh)
					Ω(err).Should(BeNil())

					return v.Name
				}).Should(Equal(vh))

				x, err := rmqc.GetVhost(vh)
				Ω(err).Should(BeNil())

				Ω(x.Name).Should(BeEquivalentTo(vh))
				Ω(x.Description).Should(BeEquivalentTo("rabbit/hole3 vhost"))
				Ω(x.DefaultQueueType).Should(BeEquivalentTo("quorum"))
				Ω(x.Tags).Should(Equal(tags))
				Ω(x.Tracing).Should(Equal(false))

				_, err = rmqc.DeleteVhost(vh)
				Ω(err).Should(BeNil())
			})
		})
	})

	Context("DELETE /vhosts/{name}", func() {
		It("deletes a vhost", func() {
			vh := "rabbit/hole2"
			vs := VhostSettings{Description: "rabbit/hole2 vhost", Tags: VhostTags{"production", "eu-west-1"}, Tracing: false}
			_, err := rmqc.PutVhost(vh, vs)
			Ω(err).Should(BeNil())

			shortSleep()
			Eventually(func(g Gomega) string {
				v, _ := rmqc.GetVhost(vh)
				Ω(err).Should(BeNil())

				return v.Name
			}).Should(Equal(vh))

			x, err := rmqc.GetVhost(vh)
			Ω(err).Should(BeNil())
			Ω(x).ShouldNot(BeNil())

			_, err = rmqc.DeleteVhost(vh)
			Ω(err).Should(BeNil())

			x2, err := rmqc.GetVhost(vh)
			Ω(x2).Should(BeNil())
			Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("GET /vhosts/{name}/connections", func() {
		It("returns decoded response", func() {
			vh := "rabbit/hole"

			conn := openConnection(vh)
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

			shortSleep()
			Eventually(func(g Gomega) []ConnectionInfo {
				xs, _ := rmqc.ListVhostConnections(vh)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			xs, err := rmqc.ListVhostConnections(vh)
			Ω(err).Should(BeNil())

			c1 := xs[0]
			Ω(c1.Protocol).Should(Equal("AMQP 0-9-1"))
			Ω(c1.User).Should(Equal("guest"))
		})
	})

	Context("vhost-limits", func() {
		maxConnections := 10
		maxQueues := 2

		It("returns an empty list of limits", func() {
			vh := "rabbit/hole"
			_, err := rmqc.DeleteVhostLimits(vh, VhostLimits{"max-connections", "max-queues"})
			Ω(err).Should(BeNil())

			xs, err2 := rmqc.GetVhostLimits(vh)
			Ω(err2).Should(BeNil())
			Ω(xs).Should(HaveLen(0))

			rmqc.DeleteVhostLimits(vh, VhostLimits{"max-connections", "max-queues"})
		})

		It("sets the limits", FlakeAttempts(3), func() {
			vh := "rabbit/hole"
			_, err := rmqc.DeleteVhostLimits(vh, VhostLimits{"max-connections", "max-queues"})
			Ω(err).Should(BeNil())

			_, err2 := rmqc.PutVhostLimits(vh, VhostLimitsValues{
				"max-connections": maxConnections,
				"max-queues":      maxQueues,
			})
			Ω(err2).Should(BeNil())

			Eventually(func(g Gomega) []VhostLimitsInfo {
				xs, _ := rmqc.GetAllVhostLimits()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			xs, err3 := rmqc.GetAllVhostLimits()
			Ω(err3).Should(BeNil())
			Ω(xs).Should(HaveLen(1))
			Ω(xs[0].Vhost).Should(Equal(vh))
			Ω(xs[0].Value["max-connections"]).Should(Equal(maxConnections))
			Ω(xs[0].Value["max-queues"]).Should(Equal(maxQueues))

			rmqc.DeleteVhostLimits(vh, VhostLimits{"max-connections", "max-queues"})
		})
	})

	Context("user-limits", func() {
		maxConnections := 1
		maxChannels := 2

		It("returns an empty list of limits", func() {
			u := "guest"
			_, err := rmqc.DeleteUserLimits(u, UserLimits{"max-connections", "max-channels"})

			xs, err := rmqc.GetUserLimits(u)
			Ω(err).Should(BeNil())
			Ω(xs).Should(HaveLen(0))
		})

		It("sets the limits", func() {
			u := "guest"
			rmqc.DeleteUserLimits(u, UserLimits{"max-connections", "max-channels"})

			_, err := rmqc.PutUserLimits(u, UserLimitsValues{
				"max-connections": maxConnections,
				"max-channels":    maxChannels,
			})
			Ω(err).Should(BeNil())

			xs, err2 := rmqc.GetUserLimits(u)
			Ω(err2).Should(BeNil())
			Ω(xs).Should(HaveLen(1))
			Ω(xs[0].User).Should(Equal(u))
			Ω(xs[0].Value["max-connections"]).Should(Equal(maxConnections))
			Ω(xs[0].Value["max-channels"]).Should(Equal(maxChannels))

			xs3, err3 := rmqc.GetAllUserLimits()
			Ω(err3).Should(BeNil())
			Ω(xs3).Should(HaveLen(1))
			Ω(xs3[0].User).Should(Equal("guest"))
			Ω(xs3[0].Value["max-connections"]).Should(Equal(maxConnections))
			Ω(xs3[0].Value["max-channels"]).Should(Equal(maxChannels))

			rmqc.DeleteUserLimits(u, UserLimits{"max-connections", "max-channels"})
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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListBindings()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListBindingsIn("/")
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListQueueBindings("/", q.Name)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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
			qn := "rabbit-hole.test.bindings.post.queue"
			rmqc.DeleteQueue(vh, qn)

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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListQueueBindings(vh, qn)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			rmqc.DeleteQueue(vh, qn)
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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListQueueBindingsBetween(vh, x, q)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListExchangeBindingsBetween(vh, sx, dx)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListExchangeBindingsWithSource(vh, sx)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListExchangeBindingsWithDestination(vh, dx)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListExchangeBindingsBetween(vh, "amq.topic", xn)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListExchangeBindingsWithDestination(vh, xn)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			// Grab the Location data from the POST response {destination}/{propertiesKey}
			propertiesKey, _ := url.QueryUnescape(strings.Split(res.Header.Get("Location"), "/")[1])
			info.PropertiesKey = propertiesKey

			_, err = rmqc.DeleteBinding(vh, info)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) []BindingInfo {
				xs, _ := rmqc.ListExchangeBindingsWithDestination(vh, xn)
				Ω(err).Should(BeNil())

				return xs
			}).Should(BeEmpty())

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

			Eventually(func(g Gomega) string {
				p, _ := rmqc.GetPermissionsIn("/", u)
				Ω(err).Should(BeNil())

				return p.User
			}).Should(Equal(u))

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

			Eventually(func(g Gomega) string {
				p, _ := rmqc.GetPermissionsIn("/", u)
				Ω(err).Should(BeNil())

				return p.User
			}).Should(Equal(u))

			_, err = rmqc.ClearPermissionsIn("/", u)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetPermissionsIn("/", u)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))

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

			Eventually(func(g Gomega) []TopicPermissionInfo {
				xs, _ := rmqc.ListTopicPermissions()
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []TopicPermissionInfo {
				xs, _ := rmqc.ListTopicPermissionsOf(u)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			xs, err := rmqc.ListTopicPermissionsOf(u)
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

			Eventually(func(g Gomega) []TopicPermissionInfo {
				xs, _ := rmqc.GetTopicPermissionsIn("/", u)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

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

			Eventually(func(g Gomega) []TopicPermissionInfo {
				xs, _ := rmqc.ListTopicPermissionsOf(u)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			_, err = rmqc.ClearTopicPermissionsIn("/", u)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetTopicPermissionsIn("/", u)

				return err
			}).ShouldNot(BeNil())

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

			permissions := TopicPermissions{Exchange: "amq.topic", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdateTopicPermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())
			permissions = TopicPermissions{Exchange: "foobar", Write: "log.*", Read: "log.*"}
			_, err = rmqc.UpdateTopicPermissionsIn("/", u, permissions)
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteTopicPermissionsIn("/", u, "foobar")
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) []TopicPermissionInfo {
				xs, err := rmqc.ListTopicPermissionsOf(u)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			xs, err := rmqc.ListTopicPermissionsOf(u)
			Ω(err).Should(BeNil())

			Ω(len(xs)).Should(Equal(1))

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

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetExchange(vh, xn)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(xn))

			x, err := rmqc.GetExchange(vh, xn)
			Ω(err).Should(BeNil())
			Ω(x.Name).Should(Equal(xn))
			Ω(x.Durable).Should(Equal(false))
			Ω(bool(x.AutoDelete)).Should(Equal(false))
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

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetExchange(vh, xn)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("PUT /queues/{vhost}/{queue}", func() {
		It("declares a queue", func() {
			vh := "rabbit/hole"
			qn := "rabbit-hole.temporary-declare"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: true, Type: "quorum"})
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) string {
				q, err := rmqc.GetQueue(vh, qn)
				Ω(err).Should(BeNil())

				return q.Name
			}).Should(Equal(qn))

			q, err := rmqc.GetQueue(vh, qn)
			Ω(err).Should(BeNil())
			Ω(q.Name).Should(Equal(qn))
			Ω(q.Durable).Should(Equal(true))
			Ω(bool(q.AutoDelete)).Should(Equal(false))
			Ω(q.Vhost).Should(Equal(vh))
			Ω(q.Arguments).To(HaveKeyWithValue("x-queue-type", "quorum"))

			_, err = rmqc.DeleteQueue(vh, qn)
			Ω(err).Should(BeNil())
		})
	})

	Context("DELETE /queues/{vhost}/{queue}", func() {
		It("deletes a queue", func() {
			vh := "rabbit/hole"
			qn := "rabbit-hole.temporary/to.be.deleted"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) string {
				q, err := rmqc.GetQueue(vh, qn)
				Ω(err).Should(BeNil())

				return q.Name
			}).Should(Equal(qn))

			_, err = rmqc.DeleteQueue(vh, qn)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetQueue(vh, qn)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})

		It("accepts IfEmpty option", func() {
			vh := "rabbit/hole"
			qn := "rabbit-hole.temporary/if.empty"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) string {
				q, err := rmqc.GetQueue(vh, qn)
				Ω(err).Should(BeNil())

				return q.Name
			}).Should(Equal(qn))

			_, err = rmqc.DeleteQueue(vh, qn, QueueDeleteOptions{IfEmpty: true})
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetQueue(vh, qn)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})

		It("accepts IfUnused option", func() {
			vh := "rabbit/hole"
			qn := "rabbit-hole.temporary/if.unused"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) string {
				q, err := rmqc.GetQueue(vh, qn)
				Ω(err).Should(BeNil())

				return q.Name
			}).Should(Equal(qn))

			_, err = rmqc.DeleteQueue(vh, qn, QueueDeleteOptions{IfUnused: true})
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetQueue(vh, qn)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("DELETE /queues/{vhost}/{queue}/contents", func() {
		It("purges a queue", func() {
			vh := "rabbit/hole"
			qn := "rabbit-hole.temporary/contents"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) string {
				q, err := rmqc.GetQueue(vh, qn)
				Ω(err).Should(BeNil())

				return q.Name
			}).Should(Equal(qn))

			_, err = rmqc.PurgeQueue(vh, qn)
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteQueue(vh, qn)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetQueue(vh, qn)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("GET /policies", func() {
		Context("when policy exists", func() {
			It("returns decoded response", func() {
				rmqc.DeleteAllPolicies()

				policy1 := Policy{
					Pattern:    "abc",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100},
					Priority:   0,
				}

				policy2 := Policy{
					Pattern:    ".*",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100},
					Priority:   0,
				}

				// prepare policies
				_, err := rmqc.PutPolicy("rabbit/hole", "woot1", policy1)
				Ω(err).Should(BeNil())

				_, err = rmqc.PutPolicy("rabbit/hole", "woot2", policy2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []Policy {
					xs, err := rmqc.ListPolicies()
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

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

				Eventually(func(g Gomega) []Policy {
					xs, _ := rmqc.ListPolicies()

					return xs
				}).Should(BeEmpty())
			})
		})
	})

	Context("GET /polices/{vhost}", func() {
		Context("when policy exists", func() {
			It("returns decoded response", func() {
				rmqc.DeleteAllPolicies()

				policy1 := Policy{
					Pattern:    "abc",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100},
					Priority:   0,
				}

				policy2 := Policy{
					Pattern:    ".*",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100, "message-ttl": 100000},
					Priority:   0,
				}

				vh := "rabbit/hole"

				// prepare policies
				_, err := rmqc.PutPolicy(vh, "woot1", policy1)
				Ω(err).Should(BeNil())

				_, err = rmqc.PutPolicy("/", "woot2", policy2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []Policy {
					xs, err := rmqc.ListPolicies()
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

				// test
				pols, err := rmqc.ListPoliciesIn(vh)
				Ω(err).Should(BeNil())
				Ω(pols).ShouldNot(BeEmpty())
				Ω(len(pols)).Should(Equal(1))
				Ω(pols[0].Name).Should(Equal("woot1"))

				// cleanup
				_, err = rmqc.DeletePolicy(vh, "woot1")
				Ω(err).Should(BeNil())

				_, err = rmqc.DeletePolicy("/", "woot2")
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []Policy {
					xs, _ := rmqc.ListPolicies()

					return xs
				}).Should(BeEmpty())
			})
		})

		Context("when no policies exist", func() {
			It("returns decoded response", func() {
				rmqc.DeleteAllPolicies()

				vh := "rabbit/hole"
				Eventually(func(g Gomega) []Policy {
					xs, _ := rmqc.ListPoliciesIn(vh)

					return xs
				}).Should(BeEmpty())
			})
		})
	})

	Context("GET /policies/{vhost}/{name}", func() {
		Context("when policy exists", func() {
			It("returns decoded response", func() {
				rmqc.DeleteAllPolicies()

				policy := Policy{
					Pattern:    ".*",
					ApplyTo:    "all",
					Definition: PolicyDefinition{"expires": 100},
					Priority:   0,
				}

				vh := "rabbit/hole"
				name := "woot"

				_, err := rmqc.PutPolicy(vh, name, policy)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetPolicy(vh, name)
					Ω(err).Should(BeNil())

					return x.Name
				}).ShouldNot(BeEmpty())

				pol, err := rmqc.GetPolicy(vh, name)
				Ω(err).Should(BeNil())
				Ω(pol.Vhost).Should(Equal(vh))
				Ω(pol.Name).Should(Equal(name))
				Ω(pol.ApplyTo).Should(Equal("all"))
				Ω(pol.Pattern).Should(Equal(".*"))
				Ω(pol.Priority).Should(BeEquivalentTo(0))
				Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
				Ω(pol.Definition["expires"]).Should(BeEquivalentTo(100))

				_, err = rmqc.DeletePolicy(vh, name)
				Ω(err).Should(BeNil())
			})
		})

		Context("when policy not found", func() {
			It("returns decoded response", func() {
				vh := "rabbit/hole"
				name := "woot"

				Eventually(func(g Gomega) error {
					_, err := rmqc.GetPolicy(vh, name)

					return err
				}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
			})
		})
	})

	Context("DELETE /polices/{vhost}/{name}", func() {
		It("deletes the policy", func() {
			rmqc.DeleteAllPolicies()

			policy := Policy{
				Pattern:    ".*",
				ApplyTo:    "queues",
				Definition: PolicyDefinition{"expires": 100, "message-ttl": 100000},
				Priority:   0,
			}

			vh := "rabbit/hole"
			name := "woot-delete-me"

			_, err := rmqc.PutPolicy(vh, name, policy)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetPolicy(vh, name)
				Ω(err).Should(BeNil())

				return x.Name
			}).ShouldNot(BeEmpty())

			resp, err := rmqc.DeletePolicy(vh, name)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetPolicy(vh, name)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("PUT /policies/{vhost}/{name}", func() {
		It("creates the policy", func() {
			rmqc.DeleteAllPolicies()

			policy := Policy{
				Pattern:    ".*",
				ApplyTo:    "all",
				Definition: PolicyDefinition{"expires": 100},
				Priority:   0,
			}

			vh := "rabbit/hole"
			name := "woot"

			resp, err := rmqc.PutPolicy(vh, name, policy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetPolicy(vh, name)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(name))

			_, err = rmqc.DeletePolicy(vh, name)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetPolicy(vh, name)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})

		It("updates the policy", func() {
			rmqc.DeleteAllPolicies()

			policy := Policy{
				Pattern:    ".*",
				ApplyTo:    "all",
				Definition: PolicyDefinition{"expires": 100},
			}

			vh := "rabbit/hole"
			name := "woot"

			// create policy
			resp, err := rmqc.PutPolicy(vh, name, policy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetPolicy(vh, name)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(name))

			pol, err := rmqc.GetPolicy(vh, name)
			Ω(err).Should(BeNil())
			Ω(pol.Vhost).Should(Equal(vh))
			Ω(pol.Name).Should(Equal(name))
			Ω(pol.Pattern).Should(Equal(".*"))
			Ω(pol.ApplyTo).Should(Equal("all"))
			Ω(pol.Priority).Should(BeEquivalentTo(0))
			Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
			Ω(pol.Definition["expires"]).Should(BeEquivalentTo(100))

			// update the policy
			newPolicy := Policy{
				Pattern:    "\\d+",
				ApplyTo:    "queues",
				Definition: PolicyDefinition{"max-length": 100},
				Priority:   1,
			}

			resp, err = rmqc.PutPolicy(vh, name, newPolicy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetPolicy(vh, name)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(name))

			pol, err = rmqc.GetPolicy(vh, name)
			Ω(err).Should(BeNil())
			Ω(pol.Vhost).Should(Equal(vh))
			Ω(pol.Name).Should(Equal("woot"))
			Ω(pol.Pattern).Should(Equal("\\d+"))
			Ω(pol.ApplyTo).Should(Equal("queues"))
			Ω(pol.Priority).Should(BeEquivalentTo(1))
			Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
			Ω(pol.Definition["max-length"]).Should(BeEquivalentTo(100))
			Ω(pol.Definition["expires"]).Should(BeNil())

			// cleanup
			_, err = rmqc.DeletePolicy(vh, name)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetPolicy(vh, name)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("GET /operator-policies", func() {
		Context("when operator policy exists", func() {
			It("returns decoded response", func() {
				policy1 := OperatorPolicy{
					Pattern:    "abc",
					ApplyTo:    "queues",
					Definition: PolicyDefinition{"expires": 100, "delivery-limit": 202},
					Priority:   0,
				}

				policy2 := OperatorPolicy{
					Pattern:    ".*",
					ApplyTo:    "queues",
					Definition: PolicyDefinition{"expires": 100, "delivery-limit": 202},
					Priority:   0,
				}

				vh := "rabbit/hole"
				name1 := "woot1"
				name2 := "woot2"

				// prepare policies
				_, err := rmqc.PutOperatorPolicy(vh, name1, policy1)
				Ω(err).Should(BeNil())

				_, err = rmqc.PutOperatorPolicy(vh, name2, policy2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []OperatorPolicy {
					xs, err := rmqc.ListOperatorPolicies()
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

				// test
				pols, err := rmqc.ListOperatorPolicies()
				Ω(err).Should(BeNil())
				Ω(pols).ShouldNot(BeEmpty())
				Ω(len(pols)).Should(BeNumerically(">=", 2))
				Ω(pols[0].Name).ShouldNot(BeNil())
				Ω(pols[1].Name).ShouldNot(BeNil())

				// cleanup
				_, err = rmqc.DeleteOperatorPolicy(vh, name1)
				Ω(err).Should(BeNil())

				_, err = rmqc.DeleteOperatorPolicy(vh, name2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []OperatorPolicy {
					xs, _ := rmqc.ListOperatorPolicies()

					return xs
				}).Should(BeEmpty())
			})
		})
	})

	Context("GET /operator-policies/{vhost}", func() {
		Context("when operator policy exists", func() {
			It("returns decoded response", func() {
				policy1 := OperatorPolicy{
					Pattern:    "abc",
					ApplyTo:    "queues",
					Definition: PolicyDefinition{"expires": 100, "delivery-limit": 202},
					Priority:   0,
				}

				policy2 := OperatorPolicy{
					Pattern:    ".*",
					ApplyTo:    "queues",
					Definition: PolicyDefinition{"expires": 100, "delivery-limit": 202},
					Priority:   0,
				}

				vh := "rabbit/hole"
				name1 := "woot1"
				name2 := "woot2"

				// prepare policies
				_, err := rmqc.PutOperatorPolicy(vh, name1, policy1)
				Ω(err).Should(BeNil())

				_, err = rmqc.PutOperatorPolicy("/", name2, policy2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []OperatorPolicy {
					xs, err := rmqc.ListOperatorPoliciesIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

				// test
				pols, err := rmqc.ListOperatorPoliciesIn(vh)
				Ω(err).Should(BeNil())
				Ω(pols).ShouldNot(BeEmpty())
				Ω(len(pols)).Should(Equal(1))
				Ω(pols[0].Name).Should(Equal(name1))

				// cleanup
				_, err = rmqc.DeleteOperatorPolicy(vh, name1)
				Ω(err).Should(BeNil())

				_, err = rmqc.DeleteOperatorPolicy("/", name2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []OperatorPolicy {
					xs, _ := rmqc.ListOperatorPolicies()

					return xs
				}).Should(BeEmpty())
			})
		})

		Context("when no operator policies exist", func() {
			It("returns decoded response", func() {
				vh := "rabbit/hole"

				Eventually(func(g Gomega) []OperatorPolicy {
					xs, err := rmqc.ListOperatorPoliciesIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
			})
		})
	})

	Context("GET /operator-policies/{vhost}/{name}", func() {
		Context("when operator policy exists", func() {
			It("returns decoded response", func() {
				policy := OperatorPolicy{
					Pattern:    ".*",
					ApplyTo:    "queues",
					Definition: PolicyDefinition{"expires": 100, "delivery-limit": 202},
					Priority:   0,
				}

				vh := "rabbit/hole"
				name := "woot"

				_, err := rmqc.PutOperatorPolicy(vh, name, policy)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetOperatorPolicy(vh, name)
					Ω(err).Should(BeNil())

					return x.Name
				}).Should(Equal(name))

				pol, err := rmqc.GetOperatorPolicy(vh, name)
				Ω(err).Should(BeNil())
				Ω(pol.Vhost).Should(Equal(vh))
				Ω(pol.Name).Should(Equal(name))
				Ω(pol.ApplyTo).Should(Equal("queues"))
				Ω(pol.Pattern).Should(Equal(".*"))
				Ω(pol.Priority).Should(BeEquivalentTo(0))
				Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
				Ω(pol.Definition["expires"]).Should(BeEquivalentTo(100))
				Ω(pol.Definition["delivery-limit"]).Should(Equal(float64(202)))

				_, err = rmqc.DeleteOperatorPolicy(vh, name)
				Ω(err).Should(BeNil())
			})
		})

		Context("when operator policy not found", func() {
			It("returns decoded response", func() {
				vh := "rabbit/hole"
				name := "woot"

				rmqc.DeleteOperatorPolicy(vh, name)

				Eventually(func(g Gomega) error {
					_, err := rmqc.GetOperatorPolicy(vh, name)

					return err
				}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
			})
		})
	})

	Context("DELETE /operator-policies/{vhost}/{name}", func() {
		It("deletes the operator policy", func() {
			policy := OperatorPolicy{
				Pattern:    ".*",
				ApplyTo:    "queues",
				Definition: PolicyDefinition{"expires": 100, "delivery-limit": 202},
				Priority:   0,
			}

			vh := "rabbit/hole"
			name := "woot"

			_, err := rmqc.PutOperatorPolicy(vh, name, policy)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetOperatorPolicy(vh, name)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(name))

			resp, err := rmqc.DeleteOperatorPolicy(vh, name)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetOperatorPolicy(vh, name)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("PUT /operator-policies/{vhost}/{name}", func() {
		It("creates the operator policy", func() {
			policy := OperatorPolicy{
				Pattern: ".*",
				ApplyTo: "all",
				Definition: PolicyDefinition{
					"expires":          100,
					"delivery-limit":   202,
					"max-length-bytes": 100000,
				},
				Priority: 0,
			}

			vh := "rabbit/hole"
			name := "woot"

			resp, err := rmqc.PutOperatorPolicy(vh, name, policy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetOperatorPolicy(vh, name)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(name))

			_, err = rmqc.DeleteOperatorPolicy(vh, name)
			Ω(err).Should(BeNil())
		})

		It("updates the operator policy", func() {
			policy := OperatorPolicy{
				Pattern:    ".*",
				ApplyTo:    "queues",
				Definition: PolicyDefinition{"expires": 100, "delivery-limit": 202},
			}

			vh := "rabbit/hole"
			name := "woot"

			// create a policy
			resp, err := rmqc.PutOperatorPolicy(vh, name, policy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetOperatorPolicy(vh, name)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(name))

			pol, err := rmqc.GetOperatorPolicy(vh, name)
			Ω(err).Should(BeNil())
			Ω(pol.Vhost).Should(Equal(vh))
			Ω(pol.Name).Should(Equal(name))
			Ω(pol.Pattern).Should(Equal(".*"))
			Ω(pol.ApplyTo).Should(Equal("queues"))
			Ω(pol.Priority).Should(BeEquivalentTo(0))
			Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
			Ω(pol.Definition["delivery-limit"]).Should(Equal(float64(202)))
			Ω(pol.Definition["expires"]).Should(BeEquivalentTo(100))

			// update the policy
			newPolicy := OperatorPolicy{
				Pattern: "\\d+",
				ApplyTo: "queues",
				Definition: PolicyDefinition{
					"max-length":       100,
					"max-length-bytes": 100000,
					"message-ttl":      606,
				},
				Priority: 1,
			}

			resp, err = rmqc.PutOperatorPolicy(vh, name, newPolicy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(HavePrefix("20"))

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetOperatorPolicy(vh, name)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(name))

			pol, err = rmqc.GetOperatorPolicy(vh, name)
			Ω(err).Should(BeNil())
			Ω(pol.Vhost).Should(Equal(vh))
			Ω(pol.Name).Should(Equal(name))
			Ω(pol.Pattern).Should(Equal("\\d+"))
			Ω(pol.ApplyTo).Should(Equal("queues"))
			Ω(pol.Priority).Should(BeEquivalentTo(1))
			Ω(pol.Definition).Should(BeAssignableToTypeOf(PolicyDefinition{}))
			Ω(pol.Definition["max-length"]).Should(BeEquivalentTo(100))
			Ω(pol.Definition["max-length-bytes"]).Should(Equal(float64(100000)))
			Ω(pol.Definition["message-ttl"]).Should(Equal(float64(606)))
			Ω(pol.Definition["expires"]).Should(BeNil())

			// cleanup
			_, err = rmqc.DeleteOperatorPolicy(vh, name)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetOperatorPolicy(vh, name)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("GET /api/parameters/federation-upstream", func() {
		Context("when there are no upstreams", func() {
			It("returns an empty response", func() {
				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreams()
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
			})
		})

		Context("when there are upstreams", func() {
			It("returns the list of upstreams", func() {
				def1 := FederationDefinition{
					Uri: URISet{"amqp://server-name/%2f"},
				}

				vh := "rabbit/hole"
				name1 := "upstream1"
				name2 := "upstream2"

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreams()
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())

				_, err := rmqc.PutFederationUpstream(vh, name1, def1)
				Ω(err).Should(BeNil())

				def2 := FederationDefinition{
					Uri: URISet{"amqp://example.com/%2f"},
				}
				_, err = rmqc.PutFederationUpstream("/", name2, def2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreams()
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

				list, err := rmqc.ListFederationUpstreams()
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(2))

				_, err = rmqc.DeleteFederationUpstream(vh, name1)
				Ω(err).Should(BeNil())

				_, err = rmqc.DeleteFederationUpstream("/", name2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreams()
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())

			})
		})
	})

	Context("GET /api/parameters/federation-upstream/{vhost}", func() {
		Context("when there are no upstreams", func() {
			It("returns an empty response", func() {
				vh := "rabbit/hole"

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreamsIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
			})
		})

		Context("when there are upstreams", func() {
			It("returns the list of upstreams", func() {
				vh := "rabbit/hole"

				def1 := FederationDefinition{
					Uri: URISet{"amqp://server-name/%2f"},
				}

				_, err := rmqc.PutFederationUpstream(vh, "vhost-upstream1", def1)
				Ω(err).Should(BeNil())

				def2 := FederationDefinition{
					Uri: URISet{"amqp://example.com/%2f"},
				}

				_, err = rmqc.PutFederationUpstream(vh, "vhost-upstream2", def2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreamsIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

				list, err := rmqc.ListFederationUpstreamsIn(vh)
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(2))

				// delete vhost-upstream1
				_, err = rmqc.DeleteFederationUpstream(vh, "vhost-upstream1")
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreamsIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

				list, err = rmqc.ListFederationUpstreamsIn(vh)
				Ω(err).Should(BeNil())
				Ω(len(list)).Should(Equal(1))

				// delete vhost-upstream2
				_, err = rmqc.DeleteFederationUpstream(vh, "vhost-upstream2")
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreamsIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
			})
		})
	})

	Context("GET /api/parameters/federation-upstream/{vhost}/{upstream}", func() {
		Context("when the upstream does not exist", func() {
			It("returns a 404 error", func() {
				vh := "rabbit/hole"
				name := "non-existing-upstream"

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(up).Should(BeNil())
				Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
			})
		})

		Context("when the upstream exists", func() {
			It("returns the upstream", func() {
				vh := "rabbit/hole"
				name := "valid-upstream"

				def := FederationDefinition{
					Uri:            []string{"amqp://127.0.0.1/%2f"},
					PrefetchCount:  1000,
					ReconnectDelay: 1,
					AckMode:        "on-confirm",
					TrustUserId:    false,
				}

				_, err := rmqc.PutFederationUpstream(vh, name, def)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetFederationUpstream(vh, name)
					Ω(err).Should(BeNil())

					return x.Name
				}).Should(Equal(name))

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))
				Ω(up.Component).Should(Equal(FederationUpstreamComponent))
				Ω(up.Definition.Uri).Should(ConsistOf(def.Uri))
				Ω(up.Definition.PrefetchCount).Should(Equal(def.PrefetchCount))
				Ω(up.Definition.ReconnectDelay).Should(Equal(def.ReconnectDelay))
				Ω(up.Definition.AckMode).Should(Equal(def.AckMode))
				Ω(up.Definition.TrustUserId).Should(Equal(def.TrustUserId))

				_, err = rmqc.DeleteFederationUpstream(vh, name)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreamsIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
			})
		})
	})

	Context("PUT /api/parameters/federation-upstream/{vhost}/{upstream}", func() {
		Context("when the upstream does not exist", func() {
			It("creates the upstream", func() {
				vh := "rabbit/hole"
				name := "create-upstream"

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(up).Should(BeNil())
				Ω(err).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))

				def := FederationDefinition{
					Uri:            []string{"amqp://127.0.0.1/%2f"},
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

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetFederationUpstream(vh, name)
					Ω(err).Should(BeNil())

					return x.Name
				}).Should(Equal(name))

				up, err = rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))
				Ω(up.Component).Should(Equal(FederationUpstreamComponent))
				Ω(up.Definition.Uri).Should(ConsistOf(def.Uri))
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

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreamsIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
			})
		})

		Context("when the upstream exists", func() {
			It("updates the upstream", func() {
				vh := "rabbit/hole"
				name := "update-upstream"

				// create the upstream
				def := FederationDefinition{
					Uri:            []string{"amqp://127.0.0.1/%2f"},
					PrefetchCount:  1000,
					ReconnectDelay: 1,
					AckMode:        "on-confirm",
					TrustUserId:    false,
				}

				_, err := rmqc.PutFederationUpstream(vh, name, def)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetFederationUpstream(vh, name)
					Ω(err).Should(BeNil())

					return x.Name
				}).Should(Equal(name))

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))
				Ω(up.Component).Should(Equal(FederationUpstreamComponent))
				Ω(up.Definition.Uri).Should(ConsistOf(def.Uri))
				Ω(up.Definition.PrefetchCount).Should(Equal(def.PrefetchCount))
				Ω(up.Definition.ReconnectDelay).Should(Equal(def.ReconnectDelay))
				Ω(up.Definition.AckMode).Should(Equal(def.AckMode))
				Ω(up.Definition.TrustUserId).Should(Equal(def.TrustUserId))

				// update the upstream
				def2 := FederationDefinition{
					Uri:            []string{"amqp://128.0.0.1/%2f", "amqp://128.0.0.7/%2f"},
					PrefetchCount:  500,
					ReconnectDelay: 10,
					AckMode:        "no-ack",
					TrustUserId:    true,
				}

				_, err = rmqc.PutFederationUpstream(vh, name, def2)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetFederationUpstream(vh, name)
					Ω(err).Should(BeNil())

					return x.Name
				}).Should(Equal(name))

				up, err = rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))
				Ω(up.Component).Should(Equal(FederationUpstreamComponent))
				Ω(up.Definition.Uri).Should(ConsistOf(def2.Uri))
				Ω(up.Definition.PrefetchCount).Should(Equal(def2.PrefetchCount))
				Ω(up.Definition.ReconnectDelay).Should(Equal(def2.ReconnectDelay))
				Ω(up.Definition.AckMode).Should(Equal(def2.AckMode))
				Ω(up.Definition.TrustUserId).Should(Equal(def2.TrustUserId))

				_, err = rmqc.DeleteFederationUpstream(vh, name)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreamsIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
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
				name := "non-existing-upstream"

				// an error is not returned by design
				resp, err := rmqc.DeleteFederationUpstream(vh, name)
				Ω(resp.Status).Should(Equal("404 Not Found"))
				Ω(err).Should(BeNil())
			})
		})

		Context("when the upstream exists", func() {
			It("deletes the upstream", func() {
				vh := "rabbit/hole"
				name := "delete-upstream"

				def := FederationDefinition{
					Uri: []string{"amqp://127.0.0.1/%2f"},
				}

				_, err := rmqc.PutFederationUpstream(vh, name, def)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetFederationUpstream(vh, name)
					Ω(err).Should(BeNil())

					return x.Name
				}).Should(Equal(name))

				up, err := rmqc.GetFederationUpstream(vh, name)
				Ω(err).Should(BeNil())
				Ω(up.Vhost).Should(Equal(vh))
				Ω(up.Name).Should(Equal(name))

				_, err = rmqc.DeleteFederationUpstream(vh, name)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []FederationUpstream {
					xs, err := rmqc.ListFederationUpstreamsIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
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
				vh := "rabbit/hole"

				// create upstream
				upstreamName := "myUpsteam"
				def := FederationDefinition{
					Uri:     []string{"amqp://localhost/%2f"},
					Expires: 1800000,
				}

				_, err := rmqc.PutFederationUpstream(vh, upstreamName, def)
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

				_, err = rmqc.PutPolicy(vh, policyName, policy)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) FederationLinkMap {
					xs, err := rmqc.ListFederationLinksIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

				// assertions
				list, err := rmqc.ListFederationLinksIn(vh)
				Ω(len(list)).Should(Equal(1))
				Ω(err).Should(BeNil())

				var link = list[0]
				Ω(link["vhost"]).Should(Equal(vh))
				Ω(link["upstream"]).Should(Equal(upstreamName))
				Ω(link["type"]).Should(Equal("exchange"))
				Ω(link["exchange"]).Should(Equal("amq.topic"))
				Ω(link["uri"]).Should(Equal(def.Uri[0]))
				Ω([]string{"running", "starting"}).Should(ContainElement(link["status"]))

				// cleanup
				_, err = rmqc.DeletePolicy(vh, policyName)
				Ω(err).Should(BeNil())

				_, err = rmqc.DeleteFederationUpstream(vh, upstreamName)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) FederationLinkMap {
					xs, err := rmqc.ListFederationLinksIn(vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
			})
		})
	})

	Context("PUT /parameters/shovel/{vhost}/{name}", func() {
		It("declares a shovel using AMQP 1.0 protocol", func() {
			vh := "rabbit/hole"
			sn := "temporary"

			ssu := URISet([]string{"amqp://127.0.0.1/%2f"})
			sdu := URISet([]string{"amqp://127.0.0.1/%2f", "amqp://localhost/%2f"})

			shovelDefinition := ShovelDefinition{
				AckMode:                          "on-confirm",
				ReconnectDelay:                   20,
				SourceURI:                        ssu,
				SourceAddress:                    "mySourceQueue",
				SourceProtocol:                   "amqp10",
				SourcePrefetchCount:              42,
				SourceDeleteAfter:                "never",
				DestinationURI:                   sdu,
				DestinationProtocol:              "amqp10",
				DestinationAddress:               "myDestQueue",
				DestinationAddForwardHeaders:     true,
				DestinationAddTimestampHeader:    true,
				DestinationApplicationProperties: map[string]interface{}{"key": "value"},
				DestinationMessageAnnotations:    map[string]interface{}{"annotation": "something"},
				DestinationProperties:            map[string]interface{}{"prop0": "value0"}}

			_, err := rmqc.DeclareShovel(vh, sn, shovelDefinition)
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetShovel(vh, sn)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(sn))

			x, err := rmqc.GetShovel(vh, sn)
			Ω(err).Should(BeNil())
			Ω(x.Name).Should(Equal(sn))
			Ω(x.Vhost).Should(Equal(vh))
			Ω(x.Component).Should(Equal("shovel"))
			Ω(x.Definition.AckMode).Should(Equal("on-confirm"))
			Ω(x.Definition.ReconnectDelay).Should(Equal(20))
			Ω(x.Definition.SourceAddress).Should(Equal("mySourceQueue"))
			Ω(x.Definition.SourceURI).Should(Equal(ssu))
			Ω(x.Definition.SourcePrefetchCount).Should(Equal(42))
			Ω(x.Definition.SourceProtocol).Should(Equal("amqp10"))
			Ω(string(x.Definition.SourceDeleteAfter)).Should(Equal("never"))
			Ω(x.Definition.DestinationAddress).Should(Equal("myDestQueue"))
			Ω(x.Definition.DestinationURI).Should(Equal(sdu))
			Ω(x.Definition.DestinationProtocol).Should(Equal("amqp10"))
			Ω(x.Definition.DestinationAddForwardHeaders).Should(Equal(true))
			Ω(x.Definition.DestinationAddTimestampHeader).Should(Equal(true))
			Ω(x.Definition.DestinationApplicationProperties).Should(HaveKeyWithValue("key", "value"))
			Ω(x.Definition.DestinationMessageAnnotations).Should(HaveKeyWithValue("annotation", "something"))
			Ω(x.Definition.DestinationProperties).Should(HaveKeyWithValue("prop0", "value0"))

			_, err = rmqc.DeleteShovel(vh, sn)
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteQueue("/", "mySourceQueue")
			Ω(err).Should(BeNil())
			_, err = rmqc.DeleteQueue("/", "myDestQueue")
			Ω(err).Should(BeNil())

			x, err = rmqc.GetShovel(vh, sn)
			Eventually(func(g Gomega) error {
				_, err := rmqc.GetShovel(vh, sn)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("PUT /parameters/shovel/{vhost}/{name}", func() {
		It("declares a shovel", func() {
			vh := "rabbit/hole"
			sn := "temporary"

			ssu := URISet([]string{"amqp://127.0.0.1/%2f"})
			sdu := URISet([]string{"amqp://127.0.0.1/%2f"})

			shovelDefinition := ShovelDefinition{
				AckMode:                       "on-confirm",
				ReconnectDelay:                20,
				SourceURI:                     ssu,
				SourceQueue:                   "mySourceQueue",
				SourceQueueArgs:               map[string]interface{}{"x-message-ttl": 12000},
				SourceConsumerArgs:            map[string]interface{}{"x-priority": 2},
				SourcePrefetchCount:           5,
				DestinationURI:                sdu,
				DestinationQueue:              "myDestQueue",
				DestinationQueueArgs:          map[string]interface{}{"x-expires": 222000},
				AddForwardHeaders:             true,
				DestinationAddTimestampHeader: true,
				DestinationPublishProperties:  map[string]interface{}{"delivery_mode": 1},
				DeleteAfter:                   "never"}

			_, err := rmqc.DeclareShovel(vh, sn, shovelDefinition)
			Ω(err).Should(BeNil(), "Error declaring shovel")

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetShovel(vh, sn)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(sn))

			x, err := rmqc.GetShovel(vh, sn)
			Ω(err).Should(BeNil(), "Error getting shovel")
			Ω(x.Name).Should(Equal(sn))
			Ω(x.Vhost).Should(Equal(vh))
			Ω(x.Component).Should(Equal("shovel"))
			Ω(x.Definition.AckMode).Should(Equal("on-confirm"))
			Ω(x.Definition.ReconnectDelay).Should(Equal(20))
			Ω(x.Definition.SourceURI).Should(Equal(ssu))
			Ω(x.Definition.SourceQueue).Should(Equal("mySourceQueue"))
			Ω(x.Definition.SourceQueueArgs).Should(HaveKeyWithValue("x-message-ttl", float64(12000)))
			Ω(x.Definition.SourceConsumerArgs).Should(HaveKeyWithValue("x-priority", float64(2)))
			Ω(x.Definition.SourcePrefetchCount).Should(Equal(5))
			Ω(x.Definition.DestinationURI).Should(Equal(sdu))
			Ω(x.Definition.DestinationQueue).Should(Equal("myDestQueue"))
			Ω(x.Definition.DestinationQueueArgs).Should(HaveKeyWithValue("x-expires", float64(222000)))
			Ω(x.Definition.AddForwardHeaders).Should(Equal(true))
			Ω(x.Definition.DestinationAddTimestampHeader).Should(Equal(true))
			Ω(x.Definition.DestinationPublishProperties).Should(HaveKeyWithValue("delivery_mode", float64(1)))
			Ω(string(x.Definition.DeleteAfter)).Should(Equal("never"))

			_, err = rmqc.DeleteShovel(vh, sn)
			Ω(err).Should(BeNil())

			_, err = rmqc.DeleteQueue("/", "mySourceQueue")
			Ω(err).Should(BeNil())
			_, err = rmqc.DeleteQueue("/", "myDestQueue")
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetShovel(vh, sn)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})

		It("declares a shovel with a numeric delete-after value", func() {
			vh := "rabbit/hole"
			sn := "temporary"

			ssu := URISet([]string{"amqp://127.0.0.1/%2f"})
			sdu := URISet([]string{"amqp://127.0.0.1/%2f"})

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

			Eventually(func(g Gomega) string {
				x, err := rmqc.GetShovel(vh, sn)
				Ω(err).Should(BeNil())

				return x.Name
			}).Should(Equal(sn))

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

			_, err = rmqc.DeleteQueue("/", "mySourceQueue")
			Ω(err).Should(BeNil())
			_, err = rmqc.DeleteQueue("/", "myDestQueue")
			Ω(err).Should(BeNil())

			Eventually(func(g Gomega) error {
				_, err := rmqc.GetShovel(vh, sn)

				return err
			}).Should(Equal(ErrorResponse{404, "Object Not Found", "Not Found"}))
		})
	})

	Context("GET /shovels/{vhost}", func() {
		It("reports status of Shovels in target virtual host", func() {
			vh := "rabbit/hole"
			sn := "temporary"

			ssu := URISet([]string{"amqp://localhost/%2f"})
			sdu := URISet([]string{"amqp://localhost/%2f"})

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

			Eventually(func(g Gomega) []ShovelStatus {
				xs, err := rmqc.ListShovelStatus(vh)
				Ω(err).Should(BeNil())

				return xs
			}).ShouldNot(BeEmpty())

			xs, err := rmqc.ListShovelStatus(vh)
			Ω(err).Should(BeNil(), "Error getting shovel")
			Expect(xs[0].Vhost).To(Equal(vh))

			_, err = rmqc.DeleteShovel(vh, sn)
			Ω(err).Should(BeNil())
		})
	})

	Context("PUT /api/parameters/{component}/{vhost}/{name}", func() {
		Context("when the parameter does not exist", func() {
			It("creates the parameter", func() {
				component := FederationUpstreamComponent
				vh := "rabbit/hole"
				name := "temporary"

				pv := RuntimeParameterValue{
					"uri":             "amqp://server-name",
					"prefetch-count":  500,
					"reconnect-delay": 5,
					"ack-mode":        "on-confirm",
					"trust-user-id":   false,
				}

				_, err := rmqc.PutRuntimeParameter(component, vh, name, pv)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetRuntimeParameter(component, vh, name)
					Ω(err).Should(BeNil())

					return x.Name
				}).Should(Equal(name))

				p, err := rmqc.GetRuntimeParameter(component, vh, name)

				Ω(err).Should(BeNil())
				Ω(p.Component).Should(Equal(FederationUpstreamComponent))
				Ω(p.Vhost).Should(Equal(vh))
				Ω(p.Name).Should(Equal(name))

				// we need to convert from interface{}
				v, ok := p.Value.(map[string]interface{})
				Ω(ok).Should(BeTrue())
				Ω(v["uri"]).Should(Equal(pv["uri"]))
				Ω(v["prefetch-count"]).Should(BeNumerically("==", pv["prefetch-count"]))
				Ω(v["reconnect-delay"]).Should(BeNumerically("==", pv["reconnect-delay"]))

				_, err = rmqc.DeleteRuntimeParameter(component, vh, name)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []RuntimeParameter {
					xs, err := rmqc.ListRuntimeParametersIn(component, vh)
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())
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
					Uri: []string{"amqp://server-name/%2f"},
				}

				vh := "rabbit/hole"
				name := "upstream1"

				_, err = rmqc.PutFederationUpstream(vh, name, fDef)
				Ω(err).Should(BeNil())

				sDef := ShovelDefinition{
					SourceURI:         URISet([]string{"amqp://127.0.0.1/%2f"}),
					SourceQueue:       "mySourceQueue",
					DestinationURI:    URISet([]string{"amqp://127.0.0.1/%2f"}),
					DestinationQueue:  "myDestQueue",
					AddForwardHeaders: true,
					AckMode:           "on-confirm",
					DeleteAfter:       "never",
				}

				_, err = rmqc.DeclareShovel("/", "shovel1", sDef)
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []RuntimeParameter {
					xs, err := rmqc.ListRuntimeParameters()
					Ω(err).Should(BeNil())

					return xs
				}).ShouldNot(BeEmpty())

				ps, err := rmqc.ListRuntimeParameters()
				Ω(err).Should(BeNil())
				Ω(len(ps)).Should(Equal(2))

				_, err = rmqc.DeleteFederationUpstream(vh, name)
				Ω(err).Should(BeNil())

				_, err = rmqc.DeleteShovel("/", "shovel1")
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) []RuntimeParameter {
					xs, err := rmqc.ListRuntimeParameters()
					Ω(err).Should(BeNil())

					return xs
				}).Should(BeEmpty())

				// cleanup
				_, err = rmqc.DeleteQueue("/", "mySourceQueue")
				Ω(err).Should(BeNil())
				_, err = rmqc.DeleteQueue("/", "myDestQueue")
				Ω(err).Should(BeNil())
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

		It("lists deprecated feature flags", func() {
			By("GET /api/deprecated-features")
			deprecatedFeatures, err := rmqc.ListDeprecatedFeatures()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(deprecatedFeatures).ShouldNot(BeEmpty())
		})

		It("lists deprecated feature flags in use", func() {
			// TODO: Enable this test after https://github.com/rabbitmq/rabbitmq-server/issues/12619 is fixed
			Skip("not possible to setup RabbitMQ 4.0 to report expected output")

			// Setup
			const queue = "transient.nonexcl.qu"
			_, err := rmqc.DeclareQueue("rabbit/hole", queue, QueueSettings{
				Type:       "classic",
				Durable:    false,
				AutoDelete: false,
			})
			Ω(err).ToNot(HaveOccurred())
			DeferCleanup(func() {
				_, _ = rmqc.DeleteQueue("rabbit/hole", queue)
			})

			By("GET /api/deprecated-features/used")
			deprecatedFeaturesUsed, err := rmqc.ListDeprecatedFeaturesUsed()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(deprecatedFeaturesUsed).ShouldNot(BeEmpty())
			Ω(deprecatedFeaturesUsed).
				To(ContainElement(
					DeprecatedFeature{
						Name:             "transient_nonexcl_queues",
						Description:      "",
						Phase:            DeprecationPermittedByDefault,
						DocumentationUrl: "https://blog.rabbitmq.com/posts/2021/08/4.0-deprecation-announcements/#removal-of-transient-non-exclusive-queues",
						ProvidedBy:       "rabbit",
					}),
				)
		})
	})

	Context("definition export", func() {
		It("returns exported definitions", func() {
			By("GET /definitions")
			defs, err := rmqc.ListDefinitions()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(defs).ShouldNot(BeNil())

			Ω(defs.RabbitMQVersion).ShouldNot(BeNil())

			Ω(*defs.Vhosts).ShouldNot(BeEmpty())
			Ω(*defs.Users).ShouldNot(BeEmpty())

			Ω(*defs.Queues).ShouldNot(BeNil())
			Ω(defs.Parameters).Should(BeNil())
			Ω(*defs.Policies).ShouldNot(BeNil())
		})

		It("returns definitions for a specific vhost", func() {
			By("GET /definitions/vhost")
			defs, err := rmqc.ListVhostDefinitions("rabbit/hole")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(defs).ShouldNot(BeNil())

			Ω(defs.RabbitMQVersion).ShouldNot(BeNil())

			// /definitions/vhost does not export vhosts or users
			Ω(defs.Vhosts).Should(BeNil())
			Ω(defs.Users).Should(BeNil())

			Ω(*defs.Queues).ShouldNot(BeNil())
			Ω(defs.Parameters).Should(BeNil())
			Ω(*defs.Policies).ShouldNot(BeNil())
		})

		It("returns exported global parameters", func() {
			By("GET /definitions")
			defs, err := rmqc.ListDefinitions()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(defs).ShouldNot(BeNil())

			foundClusterName := false
			for _, param := range *defs.GlobalParameters {
				if param.Name == "cluster_name" {
					foundClusterName = true
				}
			}
			Ω(foundClusterName).Should(Equal(true))
		})
	})

	Context("Global parameters", func() {
		Context("PUT /api/global-parameters/name/{value}", func() {
			When("the parameter does not exist", func() {
				It("works", func() {
					By("creating the parameter")
					name := "parameter-name"
					value := map[string]interface{}{
						"endpoints": []string{"amqp://server-name"},
					}

					_, err := rmqc.PutGlobalParameter(name, value)
					Ω(err).Should(BeNil())

					Eventually(func(g Gomega) string {
						x, err := rmqc.GetGlobalParameter(name)
						Ω(err).Should(BeNil())

						return x.Name
					}).Should(Equal(name))

					By("getting the parameter")
					p, err := rmqc.GetGlobalParameter(name)

					Ω(err).Should(BeNil())
					Ω(p.Name).Should(Equal(name))

					v, ok := p.Value.(map[string]interface{})
					Ω(ok).Should(BeTrue())
					endpoints := v["endpoints"].([]interface{})
					Ω(endpoints).To(HaveLen(1))
					Ω(endpoints[0].(string)).Should(Equal("amqp://server-name"))

					By("deleting the parameter")
					_, err = rmqc.DeleteGlobalParameter(name)
					Ω(err).Should(BeNil())
				})
			})
		})

		Context("GET /api/global-parameters", func() {
			It("returns all global parameters", func() {
				name1 := "a-name"
				name2 := "another-name"

				_, err := rmqc.PutGlobalParameter(name1, "a-value")
				Ω(err).Should(BeNil())
				_, err = rmqc.PutGlobalParameter(name2, []string{"another-value"})
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) string {
					x, err := rmqc.GetGlobalParameter(name1)
					Ω(err).Should(BeNil())

					return x.Name
				}).Should(Equal(name1))

				list, err := rmqc.ListGlobalParameters()
				Ω(err).Should(BeNil())
				Ω(list).To(SatisfyAll(
					HaveLen(4), // cluster_name and internal_cluster_id are set by default by RabbitMQ
					ContainElements(
						GlobalRuntimeParameter{
							Name:  "a-name",
							Value: "a-value",
						},
						GlobalRuntimeParameter{
							Name:  "another-name",
							Value: []interface{}{"another-value"},
						},
					),
				))

				_, err = rmqc.DeleteGlobalParameter("a-name")
				Ω(err).Should(BeNil())
				_, err = rmqc.DeleteGlobalParameter("another-name")
				Ω(err).Should(BeNil())

				Eventually(func(g Gomega) int {
					xs, err := rmqc.ListGlobalParameters()
					Ω(err).Should(BeNil())

					return len(xs)
				}).Should(Equal(2))
			})
		})
	})

	Context("POST /api/definitions", func() {
		queueName, exchangeName := "definitions_test_queue", "definitions_test_exchange"
		bi := BindingInfo{
			Source:          exchangeName,
			Vhost:           "/",
			DestinationType: "queue",
			Destination:     queueName,
			Arguments:       map[string]interface{}{},
		}

		It("should create queues and exchanges as specified in the definitions", func() {
			defsToUpload := &ExportedDefinitions{
				Policies: &[]PolicyDefinition{},
				Queues: &[]QueueInfo{{
					Name:      queueName,
					Vhost:     "/",
					Durable:   true,
					Arguments: map[string]interface{}{},
				}},
				Exchanges: &[]ExchangeInfo{{
					Name:      exchangeName,
					Vhost:     "/",
					Durable:   true,
					Type:      "direct",
					Arguments: map[string]interface{}{},
				}},
				Bindings: &[]BindingInfo{bi},
			}
			_, err := rmqc.UploadDefinitions(defsToUpload)
			Expect(err).Should(BeNil())

			defs, err := rmqc.ListDefinitions()
			Expect(err).Should(BeNil())

			queueDefs := defs.Queues
			q := FindQueueByName(*queueDefs, queueName)
			Ω(q).ShouldNot(BeNil())

			exchangeDefs := defs.Exchanges
			x := FindExchangeByName(*exchangeDefs, exchangeName)
			Ω(x).ShouldNot(BeNil())

			bindingDefs := defs.Bindings
			b := FindBindingBySourceAndDestinationNames(*bindingDefs, exchangeName, queueName)
			Ω(b).ShouldNot(BeNil())

			_, _ = rmqc.DeleteExchange("/", exchangeName)
			_, _ = rmqc.DeleteQueue("/", queueName)
			_, _ = rmqc.DeleteBinding("/", bi)
		})
	})

	Context("POST /api/definitions/vhost", func() {
		vhost, queueName, exchangeName := "rabbit/hole", "definitions_test_queue", "definitions_test_exchange"
		bi := BindingInfo{
			Source:          exchangeName,
			DestinationType: "queue",
			Destination:     queueName,
			Arguments:       map[string]interface{}{},
		}

		It("should create queues and exchanges as specified in the definitions for vhost "+vhost, func() {
			defsToUpload := &ExportedDefinitions{
				Policies: &[]PolicyDefinition{},
				Queues: &[]QueueInfo{{
					Name:      queueName,
					Durable:   true,
					Arguments: map[string]interface{}{},
				}},
				Exchanges: &[]ExchangeInfo{{
					Name:      exchangeName,
					Durable:   true,
					Type:      "direct",
					Arguments: map[string]interface{}{},
				}},
				Bindings: &[]BindingInfo{bi},
			}
			_, err := rmqc.UploadVhostDefinitions(defsToUpload, vhost)
			Expect(err).Should(BeNil())

			defs, err := rmqc.ListVhostDefinitions(vhost)
			Expect(err).Should(BeNil())
			Ω(defs).ShouldNot(BeNil())

			queueDefs := defs.Queues
			q := FindQueueByName(*queueDefs, queueName)
			Ω(q).ShouldNot(BeNil())
			Ω(q.Name).Should(Equal("definitions_test_queue"))

			exchangeDefs := defs.Exchanges
			x := FindExchangeByName(*exchangeDefs, exchangeName)
			Ω(x).ShouldNot(BeNil())
			Ω(x.Name).Should(Equal("definitions_test_exchange"))

			bindingDefs := defs.Bindings
			b := FindBindingBySourceAndDestinationNames(*bindingDefs, exchangeName, queueName)
			Ω(b).ShouldNot(BeNil())

			_, _ = rmqc.DeleteExchange("/", exchangeName)
			_, _ = rmqc.DeleteQueue("/", queueName)
			_, _ = rmqc.DeleteBinding("/", bi)
		})
	})
})
