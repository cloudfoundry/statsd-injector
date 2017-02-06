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
)

func main() {
	flag.Parse()

	log.Print("Starting statsd injector")
	defer log.Print("statsd injector closing")

	inputChan := make(chan *v2.Envelope)
	hostport := fmt.Sprintf("%s:%d", *statsdHost, *statsdPort)

	_, addr := statsdlistener.Start(hostport, inputChan)

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
