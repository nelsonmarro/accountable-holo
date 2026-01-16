package service

import (
	"context"
	"crypto/rand"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/sri"
)

// MailService define el contrato para enviar notificaciones, requerido por SriService.
type MailService interface {
	SendReceipt(issuer *domain.Issuer, recipientEmail string, xmlPath string, pdfPath string) error
}

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
	// Si el XMLContent está vacío (porque GetTransactionByID no lo trae por performance), lo buscamos explícitamente
	xmlContent := tx.ElectronicReceipt.XMLContent
	if xmlContent == "" {
		fullReceipt, err := s.receiptRepo.GetByAccessKey(ctx, tx.ElectronicReceipt.AccessKey)
		if err != nil {
			return "", fmt.Errorf("error recuperando contenido XML de la factura: %w", err)
		}
		if fullReceipt == nil {
			return "", errors.New("el recibo electrónico existe en la transacción pero no se encontró en la tabla de recibos")
		}
		xmlContent = fullReceipt.XMLContent
	}

	var factura sri.Factura
	if err := xml.Unmarshal([]byte(xmlContent), &factura); err != nil {
		return "", fmt.Errorf("error al leer el XML guardado: %w", err)
	}

	// 4. Definir ruta de salida (temp o persistente)
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("ride-%s-*.pdf", tx.ElectronicReceipt.AccessKey))
	if err != nil {
		return "", fmt.Errorf("error creando archivo temporal: %w", err)
	}
	outputPath := tmpFile.Name()
	tmpFile.Close()

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
	if receipt.SRIStatus == "RECIBIDA" || receipt.SRIStatus == "EN PROCESO" {
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

			case "NO AUTORIZADO", "RECHAZADA", "RECHAZADO":

				msg := "Rechazado"

				if len(auth.Mensajes.Mensaje) > 0 {
					msg = auth.Mensajes.Mensaje[0].Mensaje
				}

				_ = s.receiptRepo.UpdateStatus(ctx, receipt.AccessKey, "RECHAZADA", msg, &authDate)

				return "RECHAZADA", fmt.Errorf("%s", msg)

			}
		} else {
			// Si no hay respuesta de autorización pero tampoco error, el SRI sigue procesando
			// Forzamos estado EN PROCESO para que el siguiente ciclo no lo ignore si estaba en RECIBIDA
			_ = s.receiptRepo.UpdateStatus(ctx, receipt.AccessKey, "EN PROCESO", "SRI procesando autorización...", nil)
			return "EN PROCESO", nil
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
		}
	}

	return authorizedCount, nil
}

