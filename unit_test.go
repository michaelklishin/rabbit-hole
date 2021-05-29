package rabbithole

import (
	. "github.com/onsi/ginkgo"
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
})
