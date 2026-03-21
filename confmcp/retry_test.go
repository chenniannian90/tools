package confmcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-courier/envconf"
)

func TestRetrySetDefaults(t *testing.T) {
	retry := &Retry{}
	retry.SetDefaults()

	if retry.Repeats != 3 {
		t.Errorf("expected 3 repeats, got %d", retry.Repeats)
	}

	expectedInterval := 10 * time.Second
	if time.Duration(retry.Interval) != expectedInterval {
		t.Errorf("expected interval %v, got %v", expectedInterval, retry.Interval)
	}
}

func TestRetryDoWithSuccess(t *testing.T) {
	retry := &Retry{
		Repeats:  3,
		Interval: envconf.Duration(10 * time.Millisecond),
	}

	executed := false
	err := retry.Do(func() error {
		executed = true
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if !executed {
		t.Error("expected function to be executed")
	}
}

func TestRetryDoWithFailure(t *testing.T) {
	retry := &Retry{
		Repeats:  3,
		Interval: envconf.Duration(10 * time.Millisecond),
	}

	expectedError := errors.New("test error")
	err := retry.Do(func() error {
		return expectedError
	})

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestRetryDoWithEventualSuccess(t *testing.T) {
	retry := &Retry{
		Repeats:  3,
		Interval: envconf.Duration(10 * time.Millisecond),
	}

	attempts := 0
	err := retry.Do(func() error {
		attempts++
		if attempts < 3 {
			return errors.New("not yet")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryDoWithNoRepeats(t *testing.T) {
	retry := &Retry{
		Repeats: 0,
	}

	executed := false
	err := retry.Do(func() error {
		executed = true
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if !executed {
		t.Error("expected function to be executed")
	}
}

func TestRetryDoWithNoRepeatsFailure(t *testing.T) {
	retry := &Retry{
		Repeats: 0,
	}

	expectedError := errors.New("test error")
	err := retry.Do(func() error {
		return expectedError
	})

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestRetryContext(t *testing.T) {
	retry := &Retry{
		Repeats:  5,
		Interval: envconf.Duration(100 * time.Millisecond),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	attempts := 0
	err := retry.Do(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			attempts++
			return errors.New("keep trying")
		}
	})

	if err == nil {
		t.Error("expected error from context timeout")
	}

	// Should have made a few attempts before timeout
	if attempts == 0 {
		t.Error("expected at least one attempt")
	}

	if attempts > 4 {
		t.Errorf("expected timeout to stop retries, got %d attempts", attempts)
	}
}

func TestRetryZeroInterval(t *testing.T) {
	retry := &Retry{
		Repeats:  3,
		Interval: envconf.Duration(0),
	}
	retry.SetDefaults() // Should set default interval

	attempts := 0
	err := retry.Do(func() error {
		attempts++
		if attempts < 2 {
			return errors.New("not yet")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
