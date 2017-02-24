package app

import (
	"fmt"
	"log"

	egress "github.com/cloudfoundry/statsd-injector/internal/egress/v2"
	ingress "github.com/cloudfoundry/statsd-injector/internal/ingress/v2"
	"github.com/cloudfoundry/statsd-injector/plumbing"
	"github.com/cloudfoundry/statsd-injector/plumbing/v2"
	"google.golang.org/grpc"
)

type ExperimentalInjector struct {
	statsdPort uint
	apiVersion string
	metronPort uint

	ca   string
	cert string
	key  string

	deploymentName string
	jobName        string
	ipAddr         string
	instanceIndex  string
}

func (e *ExperimentalInjector) Start() {
	inputChan := make(chan *loggregator_v2.Envelope)
	hostport := fmt.Sprintf("localhost:%d", e.statsdPort)

	metaData := ingress.ProcessMetaData{
		Deployment: e.deploymentName,
		Job:        e.jobName,
		Index:      e.instanceIndex,
		IP:         e.ipAddr,
	}
	_, addr := ingress.Start(hostport, inputChan, metaData)

	log.Printf("Started statsd-injector listener at %s", addr)

	credentials := plumbing.NewCredentials(e.cert, e.key, e.ca, "metron")
	if credentials == nil {
		log.Fatal("Invalid TLS credentials")
	}
	statsdEmitter := egress.New(fmt.Sprintf("localhost:%d", e.metronPort),
		grpc.WithTransportCredentials(credentials),
	)
	statsdEmitter.Run(inputChan)
}
