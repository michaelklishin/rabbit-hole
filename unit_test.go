package rabbithole

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit tests", func() {
	Context("DeleteAfter marshalling", func() {
		It("unmarshalls DeleteAfter when it is a number", func() {
			var d DeleteAfter
			s := []byte("1")
			err := d.UnmarshalJSON(s)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(d).Should(Equal(DeleteAfter("1")))
		})

		It("unmarshalls DeleteAfter when it is a quoted string", func() {
			var d DeleteAfter
			s := []byte("\"3\"")
			err := d.UnmarshalJSON(s)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(d).Should(Equal(DeleteAfter("3")))
		})
	})

	Context("URISet marshalling", func() {
		It("unmarshalls a single string", func() {
			var us URISet
			bs := []byte("\"amqp://127.0.0.1:5672\"")
			err := us.UnmarshalJSON(bs)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(us).Should(Equal(URISet([]string{"amqp://127.0.0.1:5672"})))
		})

		It("unmarshalls a list of strings", func() {
			var us URISet
			bs := []byte("[\"amqp://127.0.0.1:5672\", \"amqp://localhost:5672\"]")
			err := us.UnmarshalJSON(bs)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(us).Should(Equal(URISet([]string{"amqp://127.0.0.1:5672", "amqp://localhost:5672"})))
		})
	})

	Context("Port marshalling", func() {
		It("unmarshal Port when it is a number", func() {
			var d Port
			s := []byte("123")
			err := d.UnmarshalJSON(s)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(d).Should(Equal(Port(123)))
		})

		It("unmarshal Port when it is a quoted string", func() {
			var d Port
			s := []byte("\"456\"")
			err := d.UnmarshalJSON(s)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(d).Should(Equal(Port(456)))
		})

		It("unmarshal Port when it is a undefined", func() {
			var d Port
			s := []byte("\"undefined\"")
			err := d.UnmarshalJSON(s)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(d).Should(Equal(Port(0)))
		})
	})
	Context("ConsumerDetails marshalling", func() {
		It("unmarshal ConsumerDetails when channel_details is an empty array", func() {
			var d ConsumerDetail
			s := []byte("{\"channel_details\":[]}")
			err := json.Unmarshal(s, &d)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(d.ChannelDetails).Should(Equal(ChannelDetails{}))
		})
		It("unmarshal ConsumerDetails when channel_details is an object", func() {
			var d ConsumerDetail
			s := []byte("{\"channel_details\":{\"name\":\"foo\"}}")
			err := json.Unmarshal(s, &d)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(d.ChannelDetails).Should(Equal(ChannelDetails{Name: "foo"}))
		})
	})

	Context("compareVersions", func() {
		DescribeTable("compare version strings", func(version1, version2 string, expected int) {
			Ω(compareVersions(version1, version2)).Should(HaveValue(Equal(expected)))
		},
			Entry("3.13 < 4.0", "3.13.0", "4.0.0", -1),
			Entry("4.0 > 3.13", "4.0", "3.13", 1),
			Entry("4.0.0 < 4.0.1", "4.0.0", "4.0.1", -1),
			Entry("4.0 < 4.0.0", "4.0", "4.0.0", -1),
			// not semver correct, but for our interests (feature flagging), only the MAJOR.MINOR.PATCH is relevant
			Entry("4.0.0-beta.1 < 4.0.0", "4.0.0-beta.1", "4.0.0", 0),
			Entry("4.0.0-beta.1 < 4.0.0-beta.2", "4.0.0-beta.1", "4.0.0-beta.2", -1), // edge case to the above 😮‍💨
		)
		DescribeTable("is version higher than 4.4", func(v string, ex bool) {
			Ω(isRabbitVersion44OrLater(v)).To(Equal(ex))
		},
			Entry("4.0", "4.0", false),
			Entry("4.4", "4.4", false), // special case
			Entry("4.4.0", "4.4.0", true),
			Entry("4.4.1", "4.4.1", true),
			Entry("5.0.0", "5.0.0", true),
		)
	})
})
