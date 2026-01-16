package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/application/service/mocks"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/sri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSyncReceipt(t *testing.T) {
	// Mocks
	mockReceiptRepo := new(mocks.MockElectronicReceiptRepository)
	mockSriClient := new(mocks.MockSriClient)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockIssuerRepo := new(mocks.MockIssuerRepository)
	mockClientRepo := new(mocks.MockTaxPayerRepository)
	mockMailService := new(mocks.MockMailService)
	logger := log.New(io.Discard, "", 0)

	// Service
	service := NewSriService(mockTxRepo, mockIssuerRepo, mockReceiptRepo, mockClientRepo, nil, mockSriClient, mockMailService, logger)
	ctx := context.Background()

	// Default expectations for the async finalizeAndEmail (to prevent panics)
	mockTxRepo.On("GetTransactionByID", mock.Anything, mock.Anything).Return(&domain.Transaction{ElectronicReceipt: &domain.ElectronicReceipt{}}, nil).Maybe()
	mockIssuerRepo.On("GetActive", mock.Anything).Return(&domain.Issuer{}, nil).Maybe()
	mockClientRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domain.TaxPayer{}, nil).Maybe()
	mockMailService.On("SendReceipt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	t.Run("PENDIENTE -> RECIBIDA (Envío Exitoso)", func(t *testing.T) {
		// Arrange
		receipt := &domain.ElectronicReceipt{
			AccessKey:   "1234567890",
			XMLContent:  "<xml>...</xml>",
			SRIStatus:   "PENDIENTE",
			Environment: 1,
		}

		sriResponse := &sri.RespuestaRecepcion{
			Estado: "RECIBIDA",
		}

		mockSriClient.On("EnviarComprobante", []byte(receipt.XMLContent), 1).Return(sriResponse, nil).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, receipt.AccessKey, "RECIBIDA", "Recibido por SRI", (*time.Time)(nil)).Return(nil).Once()

		// Expect immediate authorization check (since it moved to RECIBIDA)
		mockSriClient.On("AutorizarComprobante", receipt.AccessKey, 1).Return(&sri.RespuestaAutorizacion{NumeroComprobantes: "0"}, nil).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, receipt.AccessKey, "EN PROCESO", mock.Anything, (*time.Time)(nil)).Return(nil).Once()

		// Act
		status, err := service.SyncReceipt(ctx, receipt)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "EN PROCESO", status)
		mockSriClient.AssertExpectations(t)
		mockReceiptRepo.AssertExpectations(t)
	})

	t.Run("RECIBIDA -> AUTORIZADO (Autorización Exitosa)", func(t *testing.T) {
		// Arrange
		receipt := &domain.ElectronicReceipt{
			AccessKey:   "1234567890",
			SRIStatus:   "RECIBIDA",
			Environment: 1,
		}

		authDateStr := "2026-01-10T12:00:00-05:00"
		authDate, _ := time.Parse(time.RFC3339, authDateStr)

		sriResponse := &sri.RespuestaAutorizacion{
			NumeroComprobantes: "1",
			Autorizaciones: struct {
				Autorizacion []sri.Autorizacion `xml:"autorizacion"`
			}{
				Autorizacion: []sri.Autorizacion{
					{
						Estado:            "AUTORIZADO",
						FechaAutorizacion: authDateStr,
					},
				},
			},
		}

		mockSriClient.On("AutorizarComprobante", receipt.AccessKey, 1).Return(sriResponse, nil).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, receipt.AccessKey, "AUTORIZADO", "Autorización Exitosa", &authDate).Return(nil).Once()

		// Act
		status, err := service.SyncReceipt(ctx, receipt)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "AUTORIZADO", status)
	})

	t.Run("RECIBIDA -> EN PROCESO (SRI procesando)", func(t *testing.T) {
		receipt := &domain.ElectronicReceipt{AccessKey: "KEY_PENDING", SRIStatus: "RECIBIDA", Environment: 1}
		
		// Simula respuesta vacía o 0 comprobantes
		sriResponse := &sri.RespuestaAutorizacion{NumeroComprobantes: "0"}
		
		mockSriClient.On("AutorizarComprobante", receipt.AccessKey, 1).Return(sriResponse, nil).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, receipt.AccessKey, "EN PROCESO", mock.Anything, (*time.Time)(nil)).Return(nil).Once()

		status, err := service.SyncReceipt(ctx, receipt)
		require.NoError(t, err)
		assert.Equal(t, "EN PROCESO", status)
	})

	t.Run("RECIBIDA -> NO AUTORIZADO (SRI rechaza)", func(t *testing.T) {
		receipt := &domain.ElectronicReceipt{AccessKey: "KEY_REJECT", SRIStatus: "RECIBIDA", Environment: 1}
		
		sriResponse := &sri.RespuestaAutorizacion{
			NumeroComprobantes: "1",
			Autorizaciones: struct {
				Autorizacion []sri.Autorizacion `xml:"autorizacion"`
			}{
				Autorizacion: []sri.Autorizacion{
					{
						Estado: "NO AUTORIZADO",
						Mensajes: struct{ Mensaje []sri.Mensaje `xml:"mensaje"` }{
							Mensaje: []sri.Mensaje{{Identificador: "39", Mensaje: "FIRMA INVALIDA"}},
						},
					},
				},
			},
		}
		
		mockSriClient.On("AutorizarComprobante", receipt.AccessKey, 1).Return(sriResponse, nil).Once()
		// Debe guardar el mensaje de error concatenado
		mockReceiptRepo.On("UpdateStatus", ctx, receipt.AccessKey, "RECHAZADA", mock.MatchedBy(func(msg string) bool {
			return len(msg) > 0 // El mensaje debe contener el error
		}), mock.Anything).Return(nil).Once()

		status, err := service.SyncReceipt(ctx, receipt)
		require.Error(t, err) // Debe retornar error porque fue rechazada
		assert.Contains(t, err.Error(), "FIRMA INVALIDA")
		assert.Equal(t, "RECHAZADA", status)
	})
}

