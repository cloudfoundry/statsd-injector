package main

import (
	"flag"
	"fmt"
	"log"

	"google.golang.org/grpc"

	"github.com/cloudfoundry/statsd-injector/plumbing"
	v2 "github.com/cloudfoundry/statsd-injector/plumbing/v2"
	"github.com/cloudfoundry/statsd-injector/statsdemitter"
	"github.com/cloudfoundry/statsd-injector/statsdlistener"
)

var (
	statsdHost = flag.String("statsdHost", "localhost", "The hostname the injector will listen on for statsd messages")
	statsdPort = flag.Uint("statsdPort", 8125, "The UDP port the injector will listen on for statsd messages")
	metronPort = flag.Uint("metronPort", 3458, "The GRPC port the injector will forward message to")

	ca         = flag.String("ca", "", "File path to the CA certificate")
	cert       = flag.String("cert", "", "File path to the client TLS cert")
	privateKey = flag.String("key", "", "File path to the client TLS private key")

	deploymentName = flag.String("deployment-name", "", "Deployment name (envelope tag)")
	jobName        = flag.String("job-name", "", "Job name (envelope tag)")
	ipAddr         = flag.String("ip", "", "IP address of host machine (envelope tag)")
	instanceIndex  = flag.String("instance-index", "", "index of job instance")
)

func main() {
	flag.Parse()

	log.Print("Starting statsd injector")
	defer log.Print("statsd injector closing")

	inputChan := make(chan *v2.Envelope)
	hostport := fmt.Sprintf("%s:%d", *statsdHost, *statsdPort)

	metaData := statsdlistener.ProcessMetaData{
		Deployment: *deploymentName,
		Job:        *jobName,
		Index:      *instanceIndex,
		IP:         *ipAddr,
	}
	_, addr := statsdlistener.Start(hostport, inputChan, metaData)

	log.Printf("Started statsd-injector listener at %s", addr)

	credentials := plumbing.NewCredentials(*cert, *privateKey, *ca, "metron")
	if credentials == nil {
		log.Fatal("Invalid TLS credentials")
	}
	statsdEmitter := statsdemitter.New(fmt.Sprintf("localhost:%d", *metronPort),
		grpc.WithTransportCredentials(credentials),
	)
	statsdEmitter.Run(inputChan)
}