// EmitirFactura orquesta el proceso completo de facturación electrónica.
func (s *SriService) EmitirFactura(ctx context.Context, transactionID int, signaturePassword string) error {
	s.logger.Printf("Iniciando emisión de factura para transacción ID: %d", transactionID)

	// 1. Obtener Datos Base
	tx, err := s.txRepo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("error obteniendo transacción: %w", err)
	}

	// VALIDACIÓN CRÍTICA: Solo se pueden facturar INGRESOS (Ventas)
	if tx.Category.Type == domain.Outcome {
		return errors.New("no se pueden generar facturas de venta para transacciones de egreso (gastos)")
	}

	issuer, err := s.issuerRepo.GetActive(ctx)
	if err != nil || issuer == nil {
		return errors.New("no hay un emisor activo configurado")
	}

	var claveAcceso string
	var secuencialSRI string
	isNewReceipt := true

	// 2. Determinar si reusamos o creamos clave nueva
	if tx.ElectronicReceipt != nil {
		status := tx.ElectronicReceipt.SRIStatus

		// Detectar zombies: EN PROCESO por más de 2 horas
		isStuck := status == "EN PROCESO" && tx.ElectronicReceipt.CreatedAt.Add(2*time.Hour).Before(time.Now())

		// Si falló definitivamente O está trabada, FORZAMOS nueva clave
		if status == "NO AUTORIZADO" || status == "RECHAZADA" || status == "DEVUELTA" || isStuck {
			isNewReceipt = true
			if isStuck {
				s.logger.Printf("Transacción trabada (%s desde %v). Forzando nueva Clave de Acceso.", status, tx.ElectronicReceipt.CreatedAt)
			} else {
				s.logger.Printf("Anterior falló (%s). Forzando nueva Clave de Acceso.", status)
			}
		} else {
			claveAcceso = tx.ElectronicReceipt.AccessKey
			secuencialSRI = claveAcceso[30:39]
			isNewReceipt = false
			s.logger.Printf("Reusando Clave de Acceso: %s", claveAcceso)
		}
	}
	if isNewReceipt {
		// --- Lógica de Generación de Nuevo Secuencial ---
		// ... (Carga de items, cliente, etc.)
		items, err := s.txRepo.GetItemsByTransactionID(ctx, transactionID)
		if err == nil {
			tx.Items = items
		}

		var client *domain.TaxPayer
		if tx.TaxPayerID != nil {
			client, _ = s.clientRepo.GetByID(ctx, *tx.TaxPayerID)
		}
		if client == nil {
			client, _ = s.clientRepo.GetByIdentification(ctx, "9999999999999")
			if client == nil {
				client = &domain.TaxPayer{Identification: "9999999999999", Name: "CONSUMIDOR FINAL", Email: ""}
				_ = s.clientRepo.Create(ctx, client)
			}
		}

		ep, err := s.epRepo.GetByPoint(ctx, issuer.ID, issuer.EstablishmentCode, issuer.EmissionPointCode, "01")
		if err != nil {
			return err
		}
		if ep == nil {
			ep = &domain.EmissionPoint{IssuerID: issuer.ID, EstablishmentCode: issuer.EstablishmentCode, EmissionPointCode: issuer.EmissionPointCode, ReceiptType: "01", CurrentSequence: 0, IsActive: true}
			_ = s.epRepo.Create(ctx, ep)
		}
		if err := s.epRepo.IncrementSequence(ctx, ep.ID); err != nil {
			return err
		}
		secuencialSRI = fmt.Sprintf("%09d", ep.CurrentSequence+1)

		// Generación de Código Numérico Seguro
		// Usamos crypto/rand para evitar colisiones de claves en reinicios
		nSafe, _ := rand.Int(rand.Reader, big.NewInt(100000000))
		numericCode := fmt.Sprintf("%08d", nSafe.Int64())
		claveAcceso = sri.GenerateAccessKey(tx.TransactionDate, "01", issuer.RUC, issuer.Environment, issuer.EstablishmentCode, issuer.EmissionPointCode, secuencialSRI, numericCode, 1)
	}

	// 3. Generar y Firmar XML
	// Recuperar cliente para mapeo (por si cambió isNewReceipt)
	var clientMapping *domain.TaxPayer
	if tx.TaxPayerID != nil {
		clientMapping, _ = s.clientRepo.GetByID(ctx, *tx.TaxPayerID)
	}
	if clientMapping == nil {
		clientMapping, _ = s.clientRepo.GetByIdentification(ctx, "9999999999999")
	}

	facturaXML := s.mapTransactionToFactura(tx, issuer, clientMapping, claveAcceso, secuencialSRI)
	xmlBytes, err := sri.MarshalFactura(facturaXML)
	if err != nil {
		return err
	}

	s.logger.Printf("Firmando XML...")
	signerObj := sri.NewDocumentSigner(issuer.SignaturePath, signaturePassword)
	// Usamos la nueva opción de algoritmo.
	// Prueba sri.SHA1 para máxima compatibilidad o sri.SHA256 para modernidad.
	signedXML, err := signerObj.Sign(xmlBytes, sri.SHA1)
	if err != nil {
		s.logger.Printf("ERROR CRÍTICO AL FIRMAR: %v", err)
		errStr := err.Error()
		if strings.Contains(errStr, "no such file") || strings.Contains(errStr, "system cannot find") {
			return fmt.Errorf("no se encuentra el archivo de firma (.p12) en la ruta configurada. Verifique la configuración del emisor")
		}
		if strings.Contains(errStr, "password") || strings.Contains(errStr, "mac check failed") {
			return fmt.Errorf("contraseña de firma incorrecta")
		}
		return fmt.Errorf("error técnico al firmar: %w", err)
	}

	// Limpieza de seguridad post-firmado
	signedXMLStr := strings.TrimSpace(string(signedXML))
	s.logger.Printf("XML firmado exitosamente")

	// 5. Guardar o Actualizar en DB Local
	if isNewReceipt {
		receipt := &domain.ElectronicReceipt{
			TransactionID: tx.ID, IssuerID: issuer.ID, TaxPayerID: clientMapping.ID,
			AccessKey: claveAcceso, ReceiptType: "01", XMLContent: signedXMLStr,
			SRIStatus: "PENDIENTE", Environment: issuer.Environment,
		}
		if err := s.receiptRepo.Create(ctx, receipt); err != nil {
			return err
		}
	} else {
		_ = s.receiptRepo.UpdateXML(ctx, claveAcceso, signedXMLStr)
		_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, "PENDIENTE", "Re-emisión corregida", nil)
	}

	// 6. Enviar al SRI
	s.logger.Printf("Enviando al SRI (Ambiente: %d)...", issuer.Environment)
	respRecepcion, err := s.sriClient.EnviarComprobante([]byte(signedXMLStr), issuer.Environment)
	if err != nil {
		_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, "ERROR_RED", err.Error(), nil)
		return err
	}

	s.logger.Printf("Respuesta recepción SRI: %s", respRecepcion.Estado)
	if respRecepcion.Estado == "DEVUELTA" {
		msg := "Error desconocido"
		if len(respRecepcion.Comprobantes.Comprobante) > 0 && len(respRecepcion.Comprobantes.Comprobante[0].Mensajes.Mensaje) > 0 {
			m := respRecepcion.Comprobantes.Comprobante[0].Mensajes.Mensaje[0]
			msg = m.Mensaje
			if !strings.Contains(strings.ToUpper(msg), "EN PROCESAMIENTO") {
				_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, "DEVUELTA", msg, nil)
				s.logger.Printf("CRITICAL SRI ERROR (DEVUELTA): %s - %s", m.Identificador, msg)
				return fmt.Errorf("comprobante devuelto: %s", msg)
			}
			s.logger.Printf("La clave ya está en procesamiento, continuando a autorización...")
		} else {
			_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, "DEVUELTA", msg, nil)
			return fmt.Errorf("comprobante devuelto: %s", msg)
		}
	}

	// 7. Consultar Autorización
	time.Sleep(3 * time.Second)
	s.logger.Printf("Consultando autorización...")
	respAuth, err := s.sriClient.AutorizarComprobante(claveAcceso, issuer.Environment)
	if err == nil && len(respAuth.Autorizaciones.Autorizacion) > 0 {
		auth := respAuth.Autorizaciones.Autorizacion[0]
		authDate, _ := time.Parse(time.RFC3339, auth.FechaAutorizacion)
		s.logger.Printf("Estado final SRI: %s", auth.Estado)

		msg := "Procesado"
		if auth.Estado != "AUTORIZADO" && len(auth.Mensajes.Mensaje) > 0 {
			m := auth.Mensajes.Mensaje[0]
			msg = fmt.Sprintf("%s: %s (%s)", m.Identificador, m.Mensaje, m.InformacionAdicional)
			s.logger.Printf("MENSAJE SRI: %s", msg)
		}

		_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, auth.Estado, msg, &authDate)

		if auth.Estado == "AUTORIZADO" {
			receipt := &domain.ElectronicReceipt{
				TransactionID: tx.ID, AccessKey: claveAcceso, SRIStatus: "AUTORIZADO",
				XMLContent: string(signedXML), AuthorizationDate: &authDate,
			}
			go s.finalizeAndEmail(context.Background(), receipt)
		}
	} else if err == nil {
		// No hay error de red, pero tampoco autorizaciones -> SRI sigue procesando
		s.logger.Printf("El SRI aún está procesando el comprobante. Estado: EN PROCESO")
		_ = s.receiptRepo.UpdateStatus(ctx, claveAcceso, "EN PROCESO", "Esperando autorización...", nil)
	}

	return nil
}

