package account

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleSubmit(t *testing.T) {
	t.Run("should call service and trigger callback on success", func(t *testing.T) {
		// Arrange
		d, mockService, callbackFired := setupTest()
		var wg sync.WaitGroup
		wg.Add(1)

		// We expect the CreateNewAccount method to be called.
		// We use mock.Anything for the context and mock.AnythingOfType for the account struct.
		// When it's called, we tell it to return 'nil' (no error).
		// The .Run() function is a powerful hook that executes when the mock is called.
		// We use it to signal that our goroutine has finished.
		mockService.On(
			"CreateNewAccount",
			mock.Anything,
			mock.AnythingOfType("*domain.Account")).Return(nil).Run(func(args mock.Arguments) {
			defer wg.Done()
		})

		// Act
		// We call the handler method directly, simulating a form submission.
		d.handleSubmit(true)

		// Assert
		// Wait for the goroutine to finish, with a 1-second timeout.
		waitTimeout(t, &wg, 1)

		// Assert that the mock was called as we expected.
		mockService.AssertExpectations(t)

		// Assert that our success callback was fired.
		assert.True(t, callbackFired, "Expected callbackAction to be fired on success")
	})
}
