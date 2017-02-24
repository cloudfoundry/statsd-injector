package app

type Config struct {
	StatsdHost string
	StatsdPort uint
	MetronPort uint
	APIVersion string

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
	switch c.APIVersion {
	case "v2":
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
	default:
		return &DefaultInjector{
			statsdHost: c.StatsdHost,
			statsdPort: c.StatsdPort,
			metronPort: c.MetronPort,
		}
	}
}
