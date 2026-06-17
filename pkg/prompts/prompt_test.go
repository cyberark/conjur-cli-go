package prompts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAskForPrompt_ContextCancellation verifies that when the caller's context
// is canceled (e.g. an MFA challenge is satisfied externally), AskForPrompt
// returns ("", nil) without propagating the cancellation as an error and
// without leaving the terminal in a broken state.
//
// In test environments stdout is not a TTY so huh runs in accessible mode,
// which means the form blocks on a stdin read. We cancel the context while the
// form is blocked and assert the contract: empty string, no error.
func TestAskForPrompt_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context almost immediately so the form never gets real input.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := AskForPrompt(ctx, "Enter value:", 5*time.Second)

	assert.NoError(t, err, "context cancellation should not surface as an error")
	assert.Empty(t, result, "result should be empty when context is canceled")
}

// TestAskForPrompt_ContextCancellation_AlreadyCanceled covers the case where
// the context is already canceled before AskForPrompt is called.
func TestAskForPrompt_ContextCancellation_AlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := AskForPrompt(ctx, "Enter value:", 5*time.Second)

	require.NoError(t, err, "pre-canceled context should not surface as an error")
	assert.Empty(t, result)
}
