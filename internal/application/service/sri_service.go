package service

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/sri"
)

type SriService struct {
	txRepo      TransactionRepository
	issuerRepo  IssuerRepository
	receiptRepo ElectronicReceiptRepository
	clientRepo  TaxPayerRepository
	epRepo      EmissionPointRepository // Added
	sriClient   sri.Client
	rideGen     *sri.RideGenerator
	mailService MailService
	logger      *log.Logger
}

func NewSriService(
	txRepo TransactionRepository,
	issuerRepo IssuerRepository,
	receiptRepo ElectronicReceiptRepository,
	clientRepo TaxPayerRepository,
	epRepo EmissionPointRepository, // Added
	sriClient sri.Client,
	mailService MailService,
	logger *log.Logger,
) *SriService {
	return &SriService{
		txRepo:      txRepo,
		issuerRepo:  issuerRepo,
		receiptRepo: receiptRepo,
		clientRepo:  clientRepo,
		epRepo:      epRepo, // Added
		sriClient:   sriClient,
		mailService: mailService,
		rideGen:     sri.NewRideGenerator(),
		logger:      logger,
	}
}

// GenerateRide genera el PDF (RIDE) de una factura ya emitida.
// Retorna la ruta absoluta del archivo generado.
func (s *SriService) GenerateRide(ctx context.Context, transactionID int) (string, error) {
	// 1. Obtener Transacción y Recibo
	tx, err := s.txRepo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return "", fmt.Errorf("error obteniendo transacción: %w", err)
	}
	if tx.ElectronicReceipt == nil {
		return "", errors.New("esta transacción no tiene factura electrónica asociada")
	}

	// 2. Obtener Issuer para el Logo
	issuer, err := s.issuerRepo.GetActive(ctx)
	if err != nil {
		return "", fmt.Errorf("error obteniendo emisor: %w", err)
	}

	// 3. Parsear XML almacenado
	var factura sri.Factura
	if err := xml.Unmarshal([]byte(tx.ElectronicReceipt.XMLContent), &factura); err != nil {
		return "", fmt.Errorf("error al leer el XML guardado: %w", err)
	}

	// 4. Definir ruta de salida (temp o persistente)
	outputPath := fmt.Sprintf("ride-%s.pdf", tx.ElectronicReceipt.AccessKey)

	// 5. Generar PDF
	// Usamos la fecha de autorización si existe, sino la actual
	authDate := time.Now()
	if tx.ElectronicReceipt.AuthorizationDate != nil {
		authDate = *tx.ElectronicReceipt.AuthorizationDate
	}

	err = s.rideGen.GenerateFacturaRide(&factura, outputPath, issuer.LogoPath, authDate, tx.ElectronicReceipt.AccessKey)
	if err != nil {
		return "", fmt.Errorf("error generando PDF: %w", err)
	}

	// TODO: Guardar path en ElectronicReceipt si queremos cachearlo
	return outputPath, nil
}

