package app

import (
	"fmt"
	"log"

	"github.com/cloudfoundry/statsd-injector/internal/egress"
	"github.com/cloudfoundry/statsd-injector/internal/ingress"
	"github.com/cloudfoundry/statsd-injector/plumbing"
	v2 "github.com/cloudfoundry/statsd-injector/plumbing/v2"
	"google.golang.org/grpc"
)

type Config struct {
	StatsdPort uint
	APIVersion string
	MetronPort uint

	CA   string
	Cert string
	Key  string

	DeploymentName string
	JobName        string
	IPAddr         string
	InstanceIndex  string
}

type Starter interface {
	Start()
}

func NewInjector(c Config) Starter {
	return &ExperimentalInjector{
		statsdPort:     c.StatsdPort,
		metronPort:     c.MetronPort,
		ca:             c.CA,
		cert:           c.Cert,
		key:            c.Key,
		deploymentName: c.DeploymentName,
		jobName:        c.JobName,
		ipAddr:         c.IPAddr,
		instanceIndex:  c.InstanceIndex,
	}
}

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
	inputChan := make(chan *v2.Envelope)
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