// finalizeAndEmail handles the post-authorization steps: RIDE generation and Email sending.
func (s *SriService) finalizeAndEmail(ctx context.Context, receipt *domain.ElectronicReceipt) {
	// 1. Generate RIDE (PDF) in a temporary file
	// We create a temp file pattern. The generator will write to this path.
	tmpPDF, err := os.CreateTemp("", fmt.Sprintf("ride-%s-*.pdf", receipt.AccessKey))
	if err != nil {
		s.logger.Printf("Failed to create temp PDF file: %v", err)
		return
	}
	pdfPath := tmpPDF.Name()
	if err := tmpPDF.Close(); err != nil {
		s.logger.Printf("Failed to close temp PDF file: %v", err)
	}

	// Ensure PDF is removed after function exits
	defer func() { _ = os.Remove(pdfPath) }()

	// 2. Get Issuer for Logo and Data
	issuer, err := s.issuerRepo.GetActive(ctx)
	if err != nil {
		s.logger.Printf("Failed to get issuer for RIDE generation: %v", err)
		return
	}

	// 3. Parse XML stored in DB
	var factura sri.Factura
	if err := xml.Unmarshal([]byte(receipt.XMLContent), &factura); err != nil {
		s.logger.Printf("Failed to unmarshal XML for RIDE generation: %v", err)
		return
	}

	// 4. Generate the PDF content into the temp path
	// Use Authorization Date if available, else Now
	authDate := time.Now()
	if receipt.AuthorizationDate != nil {
		authDate = *receipt.AuthorizationDate
	}

	err = s.rideGen.GenerateFacturaRide(&factura, pdfPath, issuer.LogoPath, authDate, receipt.AccessKey)
	if err != nil {
		s.logger.Printf("Failed to generate RIDE PDF for %s: %v", receipt.AccessKey, err)
		return
	}

	// 5. Get Recipient
	client, _ := s.clientRepo.GetByID(ctx, receipt.TaxPayerID)
	if client == nil || client.Email == "" {
		s.logger.Printf("Skipping email for %s: no recipient email found", receipt.AccessKey)
		return
	}

	// 6. Write XML to temp file for attachment
	tmpXML, err := os.CreateTemp("", "factura-*.xml")
	if err != nil {
		s.logger.Printf("Failed to create temp XML file: %v", err)
		return
	}
	defer func() { _ = os.Remove(tmpXML.Name()) }()
	_, _ = tmpXML.WriteString(receipt.XMLContent)
	_ = tmpXML.Close()

	// 7. Send Email with both temp files
	err = s.mailService.SendReceipt(issuer, client.Email, tmpXML.Name(), pdfPath)
	if err != nil {
		s.logger.Printf("Failed to send email for %s: %v", receipt.AccessKey, err)
	} else {
		s.logger.Printf("Email successfully sent for %s to %s", receipt.AccessKey, client.Email)
	}
}