// SyncReceipt verifica el estado de un comprobante pendiente y avanza el flujo si es necesario.
// Retorna el nuevo estado y un error si ocurrió.
func (s *SriService) SyncReceipt(ctx context.Context, receipt *domain.ElectronicReceipt) (string, error) {
	// 1. Si nunca se envió exitosamente o falló el envío, intentamos enviar de nuevo
	if receipt.SRIStatus == "PENDIENTE" || receipt.SRIStatus == "ERROR_ENVIO" || receipt.SRIStatus == "ERROR_RED" {
		respRecepcion, err := s.sriClient.EnviarComprobante([]byte(receipt.XMLContent), receipt.Environment)
		if err != nil {
			// Sigue fallando la red/envío
			_ = s.receiptRepo.UpdateStatus(ctx, receipt.AccessKey, "ERROR_RED", err.Error(), nil)
			return "ERROR_RED", err
		}

		if respRecepcion.Estado == "DEVUELTA" {
			msg := "Comprobante Devuelto"
			if len(respRecepcion.Comprobantes.Comprobante) > 0 {
				msgs := respRecepcion.Comprobantes.Comprobante[0].Mensajes.Mensaje
				if len(msgs) > 0 {
					msg = fmt.Sprintf("%s: %s", msgs[0].Identificador, msgs[0].Mensaje)
				}
			}
			_ = s.receiptRepo.UpdateStatus(ctx, receipt.AccessKey, "DEVUELTA", msg, nil)
			return "DEVUELTA", fmt.Errorf("%s", msg)
		}

		// Si pasó a RECIBIDA, actualizamos y seguimos al paso de autorización
		_ = s.receiptRepo.UpdateStatus(ctx, receipt.AccessKey, "RECIBIDA", "Recibido por SRI", nil)
		receipt.SRIStatus = "RECIBIDA"
	}

	// 2. Si ya fue recibida, consultamos autorización
	if receipt.SRIStatus == "RECIBIDA" || receipt.SRIStatus == "EN_PROCESO" {
		respAuth, err := s.sriClient.AutorizarComprobante(receipt.AccessKey, receipt.Environment)
		if err != nil {
			return receipt.SRIStatus, err // Error de red al consultar, mantenemos estado
		}

		if len(respAuth.Autorizaciones.Autorizacion) > 0 {
			auth := respAuth.Autorizaciones.Autorizacion[0]
			authDate, _ := time.Parse(time.RFC3339, auth.FechaAutorizacion)

			switch auth.Estado {

			case "AUTORIZADO":

				_ = s.receiptRepo.UpdateStatus(ctx, receipt.AccessKey, "AUTORIZADO", "Autorización Exitosa", &authDate)

				receipt.SRIStatus = "AUTORIZADO"

				receipt.AuthorizationDate = &authDate

				// 3. Generar RIDE y enviar email

				go s.finalizeAndEmail(context.Background(), receipt)

				return "AUTORIZADO", nil

			case "NO AUTORIZADO", "RECHAZADA":

				msg := "Rechazado"

				if len(auth.Mensajes.Mensaje) > 0 {
					msg = auth.Mensajes.Mensaje[0].Mensaje
				}

				_ = s.receiptRepo.UpdateStatus(ctx, receipt.AccessKey, "RECHAZADA", msg, &authDate)

				return "RECHAZADA", fmt.Errorf("%s", msg)

			}

		}
	}

	return receipt.SRIStatus, nil
}

// ProcessBackgroundSync busca todos los comprobantes pendientes y los actualiza.
// Retorna el número de comprobantes actualizados a estado terminal (Autorizado/Rechazado).
func (s *SriService) ProcessBackgroundSync(ctx context.Context) (int, error) {
	pending, err := s.receiptRepo.FindPendingReceipts(ctx)
	if err != nil {
		return 0, err
	}

	authorizedCount := 0
	for _, r := range pending {
		oldStatus := r.SRIStatus
		newStatus, err := s.SyncReceipt(ctx, &r)
		if err != nil {
			s.logger.Printf("Error sincronizando comprobante %s: %v", r.AccessKey, err)
			continue
		}

		if newStatus == "AUTORIZADO" && oldStatus != "AUTORIZADO" {
			authorizedCount++
			// Aquí podríamos enviar un email al cliente automáticamente
		}
	}

	return authorizedCount, nil
}

