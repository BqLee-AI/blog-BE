package logger

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestInitLoggerKeepsExistingLoggerOnBuildFailure(t *testing.T) {
	original := L()
	originalBuilder := buildLogger
	t.Cleanup(func() {
		Set(original)
		buildLogger = originalBuilder
	})

	existing := zap.NewNop()
	Set(existing)
	buildLogger = func(config zap.Config) (*zap.Logger, error) {
		return nil, errors.New("build failed")
	}

	InitLogger("development")

	if got := L(); got != existing {
		t.Fatal("expected existing logger to be preserved when init fails")
	}
}

func TestInitLoggerReplacesLoggerOnBuildSuccess(t *testing.T) {
	original := L()
	originalBuilder := buildLogger
	t.Cleanup(func() {
		Set(original)
		buildLogger = originalBuilder
	})

	existing := zap.NewNop()
	replacement := zap.NewNop()
	Set(existing)
	buildLogger = func(config zap.Config) (*zap.Logger, error) {
		return replacement, nil
	}

	InitLogger("development")

	if got := L(); got != replacement {
		t.Fatal("expected logger to be replaced when init succeeds")
	}

	Sync()
	if got := L(); got != replacement {
		t.Fatal("expected logger to remain available after sync")
	}
}
