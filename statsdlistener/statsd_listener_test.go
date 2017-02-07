package statsdlistener_test

import (
	"net"

	v2 "github.com/cloudfoundry/statsd-injector/plumbing/v2"
	"github.com/cloudfoundry/statsd-injector/statsdlistener"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatsdListener", func() {
	Describe("Run", func() {
		var (
			envelopeChan chan *v2.Envelope
			listener     *statsdlistener.StatsdListener
			connection   net.Conn
			metaData     statsdlistener.ProcessMetaData
		)

		BeforeEach(func() {
			envelopeChan = make(chan *v2.Envelope, 100)
			metaData = statsdlistener.ProcessMetaData{
				Deployment: "deployment",
				Job:        "job-name",
				Index:      "instance-index",
				IP:         "some-ip-addr",
			}

			var addr string
			listener, addr = statsdlistener.Start("localhost:0", envelopeChan, metaData)

			var err error
			connection, err = net.Dial("udp4", addr)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			listener.Stop()
			connection.Close()
		})

		It("reads multiple gauges (on different lines) in the same packet", func() {
			statsdmsg := []byte("fake-origin.test.gauge:23|g\nfake-origin.other.thing:42|g\nfake-origin.sampled.gauge:17.5|g|@0.2")

			f := func() int {
				connection.Write(statsdmsg)
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *v2.Envelope
			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.gauge", 23, "gauge")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "other.thing", 42, "gauge")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "sampled.gauge", 87.5, "gauge")
		}, 5)

		It("processes gauge increment/decrement stats", func() {
			statsdmsg := []byte("fake-origin.test.gauge:23|g\nfake-origin.test.gauge:+7|g\nfake-origin.test.gauge:-5|g")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg)
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.gauge", 23, "gauge")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.gauge", 30, "gauge")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.gauge", 25, "gauge")
		})

		It("reads multiple timings (on different lines) in the same packet", func() {
			statsdmsg := []byte("fake-origin.test.timing:23|ms\nfake-origin.other.thing:420|ms\nfake-origin.sampled.timing:71|ms|@0.1")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg)
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.timing", 23, "ms")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "other.thing", 420, "ms")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "sampled.timing", 710, "ms")
		}, 5)

		It("reads multiple counters (on different lines) in the same packet", func() {
			statsdmsg := []byte("fake-origin.test.counter:23|c\nfake-origin.other.thing:420|c\nfake-origin.sampled.counter:71|c|@0.1")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg)
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.counter", 23, "counter")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "other.thing", 420, "counter")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "sampled.counter", 710, "counter")
		}, 5)

		It("processes counter increment/decrement stats", func() {
			statsdmsg := []byte("fake-origin.test.counter:23|c\nfake-origin.test.counter:+7|c\nfake-origin.test.counter:-5|c")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg)
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.counter", 23, "counter")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.counter", 30, "counter")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.counter", 25, "counter")
		})

		It("adds meta-data tags", func() {
			statsdmsg := []byte("fake-origin.test.counter:23|c\nfake-origin.test.counter:+7|c\nfake-origin.test.counter:-5|c")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg)
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			Expect(receivedEnvelope.GetTags()["origin"].GetText()).To(Equal("fake-origin"))
			Expect(receivedEnvelope.GetTags()["deployment"].GetText()).To(Equal(metaData.Deployment))
			Expect(receivedEnvelope.GetTags()["job"].GetText()).To(Equal(metaData.Job))
			Expect(receivedEnvelope.GetTags()["index"].GetText()).To(Equal(metaData.Index))
			Expect(receivedEnvelope.GetTags()["ip"].GetText()).To(Equal(metaData.IP))
		})
	})
})

func checkValueMetric(receivedEnvelope *v2.Envelope, origin string, name string, value float64, unit string) {
	m, ok := receivedEnvelope.GetGauge().GetMetrics()[name]
	Expect(ok).To(BeTrue())
	Expect(m.GetValue()).To(Equal(value))
	Expect(m.GetUnit()).To(Equal(unit))
}