// EmitirFactura orquesta el proceso completo de facturación electrónica.
func (s *SriService) EmitirFactura(ctx context.Context, transactionID int, signaturePassword string) error {
	// 1. Obtener Datos
	tx, err := s.txRepo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("error obteniendo transacción: %w", err)
	}

	// Load Items
	items, err := s.txRepo.GetItemsByTransactionID(ctx, transactionID)
	if err == nil {
		tx.Items = items
	}

	issuer, err := s.issuerRepo.GetActive(ctx)
	if err != nil || issuer == nil {
		return errors.New("no hay un emisor activo configurado")
	}

	// 2. Obtener Cliente (TaxPayer)
	var client *domain.TaxPayer
	if tx.TaxPayerID != nil {
		client, _ = s.clientRepo.GetByID(ctx, *tx.TaxPayerID)
	}

	if client == nil {
		// Intentar buscar Consumidor Final en la base de datos
		client, _ = s.clientRepo.GetByIdentification(ctx, "9999999999999")
		if client == nil {
			// Si no existe, crearlo
			client = &domain.TaxPayer{
				Identification:     "9999999999999",
				IdentificationType: "07",
				Name:               "CONSUMIDOR FINAL",
				Address:            "S/D",
				Email:              "",
			}
			_ = s.clientRepo.Create(ctx, client)
		}
	}

	// 3. Obtener Secuencial SRI
	// Buscamos el punto de emisión configurado para facturas (01)
	// Asumimos por defecto estab=001 pto=002, pero debería venir del Issuer config
	ep, err := s.epRepo.GetByPoint(ctx, issuer.ID, issuer.EstablishmentCode, issuer.EmissionPointCode, "01")
	if err != nil {
		// Si no existe, intentamos crearlo la primera vez
		if ep == nil {
			ep = &domain.EmissionPoint{
				IssuerID:          issuer.ID,
				EstablishmentCode: issuer.EstablishmentCode,
				EmissionPointCode: issuer.EmissionPointCode,
				ReceiptType:       "01",
				CurrentSequence:   0,
				IsActive:          true,
			}
			_ = s.epRepo.Create(ctx, ep)
		} else {
			return fmt.Errorf("error obteniendo secuencial: %w", err)
		}
	}
	
	// Incrementamos el secuencial
	if err := s.epRepo.IncrementSequence(ctx, ep.ID); err != nil {
		return fmt.Errorf("error incrementando secuencial: %w", err)
	}
	
	// Formato 9 dígitos (ej: 000000123)
	secuencialSRI := fmt.Sprintf("%09d", ep.CurrentSequence+1)

	// 4. Generar Clave de Acceso
	// Generar código numérico aleatorio de 8 dígitos
	numericCode := fmt.Sprintf("%08d", rand.Intn(100000000))

	claveAcceso := sri.GenerateAccessKey(
		tx.TransactionDate,
		"01", // Factura
		issuer.RUC,
		issuer.Environment,
		issuer.EstablishmentCode,
		issuer.EmissionPointCode,
		secuencialSRI,
		numericCode,
		1,
	)

	// 4. Mapear a Estructura XML
	facturaXML := s.mapTransactionToFactura(tx, issuer, client, claveAcceso)

	// 5. Generar XML Crudo
	xmlBytes, err := sri.MarshalFactura(facturaXML)
	if err != nil {
		return fmt.Errorf("error generando XML: %w", err)
	}

	// 6. Firmar XML Real usando el paquete propio
	signer := sri.NewDocumentSigner(issuer.SignaturePath, signaturePassword)
	signedXML, err := signer.Sign(xmlBytes)
	if err != nil {
		return fmt.Errorf("error al firmar digitalmente: %w", err)
	}

	// 7. Guardar Receipt y proceder con envío (como antes)...
	receipt := &domain.ElectronicReceipt{
		TransactionID: tx.ID,
		IssuerID:      issuer.ID,
		TaxPayerID:    client.ID,
		AccessKey:     claveAcceso,
		ReceiptType:   "01",
		XMLContent:    string(signedXML),
		SRIStatus:     "PENDIENTE",
		Environment:   issuer.Environment,
	}
	if err := s.receiptRepo.Create(ctx, receipt); err != nil {
		return fmt.Errorf("error guardando historial legal: %w", err)
	}

	// 8. Enviar y Autorizar (Lógica síncrona normativa 2026)
	respRecepcion, err := s.sriClient.EnviarComprobante(signedXML, issuer.Environment)
	if err != nil {
		_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, "ERROR_RED", err.Error(), nil)
		return fmt.Errorf("error de conexión SRI: %w", err)
	}

	if respRecepcion.Estado == "DEVUELTA" {
		_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, "DEVUELTA", "Error en esquema XML", nil)
		return errors.New("comprobante devuelto por el SRI")
	}

	// Consultar autorización
	time.Sleep(3 * time.Second)
	respAuth, err := s.sriClient.AutorizarComprobante(claveAcceso, issuer.Environment)
	if err == nil && len(respAuth.Autorizaciones.Autorizacion) > 0 {
		auth := respAuth.Autorizaciones.Autorizacion[0]
		authDate, _ := time.Parse(time.RFC3339, auth.FechaAutorizacion)
		_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, auth.Estado, "Procesado", &authDate)

		switch auth.Estado {
		case "AUTORIZADO":
			receipt.SRIStatus = "AUTORIZADO"
			receipt.AuthorizationDate = &authDate
			go s.finalizeAndEmail(context.Background(), receipt)
		}
	}

	return nil
}

