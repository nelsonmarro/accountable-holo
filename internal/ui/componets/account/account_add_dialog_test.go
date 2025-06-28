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
	testCallback            func(bool, ...*sync.WaitGroup) (func(), bool)
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
		testCallback: func(b bool, wg ...*sync.WaitGroup) (func(), bool) {
			return func() {
				b = true // Simulate callback being fired
				if len(wg) > 0 {
					wg[0].Done() // Signal completion
				}
			}, b
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
					if len(wg) > 0 {
						wg[0].Done()
					}
				})
		},
		testCallback: func(b bool, wg ...*sync.WaitGroup) (func(), bool) {
			return func() {
				b = true // Simulate callback being fired
			}, b
	},
	{
		name:                  "should not do anything if form is invalid",
		expectedCallbackFired: false,
		handleSubmitSuccess:   false,
		waitTimeoutDuration:   1 * time.Second,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
		},
		testCallback: func(b bool, wg ...*sync.WaitGroup) func() {
			return func() {
				b = true // Simulate callback being fired
			}
		},
	},
}

func TestHandleSubmit(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var wg sync.WaitGroup
			wg.Add(1)

			callbackFired := false

			d, mockService := setupTest(tc.testCallback(callbackFired, &wg))

			tc.mockServiceExpectations(mockService, &wg)

			// Act
			d.handleSubmit(tc.handleSubmitSuccess)

			// Assert
			waitTimeout(t, &wg, tc.waitTimeoutDuration)

			mockService.AssertExpectations(t)
			assert.Equal(t, tc.expectedCallbackFired, callbackFired)
		})
	}
}
