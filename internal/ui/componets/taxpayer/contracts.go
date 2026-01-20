package taxpayer

import (
	"context"

	"github.com/nelsonmarro/verith/internal/domain"
)

// TaxPayerService defines the interface for taxpayer operations required by UI components.
type TaxPayerService interface {
	Create(ctx context.Context, tp *domain.TaxPayer) error
	Update(ctx context.Context, tp *domain.TaxPayer) error
	GetPaginated(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResult[domain.TaxPayer], error)
	GetByIdentification(ctx context.Context, identification string) (*domain.TaxPayer, error)
}
