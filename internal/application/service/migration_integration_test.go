//go:build integration

package service_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/nelsonmarro/verith/internal/application/service"
	"github.com/nelsonmarro/verith/internal/application/service/mocks"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/infrastructure/persistence"
	"github.com/nelsonmarro/verith/internal/sri"
	"github.com/nelsonmarro/go_ec_sri_invoice_signer/pkg/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Este test usa la DB real (vía dbPool de persistencia) pero mocks para servicios externos (SRI, Mail)
func TestMigration_SequenceContinuity_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	
	// 1. Repositorios Reales
	db := persistence.GetTestPool(t) // Helper que asumo existe o similar
	txRepo := persistence.NewTransactionRepository(db)
	issuerRepo := persistence.NewIssuerRepository(db)
	receiptRepo := persistence.NewElectronicReceiptRepository(db)
	clientRepo := persistence.NewTaxPayerRepository(db)
	epRepo := persistence.NewEmissionPointRepository(db)
	categoryRepo := persistence.NewCategoryRepository(db)

	// 2. Mocks de Infraestructura
	mockSriClient := new(mocks.MockSRIClient)
	mockMail := new(mocks.MockMailService)
	logger := log.New(os.Stdout, "[MIGRATION-TEST] ", log.LstdFlags)

	// 3. Servicio Real
	svc := service.NewSriService(txRepo, issuerRepo, receiptRepo, clientRepo, epRepo, mockSriClient, mockMail, logger)
	
	// Mock Signer para no necesitar archivo .p12 real
	mockSigner := new(MockDocumentSigner)
	svc.SetSignerFactory(func(path, password string) service.DocumentSigner {
		return mockSigner
	})
	// XML válido para que unmarshal no falle
	validXml := []byte(`<factura><infoTributaria></infoTributaria><infoFactura></infoFactura><detalles></detalles></factura>`)
	mockSigner.On("Sign", mock.Anything, mock.Anything).Return(validXml, nil)
	mockSigner.On("SignCreditNote", mock.Anything, mock.Anything).Return(validXml, nil)

	// --- SETUP DATA ---
	// Crear Emisor
	issuer := &domain.Issuer{
		RUC: "1790012345001", BusinessName: "Migración Test S.A.",
		EstablishmentCode: "001", EmissionPointCode: "001", Environment: 1,
		SignaturePath: "dummy.p12", IsActive: true,
	}
	err := issuerRepo.Create(ctx, issuer)
	require.NoError(t, err)

	// Crear Cliente
	client := &domain.TaxPayer{Identification: "1712345678", Name: "Cliente Migrado", Email: "c@test.com", IdentificationType: "05"}
	err = clientRepo.Create(ctx, client)
	require.NoError(t, err)

	// Crear Categoría de Ingreso
	cat := &domain.Category{Name: "Ventas Migracion", Type: domain.Income}
	err = categoryRepo.Create(ctx, cat)
	require.NoError(t, err)

	// Crear Cuenta
	accRepo := persistence.NewAccountRepository(db)
	acc := &domain.Account{Name: "Caja Principal", InitialBalance: 1000}
	err = accRepo.Create(ctx, acc)
	require.NoError(t, err)

	// --- ESCENARIO 1: MIGRACIÓN DE FACTURAS ---
	t.Run("Migrate Invoice Sequence 1500 -> Next 1501", func(t *testing.T) {
		// Simular lo que hace el diálogo de migración:
		// Actualizar el punto de emisión de facturas a 1500
		ep, err := epRepo.GetByPoint(ctx, issuer.ID, "001", "001", "01")
		if ep == nil {
			// Si no existe (no se pre-inicializó en este test), lo creamos
			ep = &domain.EmissionPoint{IssuerID: issuer.ID, EstablishmentCode: "001", EmissionPointCode: "001", ReceiptType: "01"}
			epRepo.Create(ctx, ep)
		}
		ep.CurrentSequence = 1500
		ep.InitialSequence = 1501
		err = epRepo.Update(ctx, ep)
		require.NoError(t, err)

		// Crear Transacción para facturar
		tx := &domain.Transaction{
			AccountID: acc.ID, CategoryID: cat.ID, TaxPayerID: &client.ID,
			Amount: 100, Description: "Venta Post-Migración", TransactionDate: time.Now(),
		}
		err = txRepo.Create(ctx, tx)
		require.NoError(t, err)

		// Mock SRI Response
		mockSriClient.On("EnviarComprobante", mock.Anything, 1).Return(&sri.RespuestaRecepcion{Estado: "RECIBIDA"}, nil).Once()
		mockSriClient.On("AutorizarComprobante", mock.Anything, 1).Return(&sri.RespuestaAutorizacion{}, nil).Once()

		// EJECUTAR EMISIÓN
		err = svc.EmitirFactura(ctx, tx.ID, "pass")
		assert.NoError(t, err)

		// VERIFICAR: ¿Qué clave de acceso se generó?
		// El secuencial está en la posición 30-39 de la clave de acceso
		receipt, err := receiptRepo.GetByTransactionID(ctx, tx.ID)
		require.NoError(t, err)
		require.NotNil(t, receipt)
		
		secuencialEnClave := receipt.AccessKey[30:39]
		assert.Equal(t, "000001501", secuencialEnClave, "La factura debería ser la 1501")
		
		// Verificar en DB que el contador subió
		epPost, _ := epRepo.GetByPoint(ctx, issuer.ID, "001", "001", "01")
		assert.Equal(t, 1501, epPost.CurrentSequence)
	})

	// --- ESCENARIO 2: MIGRACIÓN DE NOTAS DE CRÉDITO ---
	t.Run("Migrate NC Sequence 50 -> Next 51", func(t *testing.T) {
		// Ajustar punto de emisión de NC (04) a 50
		ep, err := epRepo.GetByPoint(ctx, issuer.ID, "001", "001", "04")
		if ep == nil {
			ep = &domain.EmissionPoint{IssuerID: issuer.ID, EstablishmentCode: "001", EmissionPointCode: "001", ReceiptType: "04"}
			epRepo.Create(ctx, ep)
		}
		ep.CurrentSequence = 50
		err = epRepo.Update(ctx, ep)
		require.NoError(t, err)

		// Necesitamos una factura original AUTORIZADA para anular
		originalTx := &domain.Transaction{AccountID: acc.ID, CategoryID: cat.ID, TaxPayerID: &client.ID, Amount: 100, TransactionDate: time.Now()}
		txRepo.Create(ctx, originalTx)
		origReceipt := &domain.ElectronicReceipt{
			TransactionID: originalTx.ID, AccessKey: "1234567890123456789012345678901234567890123456789", 
			SRIStatus: "AUTORIZADO", ReceiptType: "01",
		}
		receiptRepo.Create(ctx, origReceipt)

		// Transacción de anulación
		voidTx := &domain.Transaction{AccountID: acc.ID, CategoryID: cat.ID, Amount: -100, Description: "Anulación"}
		txRepo.Create(ctx, voidTx)

		// Mock SRI
		mockSriClient.On("EnviarComprobante", mock.Anything, 1).Return(&sri.RespuestaRecepcion{Estado: "RECIBIDA"}, nil).Once()
		mockSriClient.On("AutorizarComprobante", mock.Anything, 1).Return(&sri.RespuestaAutorizacion{}, nil).Once()

		// EJECUTAR ANULACIÓN (NC)
		claveNC, err := svc.EmitirNotaCredito(ctx, voidTx.ID, originalTx.ID, "Error", "pass")
		assert.NoError(t, err)
		
		// VERIFICAR SECUENCIAL EN NC
		secuencialNC := claveNC[30:39]
		assert.Equal(t, "000000051", secuencialNC, "La Nota de Crédito debería ser la 51")
		
		// Verificar en DB
		epPost, _ := epRepo.GetByPoint(ctx, issuer.ID, "001", "001", "04")
		assert.Equal(t, 51, epPost.CurrentSequence)
	})
}
