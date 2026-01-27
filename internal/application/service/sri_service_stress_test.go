package service_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/nelsonmarro/verith/internal/application/service"
	"github.com/nelsonmarro/verith/internal/application/service/mocks"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/sri"
	"github.com/nelsonmarro/go_ec_sri_invoice_signer/pkg/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSriService_InvoiceFlow_Stress(t *testing.T) {
	// Helper setup function to ensure isolation
	setup := func() (*service.SriService, *mocks.MockTransactionRepository, *mocks.MockIssuerRepository, *mocks.MockTaxPayerRepository, *mocks.MockElectronicReceiptRepository, *mocks.MockEmissionPointRepository, *mocks.MockSRIClient, *mocks.MockMailService, *MockDocumentSigner) {
		mockTxRepo := new(mocks.MockTransactionRepository)
		mockIssuerRepo := new(mocks.MockIssuerRepository)
		mockTaxPayerRepo := new(mocks.MockTaxPayerRepository)
		mockReceiptRepo := new(mocks.MockElectronicReceiptRepository)
		mockSriClient := new(mocks.MockSRIClient)
		mockMail := new(mocks.MockMailService)
		// mockStorage is unused in main flow (temp files)
		mockEmissionRepo := new(mocks.MockEmissionPointRepository)
		
		// Use Stdout for debugging
		logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

		svc := service.NewSriService(
			mockTxRepo, mockIssuerRepo, mockReceiptRepo, mockTaxPayerRepo, mockEmissionRepo, mockSriClient, mockMail, logger,
		)
		
		mockSigner := new(MockDocumentSigner)
		svc.SetSignerFactory(func(path, password string) service.DocumentSigner {
			return mockSigner
		})

		return svc, mockTxRepo, mockIssuerRepo, mockTaxPayerRepo, mockReceiptRepo, mockEmissionRepo, mockSriClient, mockMail, mockSigner
	}

	ctx := context.Background()

	// --- Common Data ---
	txID := 100
	issuerID := 1
	taxPayerID := 50
	
	validIssuer := &domain.Issuer{
		BaseEntity: domain.BaseEntity{ID: issuerID},
		RUC: "1790012345001", BusinessName: "Verith Corp", 
		EstablishmentCode: "001", EmissionPointCode: "001", Environment: 1,
		SignaturePath: "dummy.p12", 
		MainAddress: "Test Addr", EstablishmentAddress: "Test Addr",
	}
	
	validTx := &domain.Transaction{
		BaseEntity: domain.BaseEntity{ID: txID},
		Amount: 115.0, TransactionDate: time.Now(),
		Subtotal15: 100.0, TaxAmount: 15.0,
		TaxPayerID: &taxPayerID,
		Category: &domain.Category{Type: domain.Income},
	}
	
	validClient := &domain.TaxPayer{
		BaseEntity: domain.BaseEntity{ID: taxPayerID},
		Identification: "1712345678", Email: "client@test.com", IdentificationType: "05",
		Name: "Test Client",
	}
	
	emissionPoint := &domain.EmissionPoint{BaseEntity: domain.BaseEntity{ID: 10}, CurrentSequence: 50}

	t.Run("Happy Path: Invoice Authorized", func(t *testing.T) {
		svc, mockTxRepo, mockIssuerRepo, mockTaxPayerRepo, mockReceiptRepo, mockEmissionRepo, mockSriClient, mockMail, mockSigner := setup()

		// 1. Data Retrieval
		mockTxRepo.On("GetTransactionByID", mock.Anything, txID).Return(validTx, nil).Once()
		mockTxRepo.On("GetItemsByTransactionID", mock.Anything, txID).Return([]domain.TransactionItem{{Description: "Item", Quantity: 1, UnitPrice: 100, Subtotal: 100, TaxRate: 4}}, nil).Once()
		
		mockIssuerRepo.On("GetActive", mock.Anything).Return(validIssuer, nil)
		mockTaxPayerRepo.On("GetByID", mock.Anything, taxPayerID).Return(validClient, nil)
		
		mockEmissionRepo.On("GetByPoint", mock.Anything, issuerID, "001", "001", "01").Return(emissionPoint, nil)
		mockEmissionRepo.On("IncrementSequence", mock.Anything, emissionPoint.ID).Return(nil).Once()

		validXml := []byte(`<factura><infoTributaria></infoTributaria><infoFactura></infoFactura><detalles></detalles></factura>`)
		mockSigner.On("Sign", mock.Anything, signer.SHA1).Return(validXml, nil).Once()

		// 2. SRI Reception
		mockSriClient.On("EnviarComprobante", mock.Anything, 1).Return(&sri.RespuestaRecepcion{
			Estado: "RECIBIDA",
			Comprobantes: struct{Comprobante []sri.ComprobanteRecepcion `xml:"comprobante"`}{
				Comprobante: []sri.ComprobanteRecepcion{{ClaveAcceso: "1234567890123456789012345678901234567890123456789"}},
			},
		}, nil).Once()

		// 3. Persistence
		mockReceiptRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockReceiptRepo.On("UpdateStatus", mock.Anything, mock.Anything, "RECIBIDA", "Recibido por SRI", mock.Anything).Return(nil).Once()

		// 4. SRI Authorization
		authResponse := &sri.RespuestaAutorizacion{
			Autorizaciones: struct{Autorizacion []sri.Autorizacion `xml:"autorizacion"`}{
				Autorizacion: []sri.Autorizacion{{
					Estado: "AUTORIZADO", 
					FechaAutorizacion: time.Now().Format(time.RFC3339),
					NumeroAutorizacion: "1234567890",
					Comprobante: "<xml>...</xml>",
				}},
			},
		}
		mockSriClient.On("AutorizarComprobante", mock.Anything, 1).Return(authResponse, nil).Once()

		// 5. Final Status
		mockReceiptRepo.On("UpdateStatus", mock.Anything, mock.Anything, "AUTORIZADO", mock.Anything, mock.Anything).Return(nil).Once()

		// 7. Email (Async)
		// Relaxed matchers for email args
		mockMail.On("SendReceipt", mock.Anything, "client@test.com", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mockReceiptRepo.On("UpdateEmailSent", mock.Anything, mock.Anything, true).Return(nil).Once()

		err := svc.EmitirFactura(ctx, txID, "password")

		time.Sleep(500 * time.Millisecond) // Wait for async

		assert.NoError(t, err)
		mockSriClient.AssertExpectations(t)
		mockMail.AssertExpectations(t)
	})

	t.Run("Resilience: SRI Network Error (Retry Later)", func(t *testing.T) {
		svc, mockTxRepo, mockIssuerRepo, mockTaxPayerRepo, mockReceiptRepo, mockEmissionRepo, mockSriClient, _, mockSigner := setup()

		mockTxRepo.On("GetTransactionByID", mock.Anything, txID).Return(validTx, nil).Once()
		mockTxRepo.On("GetItemsByTransactionID", mock.Anything, txID).Return([]domain.TransactionItem{{Description: "Item", Quantity: 1, UnitPrice: 100, Subtotal: 100, TaxRate: 4}}, nil).Once()
		
		mockIssuerRepo.On("GetActive", mock.Anything).Return(validIssuer, nil)
		mockTaxPayerRepo.On("GetByID", mock.Anything, taxPayerID).Return(validClient, nil)

		mockEmissionRepo.On("GetByPoint", mock.Anything, issuerID, "001", "001", "01").Return(emissionPoint, nil)
		mockEmissionRepo.On("IncrementSequence", mock.Anything, emissionPoint.ID).Return(nil).Once()
		
		validXml := []byte(`<factura><infoTributaria></infoTributaria><infoFactura></infoFactura><detalles></detalles></factura>`)
		mockSigner.On("Sign", mock.Anything, signer.SHA1).Return(validXml, nil).Once()

		mockReceiptRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		// SRI FAILS
		mockSriClient.On("EnviarComprobante", mock.Anything, 1).Return(nil, errors.New("timeout")).Once()
		
		mockReceiptRepo.On("UpdateStatus", mock.Anything, mock.Anything, "ERROR_RED", mock.MatchedBy(func(msg string) bool {
			return true
		}), mock.Anything).Return(nil).Once()

		err := svc.EmitirFactura(ctx, txID, "password")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
		mockReceiptRepo.AssertExpectations(t)
	})
}