package roundtrip

type Option func(config *metricsRoundTripper)

func WithProject(project string) Option {
	return func(c *metricsRoundTripper) {
		c.project = project
	}
}

func WithDestination(destination string) Option {
	return func(c *metricsRoundTripper) {
		c.destination = destination
	}
}

func WithOpFromResponse() Option {
	return func(c *metricsRoundTripper) {
		c.opFromResponse = true
	}
}
