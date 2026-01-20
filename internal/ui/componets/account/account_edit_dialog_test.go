package account

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var editFetchTestCases = []struct {
	name                    string
	wg                      *sync.WaitGroup
	ExpectedSuccessFires    bool
	ExpectedFailureFires    bool
	mockServiceExpectations func(*mocks.MockAccountService)
}{
	{
		name:                 "should fire onSuccess callback if service works",
		wg:                   &sync.WaitGroup{},
		ExpectedSuccessFires: true,
		ExpectedFailureFires: false,
		mockServiceExpectations: func(mockService *mocks.MockAccountService) {
			mockService.On("GetAccountByID", mock.Anything, 1).
				Return(&domain.Account{}, nil)
		},
	},
	{
		name:                 "should fire onFailure callback if service returns an error",
		wg:                   &sync.WaitGroup{},
		ExpectedSuccessFires: false,
		ExpectedFailureFires: true,
		mockServiceExpectations: func(mockService *mocks.MockAccountService) {
			mockService.On("GetAccountByID", mock.Anything, 1).
				Return(&domain.Account{}, errors.New("sql error"))
		},
	},
}

func TestFetchAccount(t *testing.T) {
	for _, tc := range editFetchTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			tc.wg.Add(1)

			onSuccessFires := new(bool)
			*onSuccessFires = false
			onSuccessCallback := func(account *domain.Account) {
				*onSuccessFires = true
				tc.wg.Done()
			}

			onFailureFires := new(bool)
			*onFailureFires = false
			onFailureCallback := func(err error) {
				*onFailureFires = true
				tc.wg.Done()
			}

			editCallback := func() {}

			d, mockService := setupTestEditDialog(editCallback)

			tc.mockServiceExpectations(mockService)

			// Act
			d.fetchAccount(onSuccessCallback, onFailureCallback)

			// Assert
			waitTimeout(t, tc.wg, 1*time.Second)

			mockService.AssertExpectations(t)
			assert.Equal(t, tc.ExpectedSuccessFires, *onSuccessFires, "Expected onSuccess callback to be fired")
			assert.Equal(t, tc.ExpectedFailureFires, *onFailureFires, "Expected onFailure callback to be fired")
		})
	}
}

var editSubmitTestCases = []struct {
	name                    string
	callbackFired           bool
	handleSubmitSuccess     bool
	waitTimeoutDuration     time.Duration
	wg                      *sync.WaitGroup
	tasksToWaint            int
	mockServiceExpectations func(*mocks.MockAccountService, ...*sync.WaitGroup)
	testCallback            func(*bool, ...*sync.WaitGroup) (func(), *bool)
}{
	{
		name:                "should call service and trigger callback on success",
		callbackFired:       true,
		handleSubmitSuccess: true,
		waitTimeoutDuration: 2 * time.Second,
		wg:                  &sync.WaitGroup{},
		tasksToWaint:        1,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
			mockService.On("UpdateAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
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
		handleSubmitSuccess: true,
		waitTimeoutDuration: 1 * time.Second,
		wg:                  &sync.WaitGroup{},
		tasksToWaint:        1,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg ...*sync.WaitGroup) {
			mockService.On("UpdateAccount", mock.Anything, mock.AnythingOfType("*domain.Account")).
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
		handleSubmitSuccess: false,
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

func TestEditHandleSumbit(t *testing.T) {
	for _, tc := range editSubmitTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			if tc.tasksToWaint > 0 {
				tc.wg.Add(tc.tasksToWaint)
			}

			callbackFired := new(bool)
			*callbackFired = false
			callbackFunc, callbackFired := tc.testCallback(callbackFired, tc.wg)

			d, mockService := setupTestEditDialog(callbackFunc)

			tc.mockServiceExpectations(mockService, tc.wg)

			// Act
			d.handleSubmit(tc.handleSubmitSuccess)

			// Assert
			if tc.tasksToWaint > 0 {
				waitTimeout(t, tc.wg, tc.waitTimeoutDuration)
			}

			mockService.AssertExpectations(t)
			assert.Equal(t, tc.callbackFired, *callbackFired)
			if !tc.handleSubmitSuccess {
				mockService.AssertNotCalled(t, "UpdateAccount", mock.Anything, mock.AnythingOfType("*domain.Account"))
			}
		})
	}
}
