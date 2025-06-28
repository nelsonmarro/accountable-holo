package account

import (
	"sync"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchAccount(t *testing.T) {
	t.Run("should fire onSuccess callback if service works", func(t *testing.T) {
		// Arrange
		var wg sync.WaitGroup
		wg.Add(1)

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

		mockService.On("GetAccountByID", mock.Anything, d.accountID).
			Return(&domain.Account{}, nil).
			Run(func(args mock.Arguments) {
				wg.Done()
			})

		// Act
		d.fetchAccount(onSuccessCallback, onFailureCallback)

		// Assert
		waitTimeout(t, &wg, 1*time.Second)

		mockService.AssertExpectations(t)
		assert.True(t, *onSuccessFires)
		assert.False(t, *onFailureFires)
	})
}
