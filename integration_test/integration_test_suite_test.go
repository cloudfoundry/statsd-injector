package integration_test

import (
	"os/exec"

	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	"github.com/pivotal-golang/localip"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"os"
	"path"
	"path/filepath"
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
	loggregatorPath, err := getGoPathForPackage("metron")
	Expect(err).ToNot(HaveOccurred(), "loggregator not found. Do you have loggregator in your GOPATH?")

	pathToMetronExecutable, err := gexec.BuildIn(loggregatorPath, "metron")
	Expect(err).ShouldNot(HaveOccurred())

	pathToGoStatsdClient, err = gexec.Build("github.com/cloudfoundry/statsd-injector/tools/statsdGoClient")
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

func getGoPathForPackage(packageName string) (string, error) {
	gopaths := os.Getenv("GOPATH")
	for _, gopath := range filepath.SplitList(gopaths) {
		absolutePath := path.Join(gopath, "src", packageName)
		if _, err := os.Stat(absolutePath); err == nil {
			return gopath, nil
		}
	}

	return "", fmt.Errorf("package %s not found", packageName)
}