// finalizeAndEmail handles the post-authorization steps: RIDE generation and Email sending.
func (s *SriService) finalizeAndEmail(ctx context.Context, receipt *domain.ElectronicReceipt) {
	// 1. Generate RIDE
	pdfPath, err := s.GenerateRide(ctx, receipt.TransactionID)
	if err != nil {
		s.logger.Printf("Failed to generate RIDE for %s: %v", receipt.AccessKey, err)
		return
	}

	// 2. Get Issuer and TaxPayer
	issuer, _ := s.issuerRepo.GetActive(ctx)
	client, _ := s.clientRepo.GetByID(ctx, receipt.TaxPayerID)

	if client == nil || client.Email == "" {
		s.logger.Printf("Skipping email for %s: no recipient email found", receipt.AccessKey)
		return
	}

	// 3. Write XML to temp file for attachment
	tmpXML, err := os.CreateTemp("", "factura-*.xml")
	if err != nil {
		s.logger.Printf("Failed to create temp XML file: %v", err)
		return
	}
	defer func() { _ = os.Remove(tmpXML.Name()) }()
	_, _ = tmpXML.WriteString(receipt.XMLContent)
	_ = tmpXML.Close()

	// 4. Send Email
	err = s.mailService.SendReceipt(issuer, client.Email, tmpXML.Name(), pdfPath)
	if err != nil {
		s.logger.Printf("Failed to send email for %s: %v", receipt.AccessKey, err)
	} else {
		s.logger.Printf("Email successfully sent for %s to %s", receipt.AccessKey, client.Email)
	}
}

