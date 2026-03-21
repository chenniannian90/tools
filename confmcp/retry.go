package confmcp

import (
	"time"

	"github.com/go-courier/envconf"
	"github.com/sirupsen/logrus"
)

// Retry configuration for MCP operations
type Retry struct {
	Repeats  int
	Interval envconf.Duration
}

// SetDefaults sets default retry values
func (r *Retry) SetDefaults() {
	if r.Repeats == 0 {
		r.Repeats = 3
	}
	if r.Interval == 0 {
		r.Interval = envconf.Duration(10 * time.Second)
	}
}

// Do executes function with retry logic
func (r Retry) Do(exec func() error) (err error) {
	if r.Repeats <= 0 {
		err = exec()
		return
	}

	for i := 0; i < r.Repeats; i++ {
		err = exec()
		if err != nil {
			if i < r.Repeats-1 {
				logrus.Warningf("MCP operation failed, retrying in %d seconds [%d/%d]",
					r.Interval, i+1, r.Repeats)
				time.Sleep(time.Duration(r.Interval))
				continue
			}
		}
		break
	}
	return
}
