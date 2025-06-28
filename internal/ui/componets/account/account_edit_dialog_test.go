package account

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var editFetchTestCases = []struct {
	name                    string
	wg                      *sync.WaitGroup
	ExpectedSuccessFires    bool
	ExpectedFailureFires    bool
	mockServiceExpectations func(*mocks.MockAccountService, *sync.WaitGroup)
}{
	{
		name:                 "should fire onSuccess callback if service works",
		wg:                   &sync.WaitGroup{},
		ExpectedSuccessFires: true,
		ExpectedFailureFires: false,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg *sync.WaitGroup) {
			mockService.On("GetAccountByID", mock.Anything, 1).
				Return(&domain.Account{}, nil).
				Run(func(args mock.Arguments) {
					wg.Done()
				})
		},
	},
	{
		name:                 "should fire onFailure callback if service returns an error",
		wg:                   &sync.WaitGroup{},
		ExpectedSuccessFires: false,
		ExpectedFailureFires: true,
		mockServiceExpectations: func(mockService *mocks.MockAccountService, wg *sync.WaitGroup) {
			mockService.On("GetAccountByID", mock.Anything, 1).
				Return(&domain.Account{}, errors.New("sql error")).
				Run(func(args mock.Arguments) {
					wg.Done()
				})
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
			}

			onFailureFires := new(bool)
			*onFailureFires = false
			onFailureCallback := func(err error) {
				*onFailureFires = true
			}

			editCallback := func() {}

			d, mockService := setupTestEditDialog(editCallback)

			tc.mockServiceExpectations(mockService, tc.wg)

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
