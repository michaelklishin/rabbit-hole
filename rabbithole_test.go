package rabbithole

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streadway/amqp"
)

// TODO: extract duplication between these
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
			false, // delete when usused
			true,  // exclusive
			false,
			nil)
		ch.Publish("", q.Name, false, false, amqp.Publishing{Body: []byte("")})
	}
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

			fanoutExchange := ExchangeType{Name: "fanout", Description: "AMQP fanout exchange, as per the AMQP specification", Enabled: true}
			Ω(res.ExchangeTypes).Should(ContainElement(fanoutExchange))

			ch.Close()
		})
	})

	Context("DELETE /api/connections/{name}", func() {
		It("closes the connection", func() {
			conn := openConnection("/")

			xs, err := rmqc.ListConnections()
			Ω(err).Should(BeNil())

			closeEvents := make(chan *amqp.Error)
			conn.NotifyClose(closeEvents)

			n := xs[0].Name
			rmqc.CloseConnection(n)

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
			Ω(res.Tags).ShouldNot(BeNil())
			Ω(res.AuthBackend).ShouldNot(BeNil())
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

			Ω(res.IsRunning).Should(Equal(true))

			Ω(res.SocketsUsed).Should(BeNumerically(">=", 0))
			Ω(res.SocketsTotal).Should(BeNumerically(">=", 1))

		})
	})

	Context("GET /nodes/{name}", func() {
		It("returns decoded response", func() {
			xs, err := rmqc.ListNodes()
			n := xs[0]
			res, err := rmqc.GetNode(n.Name)

			Ω(err).Should(BeNil())

			Ω(res.Name).ShouldNot(BeNil())
			Ω(res.NodeType).Should(Equal("disc"))

			Ω(res.FdUsed).Should(BeNumerically(">=", 0))
			Ω(res.FdTotal).Should(BeNumerically(">", 64))

			Ω(res.MemUsed).Should(BeNumerically(">", 10*1024*1024))
			Ω(res.MemLimit).Should(BeNumerically(">", 64*1024*1024))

			Ω(res.IsRunning).Should(Equal(true))

			Ω(res.SocketsUsed).Should(BeNumerically(">=", 0))
			Ω(res.SocketsTotal).Should(BeNumerically(">=", 1))

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

			xs, err := rmqc.ListConnections()
			Ω(err).Should(BeNil())

			info := xs[0]
			Ω(info.Name).ShouldNot(BeNil())
			// Host should match IPv4 regex. (This is to handle the case where rabbit is in a container or vm)
			Ω(info.Host).Should(MatchRegexp((`(?:(?:2(?:[0-4][0-9]|5[0-5])|[0-1]?[0-9]?[0-9])\.){3}(?:(?:2([0-4][0-9]|5[0-5])|[0-1]?[0-9]?[0-9]))`)))
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
				false, // delete when usused
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

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
			conn := openConnection("rabbit/hole")
			defer conn.Close()

			ch, err := conn.Channel()
			Ω(err).Should(BeNil())
			defer ch.Close()

			_, err = ch.QueueDeclare(
				"q2",  // name
				false, // durable
				false, // delete when usused
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			qs, err := rmqc.ListQueuesIn("rabbit/hole")
			Ω(err).Should(BeNil())

			q := FindQueueByName(qs, "q2")
			Ω(q.Name).Should(Equal("q2"))
			Ω(q.Vhost).Should(Equal("rabbit/hole"))
			Ω(q.Durable).Should(Equal(false))
			Ω(q.Status).ShouldNot(BeNil())
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
				false, // delete when usused
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			q, err := rmqc.GetQueue("rabbit/hole", "q3")
			Ω(err).Should(BeNil())

			Ω(q.Vhost).Should(Equal("rabbit/hole"))
			Ω(q.Durable).Should(Equal(false))
			Ω(q.Status).ShouldNot(BeNil())
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
				false, // delete when usused
				false, // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

			_, err = rmqc.GetQueue("rabbit/hole", q.Name)
			Ω(err).Should(BeNil())

			rmqc.DeleteQueue("rabbit/hole", q.Name)

			qi2, err := rmqc.GetQueue("rabbit/hole", q.Name)
			Ω(err).Should(Equal(errors.New("not found")))
			Ω(qi2).Should(BeNil())
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
			info := UserSettings{Password: "s3krE7", Tags: "policymaker, management"}
			resp, err := rmqc.PutUser("rabbithole", info)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(Equal("204 No Content"))

			u, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(BeNil())

			Ω(u.PasswordHash).ShouldNot(BeNil())
			Ω(u.Tags).Should(Equal("policymaker,management"))
		})
	})

	Context("DELETE /users/{name}", func() {
		It("deletes the user", func() {
			info := UserSettings{Password: "s3krE7", Tags: "management policymaker"}
			_, err := rmqc.PutUser("rabbithole", info)
			Ω(err).Should(BeNil())

			u, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(BeNil())
			Ω(u).ShouldNot(BeNil())

			resp, err := rmqc.DeleteUser("rabbithole")
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(Equal("204 No Content"))

			u2, err := rmqc.GetUser("rabbithole")
			Ω(err).Should(Equal(errors.New("not found")))
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
		})
	})

	Context("GET /vhosts/{name} when vhost exists", func() {
		It("returns decoded response", func() {
			x, err := rmqc.GetVhost("rabbit/hole")
			Ω(err).Should(BeNil())

			Ω(x.Name).ShouldNot(BeNil())
			Ω(x.Tracing).ShouldNot(BeNil())
		})
	})

	Context("PUT /vhosts/{name}", func() {
		It("creates a vhost", func() {
			vs := VhostSettings{Tracing: false}
			_, err := rmqc.PutVhost("rabbit/hole2", vs)
			Ω(err).Should(BeNil())

			x, err := rmqc.GetVhost("rabbit/hole2")

			Ω(x.Name).Should(BeEquivalentTo("rabbit/hole2"))
			Ω(x.Tracing).Should(Equal(false))

			rmqc.DeleteVhost("rabbit/hole2")
		})
	})

	Context("DELETE /vhosts/{name}", func() {
		It("creates a vhost", func() {
			vs := VhostSettings{Tracing: false}
			_, err := rmqc.PutVhost("rabbit/hole2", vs)
			Ω(err).Should(BeNil())

			x, err := rmqc.GetVhost("rabbit/hole2")
			Ω(x).ShouldNot(BeNil())

			rmqc.DeleteVhost("rabbit/hole2")
			x2, err := rmqc.GetVhost("rabbit/hole2")
			Ω(x2).Should(BeNil())
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
				false, // delete when usused
				true,  // exclusive
				false,
				nil)
			Ω(err).Should(BeNil())

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
		It("declares a binding to a queue", func() {
			vh := "rabbit/hole"
			qn := "test.bindings.post.queue"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{})

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

			bs, err := rmqc.ListBindingsIn(vh)
			Ω(err).Should(BeNil())
			Ω(bs).ShouldNot(BeEmpty())

			b := bs[1]
			Ω(b.Source).Should(Equal(info.Source))
			Ω(b.Vhost).Should(Equal(vh))
			Ω(b.Destination).Should(Equal(info.Destination))
			Ω(b.DestinationType).Should(Equal(info.DestinationType))
			Ω(b.RoutingKey).Should(Equal(info.RoutingKey))
			Ω(b.PropertiesKey).Should(Equal(propertiesKey))

			rmqc.DeleteBinding(vh, b)
		})
	})

	Context("POST /bindings/{vhost}/e/{source}/e/{destination}", func() {
		It("declares a binding to an exchange", func() {
			vh := "rabbit/hole"
			xn := "test.bindings.post.exchange"

			_, err := rmqc.DeclareExchange(vh, xn, ExchangeSettings{Type: "topic"})

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

			bs, err := rmqc.ListBindingsIn(vh)
			Ω(err).Should(BeNil())
			Ω(bs).ShouldNot(BeEmpty())

			b := bs[1]
			Ω(b.Source).Should(Equal(info.Source))
			Ω(b.Vhost).Should(Equal(vh))
			Ω(b.Destination).Should(Equal(info.Destination))
			Ω(b.DestinationType).Should(Equal(info.DestinationType))
			Ω(b.RoutingKey).Should(Equal(info.RoutingKey))
			Ω(b.PropertiesKey).Should(Equal(propertiesKey))

			rmqc.DeleteBinding(vh, b)
		})
	})

	Context("DELETE /bindings/{vhost}/e/{source}/e/{destination}/{propertiesKey}", func() {
		It("deletes an individual exchange binding", func() {
			vh := "rabbit/hole"
			xn := "test.bindings.post.exchange"

			rmqc.DeclareExchange(vh, xn, ExchangeSettings{Type: "topic"})

			info := BindingInfo{
				Source:          "amq.topic",
				Destination:     xn,
				DestinationType: "exchange",
				RoutingKey:      "#",
				Arguments: map[string]interface{}{
					"one": "two",
				},
			}

			res, _ := rmqc.DeclareBinding(vh, info)

			// Grab the Location data from the POST response {destination}/{propertiesKey}
			propertiesKey, _ := url.QueryUnescape(strings.Split(res.Header.Get("Location"), "/")[1])
			info.PropertiesKey = propertiesKey

			rmqc.DeleteBinding(vh, info)
			bs, _ := rmqc.ListBindingsIn(vh)

			Ω(bs).Should(HaveLen(1))
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

			fetched, err := rmqc.GetPermissionsIn("/", u)
			Ω(err).Should(BeNil())
			Ω(fetched.Configure).Should(Equal(permissions.Configure))
			Ω(fetched.Write).Should(Equal(permissions.Write))
			Ω(fetched.Read).Should(Equal(permissions.Read))

			rmqc.DeleteUser(u)
		})
	})

	Context("DELETE /permissions/{vhost}/{user}", func() {
		It("clears permissions", func() {
			u := "temporary"

			_, err := rmqc.PutUser(u, UserSettings{Password: "s3krE7"})
			Ω(err).Should(BeNil())

			_, err = rmqc.ClearPermissionsIn("/", u)
			_, err = rmqc.GetPermissionsIn("/", u)
			Ω(err).Should(Equal(errors.New("not found")))

			rmqc.DeleteUser(u)
		})
	})

	Context("PUT /exchanges/{vhost}/{exchange}", func() {
		It("declares an exchange", func() {
			vh := "rabbit/hole"
			xn := "temporary"

			_, err := rmqc.DeclareExchange(vh, xn, ExchangeSettings{Type: "fanout", Durable: false})
			Ω(err).Should(BeNil())

			x, err := rmqc.GetExchange(vh, xn)
			Ω(err).Should(BeNil())
			Ω(x.Name).Should(Equal(xn))
			Ω(x.Durable).Should(Equal(false))
			Ω(x.AutoDelete).Should(Equal(false))
			Ω(x.Type).Should(Equal("fanout"))
			Ω(x.Vhost).Should(Equal(vh))

			rmqc.DeleteExchange(vh, xn)
		})
	})

	Context("DELETE /exchanges/{vhost}/{exchange}", func() {
		It("deletes an exchange", func() {
			vh := "rabbit/hole"
			xn := "temporary"

			_, err := rmqc.DeclareExchange(vh, xn, ExchangeSettings{Type: "fanout", Durable: false})
			Ω(err).Should(BeNil())

			rmqc.DeleteExchange(vh, xn)
			x, err := rmqc.GetExchange(vh, xn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(errors.New("not found")))
		})
	})

	Context("PUT /queues/{vhost}/{queue}", func() {
		It("declares a queue", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())

			x, err := rmqc.GetQueue(vh, qn)
			Ω(err).Should(BeNil())
			Ω(x.Name).Should(Equal(qn))
			Ω(x.Durable).Should(Equal(false))
			Ω(x.AutoDelete).Should(Equal(false))
			Ω(x.Vhost).Should(Equal(vh))

			rmqc.DeleteQueue(vh, qn)
		})
	})

	Context("DELETE /queues/{vhost}/{queue}", func() {
		It("deletes a queue", func() {
			vh := "rabbit/hole"
			qn := "temporary"

			_, err := rmqc.DeclareQueue(vh, qn, QueueSettings{Durable: false})
			Ω(err).Should(BeNil())

			rmqc.DeleteQueue(vh, qn)
			x, err := rmqc.GetQueue(vh, qn)
			Ω(x).Should(BeNil())
			Ω(err).Should(Equal(errors.New("not found")))
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
				Ω(err).Should(Equal(errors.New("not found")))
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

			resp, err := rmqc.DeletePolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(Equal("204 No Content"))
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
			Ω(resp.Status).Should(Equal("204 No Content"))

			_, err = rmqc.GetPolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())

			_, err = rmqc.DeletePolicy("rabbit/hole", "woot")
			Ω(err).Should(BeNil())
		})

		It("updates the policy", func() {
			policy := Policy{
				Pattern:    ".*",
				Definition: PolicyDefinition{"expires": 100, "ha-mode": "all"},
			}

			// create Policy
			_, err := rmqc.PutPolicy("rabbit/hole", "woot", policy)
			Ω(err).Should(BeNil())

			// create new Policy
			new_policy := Policy{
				Pattern: "\\d+",
				ApplyTo: "all",
				Definition: PolicyDefinition{
					"max-length": 100,
					"ha-mode":    "nodes",
					"ha-params":  NodeNames{"a", "b", "c"},
				},
				Priority: 1,
			}

			// update old Policy
			resp, err := rmqc.PutPolicy("/", "woot2", new_policy)
			Ω(err).Should(BeNil())
			Ω(resp.Status).Should(Equal("204 No Content"))

			// old policy should not exist already
			_, err = rmqc.GetPolicy("rabbit/hole", "woot")
			Ω(err).Should(Equal(errors.New("not found")))

			// but new (updated) policy is here
			pol, err := rmqc.GetPolicy("/", "woot2")
			Ω(err).Should(BeNil())
			Ω(pol.Vhost).Should(Equal("/"))
			Ω(pol.Name).Should(Equal("woot2"))
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
			_, err = rmqc.DeletePolicy("/", "woot2")
			Ω(err).Should(BeNil())
		})
	})
})
