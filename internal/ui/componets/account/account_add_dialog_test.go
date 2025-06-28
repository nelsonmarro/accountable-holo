package account

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/ui/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testCases = []struct {
	name                    string
	expectedCallbackFired   bool
	handleSubmitSuccess     bool
	waitTimeoutDuration     time.Duration
	mockServiceExpectations func(*mocks.MockAccountService, ...*sync.WaitGroup)
}{
	{
		name:                  "should call service and trigger callback on success",
		expectedCallbackFired: true,
		handleSubmitSuccess:   true,
		waitTimeoutDuration:   2 * time.Second,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
			mockService.On("CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
				Return(nil)
		},
	},
	{
		name:                  "should not trigger callback when service returns an error",
		expectedCallbackFired: false,
		handleSubmitSuccess:   true,
		waitTimeoutDuration:   1 * time.Second,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
			mockService.On("CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
				Return(errors.New("sql error")).
				Run(func(args mock.Arguments) {
					wg[0].Done()
				})
		},
	},
	{
		name:                  "should not do anything if form is invalid",
		expectedCallbackFired: false,
		handleSubmitSuccess:   false,
		waitTimeoutDuration:   1 * time.Second,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
			mockService.On("CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
				Return(nil)
		},
	},
}

func TestHandleSubmit(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
				tc.mockServiceExpectations(mockService)

				// Act
				d.handleSubmit(tc.handleSubmitSuccess)

				// Assert
				waitTimeout(t, &wg, 2*time.Second)

				mockService.AssertExpectations(t)
				assert.True(t, callbackFired, "Expected callbackAction to be fired on success")
			})
		})
	}

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

		mockService.On("CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
			Return(errors.New("sql error")).
			Run(func(args mock.Arguments) {
				wg.Done()
			})

		// Act
		d.handleSubmit(true)

		// Assert
		waitTimeout(t, &wg, 1*time.Second)

		mockService.AssertExpectations(t)
		assert.False(t, callbackFired, "Expected callbackAction NOT to be fired on service error")
	})

	t.Run("should no to anything is form is invalid", func(t *testing.T) {
		// Arrange
		callbackFired := false
		testCallback := func() {
			callbackFired = true
		}

		d, mockService := setupTest(testCallback)

		// Act
		d.handleSubmit(false)

		// Assert
		mockService.AssertNotCalled(t, "CreateNewAccount")
		assert.False(t, callbackFired)
	})
}
