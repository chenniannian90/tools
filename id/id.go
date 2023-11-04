package id

import (
	"github.com/go-courier/snowflakeid"
	"github.com/go-courier/snowflakeid/workeridutil"
	"github.com/pkg/errors"
)

type Generator interface {
	ID() (uint64, error)
}

type GeneratorImpl = snowflakeid.Snowflake

func NewGeneratorImpl(opts ...Option) (*GeneratorImpl, error) {
	p := &params{
		bitLenWorkerID: defaultBitLenWorkerID,
		bitLenSequence: defaultBitLenSequence,
		gapMs:          defaultGapMs,
		startTime:      defaultStartTime,
	}

	for _, opt := range opts {
		opt(p)
	}

	sff := snowflakeid.NewSnowflakeFactory(
		p.bitLenWorkerID,
		p.bitLenSequence,
		p.gapMs,
		p.startTime,
	)

	workerID := workeridutil.WorkerIDFromIP(ResolveExposedIP())

	return sff.NewSnowflake(workerID)
}

func MustNewGeneratorImpl(opts ...Option) *GeneratorImpl {
	generatorImpl, err := NewGeneratorImpl(opts...)
	if err != nil {
		panic(errors.Wrap(err, "failed to new generatorImpl"))
	}

	return generatorImpl
}
