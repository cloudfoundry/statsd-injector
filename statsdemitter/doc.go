package statsdemitter

//go:generate hel

import (
	v2 "github.com/cloudfoundry/statsd-injector/plumbing/v2"
)

type MetronIngressServer interface {
	v2.MetronIngressServer
}
