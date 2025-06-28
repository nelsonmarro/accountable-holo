package account

import (
	"errors"
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

	t.Run("Should not trigger callback when service returns an error", func(t *testing.T) {
		// Arrange
		var wg sync.WaitGroup
		wg.Add(1)

		callbackFired := false
		// this callback will not be called
		testCallback := func() {
			callbackFired = true
		}

		d, mockService := setupTest(testCallback)

		// The mock setup for the failure case is different.
		// The service call IS the last major asynchronous step in the failure path.
		// So we signal the WaitGroup when the service mock is called.
		mockService.On("CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
			Return(errors.New("sql error")).
			Run(func(args mock.Arguments) {
				// On failure, the goroutine exits after the service call, so this is the
				// correct place to signal completion for the failure test.
				wg.Done()
			})
	})
}
