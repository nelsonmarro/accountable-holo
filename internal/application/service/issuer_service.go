package service

import (
	"context"
	"fmt"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/zalando/go-keyring"
)

type IssuerService struct {
	repo   IssuerRepository
	epRepo EmissionPointRepository
}

func NewIssuerService(repo IssuerRepository, epRepo EmissionPointRepository) *IssuerService {
	return &IssuerService{repo: repo, epRepo: epRepo}
}

func (s *IssuerService) GetActive(ctx context.Context) (*domain.Issuer, error) {
	return s.repo.GetActive(ctx)
}

func (s *IssuerService) GetEmissionPoints(ctx context.Context) ([]domain.EmissionPoint, error) {
	issuer, err := s.repo.GetActive(ctx)
	if err != nil || issuer == nil {
		return nil, fmt.Errorf("no hay emisor activo")
	}
	return s.epRepo.GetAllByIssuer(ctx, issuer.ID)
}

func (s *IssuerService) UpdateEmissionPoint(ctx context.Context, ep *domain.EmissionPoint) error {
	return s.epRepo.Update(ctx, ep)
}

// SaveIssuerConfig guarda la configuración del emisor en la DB y la contraseña en el Keyring.
func (s *IssuerService) SaveIssuerConfig(ctx context.Context, issuer *domain.Issuer, password string) error {
	// 1. Guardar/Actualizar en DB
	existing, err := s.repo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("error verificando emisor existente: %w", err)
	}

	if existing != nil {
		issuer.ID = existing.ID
		// Mantener fecha de creación
		issuer.CreatedAt = existing.CreatedAt
		if err := s.repo.Update(ctx, issuer); err != nil {
			return fmt.Errorf("error actualizando emisor: %w", err)
		}
	} else {
		if err := s.repo.Create(ctx, issuer); err != nil {
			return fmt.Errorf("error creando emisor: %w", err)
		}
	}

	// 1.5 Pre-inicializar Puntos de Emisión si no existen
	// Esto permite que el usuario pueda migrar secuenciales inmediatamente después de guardar
	receiptTypes := []string{"01", "04"} // Factura y Nota de Crédito
	for _, rt := range receiptTypes {
		ep, err := s.epRepo.GetByPoint(ctx, issuer.ID, issuer.EstablishmentCode, issuer.EmissionPointCode, rt)
		if err == nil && ep == nil {
			newEP := &domain.EmissionPoint{
				IssuerID:          issuer.ID,
				EstablishmentCode: issuer.EstablishmentCode,
				EmissionPointCode: issuer.EmissionPointCode,
				ReceiptType:       rt,
				CurrentSequence:   0,
				IsActive:          true,
			}
			_ = s.epRepo.Create(ctx, newEP)
		}
	}

	// 2. Guardar contraseña en Keyring del SO de forma segura
	// Service: "Verith", User: RUC
	// Solo actualizamos si se proporcionó una nueva contraseña
	if password != "" {
		err := keyring.Set("Verith", issuer.RUC, password)
		if err != nil {
			return fmt.Errorf("error guardando contraseña en llavero seguro: %w", err)
		}
	}

	return nil
}

func (s *IssuerService) GetIssuerConfig(ctx context.Context) (*domain.Issuer, error) {
	return s.repo.GetActive(ctx)
}

// GetSignaturePassword recupera la contraseña del certificado desde el llavero seguro.
func (s *IssuerService) GetSignaturePassword(ruc string) (string, error) {
	return keyring.Get("Verith", ruc)
}
