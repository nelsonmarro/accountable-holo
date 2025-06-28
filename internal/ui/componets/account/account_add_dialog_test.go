package account

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleSubmit(t *testing.T) {
	t.Run("should call service and trigger callback on success", func(t *testing.T) {
		// Arrange
		var wg sync.WaitGroup
		wg.Add(1)

		callbackFired := false
		testCallback := func() {
			callbackFired = true
			wg.Done() // Signal completion
		}

		d, mockService := setupTest(testCallback)

		// The mock service now just needs to return success.
		mockService.On("CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)

		// Act
		d.handleSubmit(true)

		// Assert
		waitTimeout(t, &wg, 2*time.Second)

		mockService.AssertExpectations(t)
		assert.True(t, callbackFired, "Expected callbackAction to be fired on success")
	})
}
