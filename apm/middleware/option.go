package middleware

type Option func(config *metricsHandler)

func WithProject(project string) Option {
	return func(c *metricsHandler) {
		c.project = project
	}
}
