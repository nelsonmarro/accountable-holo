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
	"github.com/nelsonmarro/verith/internal/sri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSriService_MigrationSequences(t *testing.T) {
	// Setup Mocks
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockIssuerRepo := new(mocks.MockIssuerRepository)
	mockTaxPayerRepo := new(mocks.MockTaxPayerRepository)
	mockReceiptRepo := new(mocks.MockElectronicReceiptRepository)
	mockSriClient := new(mocks.MockSRIClient)
	mockMail := new(mocks.MockMailService)
	mockEmissionRepo := new(mocks.MockEmissionPointRepository)
	logger := log.New(os.Stdout, "[MIGRATION-STRESS] ", log.LstdFlags)

	svc := service.NewSriService(
		mockTxRepo, mockIssuerRepo, mockReceiptRepo, mockTaxPayerRepo, mockEmissionRepo, mockSriClient, mockMail, logger,
	)
	
	mockSigner := new(MockDocumentSigner)
	svc.SetSignerFactory(func(path, password string) service.DocumentSigner {
		return mockSigner
	})
	validXml := []byte(`<factura><infoTributaria></infoTributaria><infoFactura></infoFactura><detalles></detalles></factura>`)
	mockSigner.On("Sign", mock.Anything, mock.Anything).Return(validXml, nil)
	mockSigner.On("SignCreditNote", mock.Anything, mock.Anything).Return(validXml, nil)

	ctx := context.Background()

	// --- DATA ---
	issuer := &domain.Issuer{
		BaseEntity: domain.BaseEntity{ID: 1},
		RUC: "1790012345001", EstablishmentCode: "001", EmissionPointCode: "001", Environment: 1,
	}
	client := &domain.TaxPayer{BaseEntity: domain.BaseEntity{ID: 5}, Identification: "1712345678", IdentificationType: "05", Email: "c@t.com"}
	
	t.Run("Scenario: Migration of Invoice from 1500", func(t *testing.T) {
		txID := 200
		tx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: txID},
			Amount: 100, TransactionDate: time.Now(), TaxPayerID: &client.ID,
			Category: &domain.Category{Type: domain.Income},
		}

		// Simulamos que en la DB el secuencial es 1500 (migrado)
		epMigrated := &domain.EmissionPoint{
			BaseEntity: domain.BaseEntity{ID: 10},
			CurrentSequence: 1500, // <--- VALOR MIGRADO
		}

		// EXPECTATIONS
		mockTxRepo.On("GetTransactionByID", mock.Anything, txID).Return(tx, nil).Once()
		mockTxRepo.On("GetItemsByTransactionID", mock.Anything, txID).Return([]domain.TransactionItem{}, nil).Once()
		mockIssuerRepo.On("GetActive", mock.Anything).Return(issuer, nil)
		mockTaxPayerRepo.On("GetByID", mock.Anything, client.ID).Return(client, nil)
		
		// El servicio debe: 
		// 1. Consultar el punto
		mockEmissionRepo.On("GetByPoint", mock.Anything, issuer.ID, "001", "001", "01").Return(epMigrated, nil).Once()
		// 2. Incrementar (Simulamos que el Increment funciona)
		mockEmissionRepo.On("IncrementSequence", mock.Anything, epMigrated.ID).Return(nil).Once()
		
		// 3. Consultar de nuevo para obtener el valor actualizado (Simulamos que ahora es 1501)
		epUpdated := &domain.EmissionPoint{CurrentSequence: 1501}
		mockEmissionRepo.On("GetByPoint", mock.Anything, issuer.ID, "001", "001", "01").Return(epUpdated, nil).Once()

		// 4. Verificar Clave de Acceso
		mockReceiptRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *domain.ElectronicReceipt) bool {
			// El secuencial en la clave de acceso debe ser 1501 (000001501)
			seq := r.AccessKey[30:39]
			return seq == "000001501"
		})).Return(nil).Once()

		// Mock SRI (para que termine el flujo)
		mockSriClient.On("EnviarComprobante", mock.Anything, 1).Return(&sri.RespuestaRecepcion{Estado: "RECIBIDA"}, nil).Once()
		mockSriClient.On("AutorizarComprobante", mock.Anything, 1).Return(&sri.RespuestaAutorizacion{}, nil).Once()
		mockReceiptRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		// ACT
		err := svc.EmitirFactura(ctx, txID, "p")

		// ASSERT
		assert.NoError(t, err)
		mockEmissionRepo.AssertExpectations(t)
		mockReceiptRepo.AssertExpectations(t)
	})

	t.Run("Scenario: Migration of Credit Note from 50", func(t *testing.T) {
		originalTxID := 300
		voidTxID := 301
		
		originalTx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: originalTxID},
			TaxPayerID: &client.ID,
			ElectronicReceipt: &domain.ElectronicReceipt{SRIStatus: "AUTORIZADO", AccessKey: "1234567890123456789012345678901234567890123456789"},
		}
		voidTx := &domain.Transaction{BaseEntity: domain.BaseEntity{ID: voidTxID}}

		// Ep de Nota de Credito migrado a 50
		epMigrated := &domain.EmissionPoint{BaseEntity: domain.BaseEntity{ID: 11}, CurrentSequence: 50}

		mockTxRepo.On("GetTransactionByID", mock.Anything, voidTxID).Return(voidTx, nil).Once()
		mockTxRepo.On("GetTransactionByID", mock.Anything, originalTxID).Return(originalTx, nil).Once()
		mockTxRepo.On("GetItemsByTransactionID", mock.Anything, originalTxID).Return([]domain.TransactionItem{}, nil).Once()
		mockIssuerRepo.On("GetActive", mock.Anything).Return(issuer, nil)
		mockTaxPayerRepo.On("GetByID", mock.Anything, client.ID).Return(client, nil)

		// Secuencial de NC
		mockEmissionRepo.On("GetByPoint", mock.Anything, issuer.ID, "001", "001", "04").Return(epMigrated, nil).Once()
		mockEmissionRepo.On("IncrementSequence", mock.Anything, epMigrated.ID).Return(nil).Once()
		
		epUpdated := &domain.EmissionPoint{CurrentSequence: 51}
		mockEmissionRepo.On("GetByPoint", mock.Anything, issuer.ID, "001", "001", "04").Return(epUpdated, nil).Once()

		// Verificar que la clave de acceso de la NC tenga el secuencial 51 (000000051)
		mockReceiptRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *domain.ElectronicReceipt) bool {
			seq := r.AccessKey[30:39]
			return r.ReceiptType == "04" && seq == "000000051"
		})).Return(nil).Once()

		mockSriClient.On("EnviarComprobante", mock.Anything, 1).Return(&sri.RespuestaRecepcion{Estado: "RECIBIDA"}, nil).Once()
		mockSriClient.On("AutorizarComprobante", mock.Anything, 1).Return(&sri.RespuestaAutorizacion{}, nil).Once()
		mockReceiptRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		// ACT
		claveNC, err := svc.EmitirNotaCredito(ctx, voidTxID, originalTxID, "Error", "p")

		// ASSERT
		assert.NoError(t, err)
		assert.Contains(t, claveNC, "000000051")
		mockEmissionRepo.AssertExpectations(t)
	})
}
