package integration_test

import (
	"net/http"
	"os/exec"

	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	"github.com/pivotal-golang/localip"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func TestIntegrationTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTest Suite")
}

var metronSession *gexec.Session
var statsdInjectorSession *gexec.Session
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var etcdPort int
var localIPAddress string

var pathToGoStatsdClient string

var _ = BeforeSuite(func() {
	pathToMetronExecutable, err := gexec.Build("github.com/cloudfoundry/loggregator/src/metron")
	Expect(err).ShouldNot(HaveOccurred())

	pathToGoStatsdClient, err = gexec.Build("tools/statsdGoClient")
	Expect(err).NotTo(HaveOccurred())

	command := exec.Command(pathToMetronExecutable, "--config=fixtures/metron.json", "--debug")

	metronSession, err = gexec.Start(command, gexec.NewPrefixedWriter("[o][metron]", GinkgoWriter), gexec.NewPrefixedWriter("[e][metron]", GinkgoWriter))
	Expect(err).ShouldNot(HaveOccurred())

	pathToStatsdInjectorExecutable, err := gexec.Build("github.com/cloudfoundry/statsd-injector")
	Expect(err).NotTo(HaveOccurred())
	command = exec.Command(pathToStatsdInjectorExecutable, "-statsdPort=51162", "-logLevel=debug")

	statsdInjectorSession, err = gexec.Start(command, gexec.NewPrefixedWriter("[o][statsdInjector]", GinkgoWriter), gexec.NewPrefixedWriter("[e][statsdInjector]", GinkgoWriter))
	Expect(err).ShouldNot(HaveOccurred())

	localIPAddress, _ = localip.LocalIP()

	// wait for server to be up
	Eventually(func() error {
		_, err := http.Get("http://" + localIPAddress + ":1234")
		return err
	}, 3).ShouldNot(HaveOccurred())

	etcdPort = 5800 + (config.GinkgoConfig.ParallelNode-1)*10
	etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1)
	etcdRunner.Start()

	Eventually(statsdInjectorSession).Should(gbytes.Say("Listening for statsd on host localhost:51162"))
})

var _ = AfterSuite(func() {
	metronSession.Kill().Wait()
	statsdInjectorSession.Kill().Wait()
	gexec.CleanupBuildArtifacts()

	etcdRunner.Adapter().Disconnect()
	etcdRunner.Stop()
})