func TestProcessBackgroundSync(t *testing.T) {
	mockReceiptRepo := new(mocks.MockElectronicReceiptRepository)
	mockSriClient := new(mocks.MockSriClient)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockIssuerRepo := new(mocks.MockIssuerRepository)
	mockClientRepo := new(mocks.MockTaxPayerRepository)
	mockMailService := new(mocks.MockMailService)
	logger := log.New(io.Discard, "", 0)
	
	service := NewSriService(mockTxRepo, mockIssuerRepo, mockReceiptRepo, mockClientRepo, nil, mockSriClient, mockMailService, logger)
	ctx := context.Background()

	// Default expectations for the async finalizeAndEmail (to prevent panics)
	mockTxRepo.On("GetTransactionByID", mock.Anything, mock.Anything).Return(&domain.Transaction{ElectronicReceipt: &domain.ElectronicReceipt{}}, nil).Maybe()
	mockIssuerRepo.On("GetActive", mock.Anything).Return(&domain.Issuer{}, nil).Maybe()
	mockClientRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domain.TaxPayer{}, nil).Maybe()
	mockMailService.On("SendReceipt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()


	t.Run("Should process list of pending receipts", func(t *testing.T) {
		// Arrange
		pending := []domain.ElectronicReceipt{
			{AccessKey: "KEY1", SRIStatus: "RECIBIDA", Environment: 1}, // Will become AUTHORIZED
			{AccessKey: "KEY2", SRIStatus: "PENDIENTE", Environment: 1}, // Will fail send (simulate error)
			{AccessKey: "KEY3", SRIStatus: "EN PROCESO", Environment: 1}, // Will stay EN PROCESO
		}

		// Mock Repo Find (Simulamos que el SQL ya filtró los antiguos, devolviendo estos 3 recientes)
		mockReceiptRepo.On("FindPendingReceipts", ctx).Return(pending, nil).Once()

		// Mock SRI calls
		// 1. KEY1: Authorize Success
		sriAuthResp := &sri.RespuestaAutorizacion{
			NumeroComprobantes: "1",
			Autorizaciones: struct{Autorizacion []sri.Autorizacion `xml:"autorizacion"`}{
				Autorizacion: []sri.Autorizacion{{Estado: "AUTORIZADO", FechaAutorizacion: time.Now().Format(time.RFC3339)}},
			},
		}
		mockSriClient.On("AutorizarComprobante", "KEY1", 1).Return(sriAuthResp, nil).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, "KEY1", "AUTORIZADO", mock.Anything, mock.Anything).Return(nil).Once()

		// 2. KEY2: Send Fail (Network Error)
		mockSriClient.On("EnviarComprobante", mock.Anything, 1).Return(nil, fmt.Errorf("timeout")).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, "KEY2", "ERROR_RED", mock.Anything, mock.Anything).Return(nil).Once()

		// 3. KEY3: Still Processing
		sriPendingResp := &sri.RespuestaAutorizacion{NumeroComprobantes: "0"}
		mockSriClient.On("AutorizarComprobante", "KEY3", 1).Return(sriPendingResp, nil).Once()
		mockReceiptRepo.On("UpdateStatus", ctx, "KEY3", "EN PROCESO", mock.Anything, mock.Anything).Return(nil).Once()

		// Act
		count, err := service.ProcessBackgroundSync(ctx)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should count 1 authorized receipt") // KEY1 es el único finalizado OK
		mockSriClient.AssertExpectations(t)
	})

	t.Run("Should handle empty list (Filter applied by Repo)", func(t *testing.T) {
		// Arrange: Repo returns empty list (SQL filtered old pending receipts)
		mockReceiptRepo.On("FindPendingReceipts", ctx).Return([]domain.ElectronicReceipt{}, nil).Once()

		// Act
		count, err := service.ProcessBackgroundSync(ctx)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		mockSriClient.AssertNotCalled(t, "AutorizarComprobante")
	})
}