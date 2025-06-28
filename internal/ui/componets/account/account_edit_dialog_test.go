package account

import (
	"sync"
	"testing"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
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
			wg.Done()
		}

		onFailureFires := new(bool)
		*onFailureFires = false

		onFailureCallback := func(err error) {
			*onFailureFires = true
			wg.Done()
		}

		_, _, mockService := setupDependencies()
		// Act
	})
}
