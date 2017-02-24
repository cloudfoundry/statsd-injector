package app

import (
	"fmt"
	"log"

	"github.com/cloudfoundry/sonde-go/events"
	egress "github.com/cloudfoundry/statsd-injector/internal/egress/v1"
	ingress "github.com/cloudfoundry/statsd-injector/internal/ingress/v1"
)

type DefaultInjector struct {
	statsdHost string
	statsdPort uint
	metronPort uint
}

func (d *DefaultInjector) Start() {
	log.Print("Starting statsd injector")

	hostport := fmt.Sprintf("%s:%d", d.statsdHost, d.statsdPort)
	statsdMessageListener := ingress.New(hostport)
	statsdEmitter := egress.New(d.metronPort)

	inputChan := make(chan *events.Envelope)

	go statsdMessageListener.Run(inputChan)
	statsdEmitter.Run(inputChan)
}
