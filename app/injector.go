package app

import (
	"fmt"
	"log"

	egress "github.com/cloudfoundry/statsd-injector/internal/egress/v2"
	ingress "github.com/cloudfoundry/statsd-injector/internal/ingress/v2"
	"github.com/cloudfoundry/statsd-injector/internal/plumbing"
	loggregator "github.com/cloudfoundry/statsd-injector/internal/plumbing/v2"
	"google.golang.org/grpc"
)

type Config struct {
	StatsdHost string
	StatsdPort uint
	MetronPort uint

	CA   string
	Cert string
	Key  string

	DeploymentName string
	JobName        string
	IPAddr         string
	InstanceIndex  string
}

type Injector struct {
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

func NewInjector(c Config) *Injector {
	return &Injector{
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

func (i *Injector) Start() {
	inputChan := make(chan *loggregator.Envelope)
	hostport := fmt.Sprintf("localhost:%d", i.statsdPort)

	metaData := ingress.ProcessMetaData{
		Deployment: i.deploymentName,
		Job:        i.jobName,
		Index:      i.instanceIndex,
		IP:         i.ipAddr,
	}
	_, addr := ingress.Start(hostport, inputChan, metaData)

	log.Printf("Started statsd-injector listener at %s", addr)

	credentials := plumbing.NewCredentials(i.cert, i.key, i.ca, "metron")
	if credentials == nil {
		log.Fatal("Invalid TLS credentials")
	}
	statsdEmitter := egress.New(fmt.Sprintf("localhost:%d", i.metronPort),
		grpc.WithTransportCredentials(credentials),
	)
	statsdEmitter.Run(inputChan)
}