func (s *SriService) mapTransactionToFactura(tx *domain.Transaction, issuer *domain.Issuer, client *domain.TaxPayer, claveAcceso string) *sri.Factura {
	f := &sri.Factura{}
	dirMatriz := issuer.MainAddress
	obligadoContab := "NO"
	if issuer.KeepAccounting {
		obligadoContab = "SI"
	}

	f.InfoTributaria = sri.InfoTributaria{
		Ambiente:        strconv.Itoa(issuer.Environment),
		TipoEmision:     "1",
		RazonSocial:     issuer.BusinessName,
		NombreComercial: issuer.TradeName,
		Ruc:             issuer.RUC,
		ClaveAcceso:     claveAcceso,
		CodDoc:          "01",
		Estab:           issuer.EstablishmentCode,
		PtoEmi:          issuer.EmissionPointCode,
		Secuencial:      tx.TransactionNumber,
		DirMatriz:       dirMatriz,
	}

	f.InfoFactura = sri.InfoFactura{
		FechaEmision:                tx.TransactionDate.Format("02/01/2006"),
		DirEstablecimiento:          issuer.EstablishmentAddress,
		ObligadoContabilidad:        obligadoContab,
		TipoIdentificacionComprador: client.IdentificationType,
		RazonSocialComprador:        client.Name,
		IdentificacionComprador:     client.Identification,
		DireccionComprador:          client.Address,
		TotalSinImpuestos:           fmt.Sprintf("%.2f", tx.Subtotal15+tx.Subtotal0),
		TotalDescuento:              "0.00",
		Propina:                     "0.00",
		ImporteTotal:                fmt.Sprintf("%.2f", tx.Amount),
		Moneda:                      "DOLAR",
	}

	// Impuestos Totales
	if tx.Subtotal15 > 0 {
		f.InfoFactura.TotalConImpuestos.TotalImpuesto = append(f.InfoFactura.TotalConImpuestos.TotalImpuesto, sri.TotalImpuesto{
			Codigo:           "2",
			CodigoPorcentaje: "4", // 15%
			BaseImponible:    fmt.Sprintf("%.2f", tx.Subtotal15),
			Valor:            fmt.Sprintf("%.2f", tx.TaxAmount),
		})
	}
	if tx.Subtotal0 > 0 {
		f.InfoFactura.TotalConImpuestos.TotalImpuesto = append(f.InfoFactura.TotalConImpuestos.TotalImpuesto, sri.TotalImpuesto{
			Codigo:           "2",
			CodigoPorcentaje: "0", // 0%
			BaseImponible:    fmt.Sprintf("%.2f", tx.Subtotal0),
			Valor:            "0.00",
		})
	}

	// Pago (Efectivo por defecto)
	f.InfoFactura.Pagos.Pago = append(f.InfoFactura.Pagos.Pago, sri.Pago{
		FormaPago: "01",
		Total:     fmt.Sprintf("%.2f", tx.Amount),
	})

	// Detalles
	if len(tx.Items) > 0 {
		for _, item := range tx.Items {
			det := sri.Detalle{
				Descripcion:            item.Description,
				Cantidad:               fmt.Sprintf("%.6f", item.Quantity),
				PrecioUnitario:         fmt.Sprintf("%.6f", item.UnitPrice),
				Descuento:              "0.00",
				PrecioTotalSinImpuesto: fmt.Sprintf("%.2f", item.Subtotal),
			}

			// Impuesto del detalle
			codigoPorcentaje := "0"
			valorImpuesto := 0.0
			tarifa := "0"
			if item.TaxRate == 4 {
				codigoPorcentaje = "4"
				valorImpuesto = item.Subtotal * 0.15
				tarifa = "15"
			}

			impDet := sri.Impuesto{
				Codigo:           "2",
				CodigoPorcentaje: codigoPorcentaje,
				Tarifa:           tarifa,
				BaseImponible:    fmt.Sprintf("%.2f", item.Subtotal),
				Valor:            fmt.Sprintf("%.2f", valorImpuesto),
			}
			det.Impuestos.Impuesto = append(det.Impuestos.Impuesto, impDet)
			f.Detalles.Detalle = append(f.Detalles.Detalle, det)
		}
	} else {
		// Fallback: Un solo item basado en la descripción general si no hay ítems desglosados
		f.Detalles.Detalle = append(f.Detalles.Detalle, sri.Detalle{
			Descripcion:            tx.Description,
			Cantidad:               "1.000000",
			PrecioUnitario:         fmt.Sprintf("%.6f", tx.Subtotal15+tx.Subtotal0),
			Descuento:              "0.00",
			PrecioTotalSinImpuesto: fmt.Sprintf("%.2f", tx.Subtotal15+tx.Subtotal0),
		})

		// IVA del fallback
		codigoPorcentaje := "0"
		if tx.TaxAmount > 0 {
			codigoPorcentaje = "4"
		}

		impDet := sri.Impuesto{
			Codigo:           "2",
			CodigoPorcentaje: codigoPorcentaje,
			Tarifa:           map[string]string{"4": "15", "0": "0"}[codigoPorcentaje],
			BaseImponible:    fmt.Sprintf("%.2f", tx.Subtotal15+tx.Subtotal0),
			Valor:            fmt.Sprintf("%.2f", tx.TaxAmount),
		}
		f.Detalles.Detalle[0].Impuestos.Impuesto = append(f.Detalles.Detalle[0].Impuestos.Impuesto, impDet)
	}

	return f
}
