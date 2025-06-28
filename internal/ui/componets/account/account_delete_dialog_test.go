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

var delTestCases = []struct {
	name                    string
	callbackFired           bool
	waitTimeoutDuration     time.Duration
	wg                      *sync.WaitGroup
	tasksToWaint            int
	mockServiceExpectations func(*mocks.MockAccountService, ...*sync.WaitGroup)
	testCallback            func(*bool, ...*sync.WaitGroup) (func(), *bool)
}{
	{
		name:                "should call service and trigger callback on success",
		callbackFired:       true,
		waitTimeoutDuration: 2 * time.Second,
		wg:                  &sync.WaitGroup{},
		tasksToWaint:        1,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
			mockService.On("CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
				Return(nil)
		},
		testCallback: func(b *bool, wg ...*sync.WaitGroup) (func(), *bool) {
			return func() {
				*b = true // Simulate callback being fired
				if len(wg) > 0 {
					wg[0].Done() // Signal completion
				}
			}, b
		},
	},
	{
		name:                "should not trigger callback when service returns an error",
		callbackFired:       false,
		waitTimeoutDuration: 1 * time.Second,
		wg:                  &sync.WaitGroup{},
		tasksToWaint:        1,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
			mockService.On("CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
				Return(errors.New("sql error")).
				Run(func(args mock.Arguments) {
					if len(wg) > 0 {
						wg[0].Done()
					}
				})
		},
		testCallback: func(b *bool, wg ...*sync.WaitGroup) (func(), *bool) {
			return func() {
				*b = true // Simulate callback being fired
			}, b
		},
	},
	{
		name:                "should not do anything if form is invalid",
		callbackFired:       false,
		waitTimeoutDuration: 1 * time.Second,
		wg:                  &sync.WaitGroup{},
		tasksToWaint:        0,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
		},
		testCallback: func(b *bool, wg ...*sync.WaitGroup) (func(), *bool) {
			return func() {
				*b = true // Simulate callback being fired
			}, b
		},
	},
}

func ExecuteDelete(t *testing.T) {
	for _, tc := range delTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			if tc.tasksToWaint > 0 {
				tc.wg.Add(tc.tasksToWaint)
			}

			callbackFired := new(bool)
			*callbackFired = false
			callbackFunc, callbackFired := tc.testCallback(callbackFired, tc.wg)

			d, mockService := setupTestAddDialog(callbackFunc)

			tc.mockServiceExpectations(mockService, tc.wg)

			// Act
			d.handleSubmit()

			// Assert
			if tc.tasksToWaint > 0 {
				waitTimeout(t, tc.wg, tc.waitTimeoutDuration)
			}

			mockService.AssertExpectations(t)
			assert.Equal(t, tc.callbackFired, *callbackFired)
			if !tc.handleSubmitSuccess {
				mockService.AssertNotCalled(t, "CreateNewAccount", mock.Anything, mock.AnythingOfType("*domain.Account"))
			}
		})
	}
}

