package account

import (
	"sync"
	"testing"
)

func TestHandleSubmit(t *testing.T) {
	t.Run("should call service and trigger callback on success", func(t *testing.T) {
		// Arrange
		d, mockService, callbackFired := setupTest()
		var wg sync.WaitGroup
	})
}
