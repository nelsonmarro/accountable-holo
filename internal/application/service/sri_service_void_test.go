package service_test

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/application/service"
	"github.com/nelsonmarro/accountable-holo/internal/application/service/mocks"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/sri"
	"github.com/nelsonmarro/go_ec_sri_invoice_signer/pkg/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDocumentSigner simula el proceso de firma digital
type MockDocumentSigner struct {
	mock.Mock
}

func (m *MockDocumentSigner) Sign(xmlBytes []byte, algo signer.HashAlgorithm) ([]byte, error) {
	args := m.Called(xmlBytes, algo)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDocumentSigner) SignCreditNote(xmlBytes []byte, algo signer.HashAlgorithm) ([]byte, error) {
	args := m.Called(xmlBytes, algo)
	return args.Get(0).([]byte), args.Error(1)
}

func TestEmitirNotaCredito(t *testing.T) {
	// Setup Mocks
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockIssuerRepo := new(mocks.MockIssuerRepository)
	mockReceiptRepo := new(mocks.MockElectronicReceiptRepository)
	mockClientRepo := new(mocks.MockTaxPayerRepository)
	mockEpRepo := new(mocks.MockEmissionPointRepository)
	mockSriClient := new(mocks.MockSriClient)
	mockMailService := new(mocks.MockMailService)
	logger := log.New(io.Discard, "", 0)

	// Service Under Test
	sriService := service.NewSriService(
		mockTxRepo,
		mockIssuerRepo,
		mockReceiptRepo,
		mockClientRepo,
		mockEpRepo,
		mockSriClient,
		mockMailService,
		logger,
	)

	ctx := context.Background()
	originalTxID := 100
	voidTxID := 101
	password := "testpass"
	motivo := "Devolución"

	// Datos comunes
	now := time.Now()
	issuer := &domain.Issuer{
		RUC:               "1790000000001",
		Environment:       1,
		SignaturePath:     "/dummy/path.p12",
		EstablishmentCode: "001",
		EmissionPointCode: "001",
		MainAddress:       "Calle Principal",
	}
	issuer.ID = 1
	
	client := &domain.TaxPayer{
		Identification: "1710000000",
		Name:           "Juan Perez",
		Email:          "juan@test.com",
	}
	client.ID = 5

	// Receipt Original Autorizado
	originalReceipt := &domain.ElectronicReceipt{
		SRIStatus: "AUTORIZADO", 
		AccessKey: "1234567890123456789012345678901234567890123456789",
		TaxPayerID: 5,
	}

	// 1. Caso de Éxito Completo
	t.Run("Success: Emitir NC y Autorizar", func(t *testing.T) {
		// Mock Signer Factory Injection
		mockSigner := new(MockDocumentSigner)
		sriService.SetSignerFactory(func(path, password string) service.DocumentSigner {
			return mockSigner
		})

		// 1. Cargar Datos
		mockTxRepo.On("GetTransactionByID", ctx, voidTxID).Return(&domain.Transaction{BaseEntity: domain.BaseEntity{ID: voidTxID}}, nil).Once()
		originalTx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: originalTxID},
			TransactionDate:   now.Add(-24 * time.Hour),
			ElectronicReceipt: originalReceipt,
			TaxPayerID:        &client.ID,
			Amount:            112.00,
			Subtotal15:        100.00,
			TaxAmount:         12.00,
		}
		mockTxRepo.On("GetTransactionByID", ctx, originalTxID).Return(originalTx, nil).Once()

		mockTxRepo.On("GetItemsByTransactionID", ctx, originalTxID).Return([]domain.TransactionItem{
			{Description: "Item 1", Quantity: 1, UnitPrice: 100, Subtotal: 100, TaxRate: 4},
		}, nil).Once()

		mockIssuerRepo.On("GetActive", ctx).Return(issuer, nil).Once()
		mockClientRepo.On("GetByID", ctx, client.ID).Return(client, nil).Once()

		// 2. Generar Secuencial
		mockEpRepo.On("GetByPoint", ctx, issuer.ID, "001", "001", "04").Return(&domain.EmissionPoint{BaseEntity: domain.BaseEntity{ID: 10}, ReceiptType: "04", CurrentSequence: 5}, nil).Times(2)
		mockEpRepo.On("IncrementSequence", ctx, mock.Anything).Return(nil).Once()

		// 3. Firma (Simulada)
		signedXml := []byte("<xml>signed</xml>")
		mockSigner.On("SignCreditNote", mock.Anything, signer.SHA1).Return(signedXml, nil).Once()

		// 4. Guardar Recibo PENDIENTE
		mockReceiptRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.ElectronicReceipt) bool {
			return r.ReceiptType == "04" && r.SRIStatus == "PENDIENTE" && r.XMLContent == string(signedXml)
		})).Return(nil).Once()

		// 5. Enviar al SRI
		mockSriClient.On("EnviarComprobante", signedXml, issuer.Environment).Return(&sri.RespuestaRecepcion{Estado: "RECIBIDA"}, nil).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, mock.AnythingOfType("string"), "RECIBIDA", "Enviado a SRI", mock.Anything).Return(nil).Once()

		// 6. Autorizar (Simular espera)
		authResp := &sri.RespuestaAutorizacion{
			NumeroComprobantes: "1",
			Autorizaciones: struct {
				Autorizacion []sri.Autorizacion `xml:"autorizacion"`
			}{
				Autorizacion: []sri.Autorizacion{
					{Estado: "AUTORIZADO", FechaAutorizacion: time.Now().Format(time.RFC3339)},
				},
			},
		}
		// Match any access key generated (since it's random)
		mockSriClient.On("AutorizarComprobante", mock.AnythingOfType("string"), issuer.Environment).Return(authResp, nil).Once()

		// 7. Actualizar Estado Final
		mockReceiptRepo.On("UpdateStatus", ctx, mock.AnythingOfType("string"), "AUTORIZADO", "Procesado", mock.Anything).Return(nil).Once()
		
		// 8. Mock finalizeAndEmail requirements (Async calls)
		// GetActive is called again inside finalizeAndEmail
		mockIssuerRepo.On("GetActive", mock.Anything).Return(issuer, nil).Maybe()
		mockClientRepo.On("GetByID", mock.Anything, client.ID).Return(client, nil).Maybe()
		mockMailService.On("SendReceipt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		// Ejecución
		_, err := sriService.EmitirNotaCredito(ctx, voidTxID, originalTxID, motivo, password)

		// Esperar un poco a que finalizeAndEmail termine para evitar panics de mock inesperados
		time.Sleep(100 * time.Millisecond)

		// Verificación
		assert.NoError(t, err)
		mockTxRepo.AssertExpectations(t)
		mockSriClient.AssertExpectations(t)
		mockSigner.AssertExpectations(t)
	})

	t.Run("Fallo: Transacción original sin recibo autorizado", func(t *testing.T) {
		// Mock Original Tx (Sin recibo)
		mockTxRepo.On("GetTransactionByID", ctx, voidTxID).Return(&domain.Transaction{BaseEntity: domain.BaseEntity{ID: voidTxID}}, nil).Once()
		originalTx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: originalTxID},
			ElectronicReceipt: nil, // Fallo aquí
		}
		mockTxRepo.On("GetTransactionByID", ctx, originalTxID).Return(originalTx, nil).Once()

				_, err := sriService.EmitirNotaCredito(ctx, voidTxID, originalTxID, motivo, password)

				assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tiene una factura autorizada")
	})

	t.Run("Fallo: SRI Devuelve Comprobante", func(t *testing.T) {
		// Setup feliz hasta el envío
		mockSigner := new(MockDocumentSigner)
		sriService.SetSignerFactory(func(path, password string) service.DocumentSigner { return mockSigner })

		mockTxRepo.On("GetTransactionByID", ctx, voidTxID).Return(&domain.Transaction{BaseEntity: domain.BaseEntity{ID: voidTxID}}, nil).Once()
		originalTx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: originalTxID}, ElectronicReceipt: originalReceipt, TaxPayerID: &client.ID, Amount: 100, Subtotal15: 100,
		}
		mockTxRepo.On("GetTransactionByID", ctx, originalTxID).Return(originalTx, nil).Once()
		mockTxRepo.On("GetItemsByTransactionID", ctx, originalTxID).Return([]domain.TransactionItem{{Description: "Item", Quantity: 1, UnitPrice: 100, Subtotal: 100, TaxRate: 4}}, nil).Once()
		mockIssuerRepo.On("GetActive", ctx).Return(issuer, nil).Once()
		mockClientRepo.On("GetByID", ctx, client.ID).Return(client, nil).Once()
		mockEpRepo.On("GetByPoint", ctx, issuer.ID, "001", "001", "04").Return(&domain.EmissionPoint{BaseEntity: domain.BaseEntity{ID: 10}, ReceiptType: "04", CurrentSequence: 6}, nil).Times(2)
		mockEpRepo.On("IncrementSequence", ctx, mock.Anything).Return(nil).Once()
		mockSigner.On("SignCreditNote", mock.Anything, signer.SHA1).Return([]byte("<xml>"), nil).Once()
		
		// Create receipt PENDING
		mockReceiptRepo.On("Create", ctx, mock.Anything).Return(nil).Once()

		// SRI Devuelve
		sriResp := &sri.RespuestaRecepcion{
			Estado: "DEVUELTA",
		}
		sriResp.Comprobantes.Comprobante = []sri.ComprobanteRecepcion{
			{
				Mensajes: struct{Mensaje []sri.Mensaje `xml:"mensaje"`}{
					Mensaje: []sri.Mensaje{{Mensaje: "Error de esquema"}},
				},
			},
		}
		mockSriClient.On("EnviarComprobante", mock.Anything, issuer.Environment).Return(sriResp, nil).Once()
		
		// Update Status DEVUELTA
		mockReceiptRepo.On("UpdateStatus", ctx, mock.Anything, "DEVUELTA", "Error de esquema", mock.Anything).Return(nil).Once()

		_, err := sriService.EmitirNotaCredito(ctx, voidTxID, originalTxID, motivo, password)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SRI devolvió la NC")
	})

	t.Run("Fallo: SRI No Autoriza", func(t *testing.T) {
		// Setup feliz hasta la autorización
		mockSigner := new(MockDocumentSigner)
		sriService.SetSignerFactory(func(path, password string) service.DocumentSigner { return mockSigner })

		mockTxRepo.On("GetTransactionByID", ctx, voidTxID).Return(&domain.Transaction{BaseEntity: domain.BaseEntity{ID: voidTxID}}, nil).Once()
		originalTx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: originalTxID}, ElectronicReceipt: originalReceipt, TaxPayerID: &client.ID, Amount: 100, Subtotal15: 100,
		}
		mockTxRepo.On("GetTransactionByID", ctx, originalTxID).Return(originalTx, nil).Once()
		mockTxRepo.On("GetItemsByTransactionID", ctx, originalTxID).Return([]domain.TransactionItem{{Description: "Item", Quantity: 1, UnitPrice: 100, Subtotal: 100, TaxRate: 4}}, nil).Once()
		mockIssuerRepo.On("GetActive", ctx).Return(issuer, nil).Once()
		mockClientRepo.On("GetByID", ctx, client.ID).Return(client, nil).Once()
		mockEpRepo.On("GetByPoint", ctx, issuer.ID, "001", "001", "04").Return(&domain.EmissionPoint{BaseEntity: domain.BaseEntity{ID: 10}, ReceiptType: "04", CurrentSequence: 7}, nil).Times(2)
		mockEpRepo.On("IncrementSequence", ctx, mock.Anything).Return(nil).Once()
		mockSigner.On("SignCreditNote", mock.Anything, signer.SHA1).Return([]byte("<xml>"), nil).Once()
		mockReceiptRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
		
		// Envío OK
		mockSriClient.On("EnviarComprobante", mock.Anything, issuer.Environment).Return(&sri.RespuestaRecepcion{Estado: "RECIBIDA"}, nil).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, mock.Anything, "RECIBIDA", "Enviado a SRI", mock.Anything).Return(nil).Once()

		// Autorización FALLIDA
		authResp := &sri.RespuestaAutorizacion{
			NumeroComprobantes: "1",
			Autorizaciones: struct { Autorizacion []sri.Autorizacion `xml:"autorizacion"` }{
				Autorizacion: []sri.Autorizacion{{Estado: "NO AUTORIZADO"}},
			},
		}
		mockSriClient.On("AutorizarComprobante", mock.Anything, issuer.Environment).Return(authResp, nil).Once()
		
		// Update Status PROCESADO (pero estado interno NO AUTORIZADO)
		mockReceiptRepo.On("UpdateStatus", ctx, mock.Anything, "NO AUTORIZADO", "Procesado", mock.Anything).Return(nil).Once()

		_, err := sriService.EmitirNotaCredito(ctx, voidTxID, originalTxID, motivo, password)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "NC no autorizada")
	})
}