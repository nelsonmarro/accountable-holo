package account

import (
	"sync"
	"testing"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/mock"
)

func TestFetchAccount(t *testing.T) {
	t.Run("should fire onSuccess callback if service works", func(t *testing.T) {
		// Arrange
		expectedAccount := &domain.Account{
			BaseEntity: domain.BaseEntity{
				ID: 1,
			},
			Name:           "Account 1",
			Number:         "11111",
			InitialBalance: 10,
			Type:           domain.OrdinaryAccount,
		}
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
			Return(expectedAccount).
			Run(func(args mock.Arguments) {
				wg.Done()
			})

		// Act
		d.fetchAccount(onSuccessCallback, onFailureCallback)
	})
}
