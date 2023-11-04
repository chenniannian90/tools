package id

import (
	"time"
)

const (
	defaultBitLenWorkerID = 16
	defaultBitLenSequence = 8
	defaultGapMs          = 5
)

var defaultStartTime, _ = time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")

type params struct {
	bitLenWorkerID uint
	bitLenSequence uint
	gapMs          uint
	startTime      time.Time
}

type Option func(*params)

func WithBitLenWorkerID(bitLenWorkerID uint) Option {
	return func(p *params) {
		p.bitLenWorkerID = bitLenWorkerID
	}
}

func WithBitLenSequence(bitLenSequence uint) Option {
	return func(p *params) {
		p.bitLenSequence = bitLenSequence
	}
}

func WithGapMs(gapMs uint) Option {
	return func(p *params) {
		p.gapMs = gapMs
	}
}

func WithStartTime(startTime time.Time) Option {
	return func(p *params) {
		p.startTime = startTime
	}
}
