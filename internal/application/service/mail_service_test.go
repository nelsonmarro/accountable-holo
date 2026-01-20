package service

import (
	"testing"

	"github.com/nelsonmarro/accountable-holo/config"
	"github.com/stretchr/testify/assert"
)

// NOTE: MailService tests are simplified because we switched to Resend API.
// Comprehensive integration testing would require a mock for the Resend client.
// For now, we verify instantiation and config mapping.

func TestNewMailService(t *testing.T) {
	cfg := &config.Config{
		Email: config.Email{
			APIKey: "test-key",
			From:   "test@test.com",
		},
	}

	t.Run("Initialize with Config API Key", func(t *testing.T) {
		svc := NewMailService(cfg, "")
		assert.NotNil(t, svc)
		assert.Equal(t, cfg, svc.config)
	})

	t.Run("Initialize with Legacy Pass", func(t *testing.T) {
		svc := NewMailService(&config.Config{}, "legacy-key")
		assert.NotNil(t, svc)
	})
}