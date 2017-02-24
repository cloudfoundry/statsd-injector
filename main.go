package main

import (
	"flag"

	"github.com/cloudfoundry/statsd-injector/app"
)

const defaultAPIVersion = "v1"

func main() {
	statsdPort := flag.Uint("statsd-port", 8125, "The UDP port the injector will listen on for statsd messages")
	apiVersion := flag.String("metron-api", defaultAPIVersion, "API version of Metron to which to send envelopes")
	metronPort := flag.Uint("metron-port", 3458, "The GRPC port the injector will forward message to")

	ca := flag.String("ca", "", "File path to the CA certificate")
	cert := flag.String("cert", "", "File path to the client TLS cert")
	privateKey := flag.String("key", "", "File path to the client TLS private key")

	deploymentName := flag.String("deployment-name", "", "Deployment name (envelope tag)")
	jobName := flag.String("job-name", "", "Job name (envelope tag)")
	ipAddr := flag.String("ip", "", "IP address of host machine (envelope tag)")
	instanceIndex := flag.String("instance-index", "", "index of job instance")
	flag.Parse()

	injector := app.NewInjector(app.Config{
		StatsdPort:     *statsdPort,
		APIVersion:     *apiVersion,
		MetronPort:     *metronPort,
		CA:             *ca,
		Cert:           *cert,
		Key:            *privateKey,
		DeploymentName: *deploymentName,
		JobName:        *jobName,
		IPAddr:         *ipAddr,
		InstanceIndex:  *instanceIndex,
	})
	injector.Start()
}