func (s *SriService) mapTransactionToFactura(tx *domain.Transaction, issuer *domain.Issuer, client *domain.TaxPayer, claveAcceso string, secuencialSRI string) *sri.Factura {
	f := &sri.Factura{}

	// Función auxiliar para limpiar strings de saltos de línea y tabulaciones
	clean := func(str string) string {
		str = strings.ReplaceAll(str, "\n", " ")
		str = strings.ReplaceAll(str, "\r", "")
		str = strings.ReplaceAll(str, "\t", " ")
		return strings.TrimSpace(str)
	}

	f.InfoTributaria = sri.InfoTributaria{
		Ambiente:        strconv.Itoa(issuer.Environment),
		TipoEmision:     "1",
		RazonSocial:     clean(issuer.BusinessName),
		NombreComercial: clean(issuer.TradeName),
		Ruc:             issuer.RUC,
		ClaveAcceso:     claveAcceso,
		CodDoc:          "01",
		Estab:           issuer.EstablishmentCode,
		PtoEmi:          issuer.EmissionPointCode,
		Secuencial:      secuencialSRI,
		DirMatriz:       clean(issuer.MainAddress),
	}

	// Calculamos subtotales formateados para asegurar que la suma final coincida con lo que el SRI lee
	s15Str := fmt.Sprintf("%.2f", tx.Subtotal15)
	s0Str := fmt.Sprintf("%.2f", tx.Subtotal0)
	taxStr := fmt.Sprintf("%.2f", tx.TaxAmount)
	totalStr := fmt.Sprintf("%.2f", tx.Amount)

	f.InfoFactura = sri.InfoFactura{
		FechaEmision:                tx.TransactionDate.Format("02/01/2006"),
		DirEstablecimiento:          clean(issuer.EstablishmentAddress),
		ObligadoContabilidad:        map[bool]string{true: "SI", false: "NO"}[issuer.KeepAccounting],
		TipoIdentificacionComprador: client.IdentificationType,
		RazonSocialComprador:        clean(client.Name),
		IdentificacionComprador:     client.Identification,
		DireccionComprador:          clean(client.Address),
		TotalSinImpuestos:           fmt.Sprintf("%.2f", tx.Subtotal15+tx.Subtotal0),
		TotalDescuento:              "0.00",
		Propina:                     "0.00",
		ImporteTotal:                totalStr,
		Moneda:                      "DOLAR",
	}

	// Impuestos Totales
	if tx.Subtotal15 > 0 {
		f.InfoFactura.TotalConImpuestos.TotalImpuesto = append(f.InfoFactura.TotalConImpuestos.TotalImpuesto, sri.TotalImpuesto{
			Codigo:           "2",
			CodigoPorcentaje: "4", // 15% (Tarifa vigente 2026)
			BaseImponible:    s15Str,
			Valor:            taxStr,
		})
	}

	if tx.Subtotal0 > 0 {
		f.InfoFactura.TotalConImpuestos.TotalImpuesto = append(f.InfoFactura.TotalConImpuestos.TotalImpuesto, sri.TotalImpuesto{
			Codigo:           "2",
			CodigoPorcentaje: "0", // 0%
			BaseImponible:    s0Str,
			Valor:            "0.00",
		})
	}

	// Pago
	f.InfoFactura.Pagos.Pago = append(f.InfoFactura.Pagos.Pago, sri.Pago{
		FormaPago: "01", // Efectivo/Otros sin utilizacion sistema financiero por defecto
		Total:     totalStr,
	})

	// Detalles
	if len(tx.Items) > 0 {
		for _, item := range tx.Items {
			valTax := 0.0
			codPerc := "0"
			tarifa := "0"
			if item.TaxRate == 4 {
				codPerc = "4"
				valTax = item.Subtotal * 0.15
				tarifa = "15"
			}

			det := sri.Detalle{
				Descripcion:            clean(item.Description),
				Cantidad:               fmt.Sprintf("%.6f", item.Quantity),
				PrecioUnitario:         fmt.Sprintf("%.6f", item.UnitPrice),
				Descuento:              "0.00",
				PrecioTotalSinImpuesto: fmt.Sprintf("%.2f", item.Subtotal),
			}

			impDet := sri.Impuesto{
				Codigo:           "2",
				CodigoPorcentaje: codPerc,
				Tarifa:           tarifa,
				BaseImponible:    fmt.Sprintf("%.2f", item.Subtotal),
				Valor:            fmt.Sprintf("%.2f", valTax),
			}
			det.Impuestos.Impuesto = append(det.Impuestos.Impuesto, impDet)
			f.Detalles.Detalle = append(f.Detalles.Detalle, det)

		}
	} else {

		// Fallback sanitizado
		codPerc := "0"
		tarifa := "0"
		if tx.TaxAmount > 0 {
			codPerc = "4"
			tarifa = "15"
		}

		f.Detalles.Detalle = append(f.Detalles.Detalle, sri.Detalle{
			Descripcion:            clean(tx.Description),
			Cantidad:               "1.000000",
			PrecioUnitario:         fmt.Sprintf("%.6f", tx.Subtotal15+tx.Subtotal0),
			Descuento:              "0.00",
			PrecioTotalSinImpuesto: fmt.Sprintf("%.2f", tx.Subtotal15+tx.Subtotal0),
		})

		f.Detalles.Detalle[0].Impuestos.Impuesto = append(f.Detalles.Detalle[0].Impuestos.Impuesto, sri.Impuesto{
			Codigo:           "2",
			CodigoPorcentaje: codPerc,
			Tarifa:           tarifa,
			BaseImponible:    fmt.Sprintf("%.2f", tx.Subtotal15+tx.Subtotal0),
			Valor:            taxStr,
		})
	}

	return f
}
