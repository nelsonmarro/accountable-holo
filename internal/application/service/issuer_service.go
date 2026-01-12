package service

import (
	"context"
	"fmt"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/zalando/go-keyring"
)

type IssuerService struct {
	repo IssuerRepository
}

func NewIssuerService(repo IssuerRepository) *IssuerService {
	return &IssuerService{repo: repo}
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

	// 2. Guardar contraseña en Keyring del SO de forma segura
	// Service: "AccountableHolo", User: RUC
	// Solo actualizamos si se proporcionó una nueva contraseña
	if password != "" {
		err := keyring.Set("AccountableHolo", issuer.RUC, password)
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
	return keyring.Get("AccountableHolo", ruc)
}
