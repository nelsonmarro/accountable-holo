package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/verith/internal/domain"
)

type IssuerRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewIssuerRepository(db *pgxpool.Pool) *IssuerRepositoryImpl {
	return &IssuerRepositoryImpl{db: db}
}

func (r *IssuerRepositoryImpl) GetActive(ctx context.Context) (*domain.Issuer, error) {
	// Buscamos el emisor que esté marcado como activo. Asumimos uno solo por instalación Desktop.
	var i domain.Issuer
	safeQuery := `
		SELECT id, ruc, business_name, COALESCE(trade_name, ''), main_address, establishment_address,
		       establishment_code, emission_point_code, COALESCE(contribution_class, ''), COALESCE(withholding_agent, ''),
		       COALESCE(rimpe_type, ''), environment, keep_accounting, signature_path, COALESCE(logo_path, ''), is_active, created_at, updated_at,
		       smtp_server, smtp_port, smtp_user, smtp_password, smtp_ssl, COALESCE(default_tax_rate, 4)
		FROM issuers
		WHERE is_active = TRUE
		LIMIT 1
	`

	err := r.db.QueryRow(ctx, safeQuery).Scan(
		&i.ID, &i.RUC, &i.BusinessName, &i.TradeName, &i.MainAddress, &i.EstablishmentAddress,
		&i.EstablishmentCode, &i.EmissionPointCode, &i.ContributionClass, &i.WithholdingAgent,
		&i.RimpeType, &i.Environment, &i.KeepAccounting, &i.SignaturePath, &i.LogoPath, &i.IsActive, &i.CreatedAt, &i.UpdatedAt,
		&i.SMTPServer, &i.SMTPPort, &i.SMTPUser, &i.SMTPPassword, &i.SMTPSSL, &i.DefaultTaxRate,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No hay emisor configurado, no es un error técnico
		}
		return nil, fmt.Errorf("failed to get active issuer: %w", err)
	}

	return &i, nil
}

func (r *IssuerRepositoryImpl) Create(ctx context.Context, issuer *domain.Issuer) error {
	query := `
		INSERT INTO issuers (
			ruc, business_name, trade_name, main_address, establishment_address,
			establishment_code, emission_point_code, contribution_class, withholding_agent,
			rimpe_type, environment, keep_accounting, signature_path, logo_path, is_active, created_at, updated_at,
			smtp_server, smtp_port, smtp_user, smtp_password, smtp_ssl, default_tax_rate
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		issuer.RUC, issuer.BusinessName, issuer.TradeName, issuer.MainAddress, issuer.EstablishmentAddress,
		issuer.EstablishmentCode, issuer.EmissionPointCode, issuer.ContributionClass, issuer.WithholdingAgent,
		issuer.RimpeType, issuer.Environment, issuer.KeepAccounting, issuer.SignaturePath, issuer.LogoPath, issuer.IsActive, now, now,
		issuer.SMTPServer, issuer.SMTPPort, issuer.SMTPUser, issuer.SMTPPassword, issuer.SMTPSSL, issuer.DefaultTaxRate,
	).Scan(&issuer.ID, &issuer.CreatedAt, &issuer.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create issuer: %w", err)
	}
	return nil
}

func (r *IssuerRepositoryImpl) Update(ctx context.Context, issuer *domain.Issuer) error {
	query := `
		UPDATE issuers SET
			ruc=$1, business_name=$2, trade_name=$3, main_address=$4, establishment_address=$5,
			establishment_code=$6, emission_point_code=$7, contribution_class=$8, withholding_agent=$9,
			rimpe_type=$10, environment=$11, signature_path=$12, logo_path=$13, updated_at=$14,
			smtp_server=$15, smtp_port=$16, smtp_user=$17, smtp_password=$18, smtp_ssl=$19, default_tax_rate=$20
		WHERE id=$21
	`
	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		issuer.RUC, issuer.BusinessName, issuer.TradeName, issuer.MainAddress, issuer.EstablishmentAddress,
		issuer.EstablishmentCode, issuer.EmissionPointCode, issuer.ContributionClass, issuer.WithholdingAgent,
		issuer.RimpeType, issuer.Environment, issuer.SignaturePath, issuer.LogoPath, now,
		issuer.SMTPServer, issuer.SMTPPort, issuer.SMTPUser, issuer.SMTPPassword, issuer.SMTPSSL, issuer.DefaultTaxRate,
		issuer.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update issuer: %w", err)
	}
	issuer.UpdatedAt = now
	return nil
}
