package sri

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

const (
	// Ambiente de Pruebas (Sandbox)
	URLRecepcionPruebas    = "https://celcer.sri.gob.ec/comprobantes-electronicos-ws/RecepcionComprobantesOffline"
	URLAutorizacionPruebas = "https://celcer.sri.gob.ec/comprobantes-electronicos-ws/AutorizacionComprobantesOffline"

	// Ambiente de Producción (Real)
	URLRecepcionProduccion    = "https://cel.sri.gob.ec/comprobantes-electronicos-ws/RecepcionComprobantesOffline"
	URLAutorizacionProduccion = "https://cel.sri.gob.ec/comprobantes-electronicos-ws/AutorizacionComprobantesOffline"
)

// Client defines the interface for interacting with SRI Web Services.
type Client interface {
	EnviarComprobante(xmlFirmado []byte, environment int) (*RespuestaRecepcion, error)
	AutorizarComprobante(claveAcceso string, environment int) (*RespuestaAutorizacion, error)
}

// SoapClient implements the Client interface using standard HTTP.
type SoapClient struct {
	Timeout time.Duration
}

// NewSoapClient creates a new SRI client.
func NewSoapClient() *SoapClient {
	return &SoapClient{
		Timeout: 30 * time.Second,
	}
}

// --- RECEPCIÓN ---

type RespuestaRecepcion struct {
	Estado       string `xml:"estado"`
	Comprobantes struct {
		Comprobante []ComprobanteRecepcion `xml:"comprobante"`
	} `xml:"comprobantes"`
}

type ComprobanteRecepcion struct {
	ClaveAcceso string `xml:"claveAcceso"`
	Mensajes    struct {
		Mensaje []Mensaje `xml:"mensaje"`
	} `xml:"mensajes"`
}

type Mensaje struct {
	Identificador        string `xml:"identificador"`
	Mensaje              string `xml:"mensaje"`
	InformacionAdicional string `xml:"informacionAdicional"`
	Tipo                 string `xml:"tipo"`
}

type SoapRecepcionEnvelope struct {
	Body struct {
		ValidarComprobanteResponse struct {
			RespuestaRecepcion RespuestaRecepcion `xml:"RespuestaRecepcionComprobante"`
		} `xml:"validarComprobanteResponse"`
	} `xml:"Body"`
}

// --- AUTORIZACIÓN ---

type RespuestaAutorizacion struct {
	ClaveAccesoConsultada string `xml:"claveAccesoConsultada"`
	NumeroComprobantes    string `xml:"numeroComprobantes"`
	Autorizaciones        struct {
		Autorizacion []Autorizacion `xml:"autorizacion"`
	} `xml:"autorizaciones"`
}

type Autorizacion struct {
	Estado             string `xml:"estado"` // AUTORIZADO, NO AUTORIZADO, EN PROCESO
	NumeroAutorizacion string `xml:"numeroAutorizacion"`
	FechaAutorizacion  string `xml:"fechaAutorizacion"`
	Ambiente           string `xml:"ambiente"`
	Comprobante        string `xml:"comprobante"` // El XML autorizado (CDATA)
	Mensajes           struct {
		Mensaje []Mensaje `xml:"mensaje"`
	} `xml:"mensajes"`
}

type SoapAutorizacionEnvelope struct {
	Body struct {
		AutorizacionComprobanteResponse struct {
			RespuestaAutorizacion RespuestaAutorizacion `xml:"RespuestaAutorizacionComprobante"`
		} `xml:"autorizacionComprobanteResponse"`
	} `xml:"Body"`
}

// EnviarComprobante envía el XML firmado al Web Service de Recepción del SRI.
func (c *SoapClient) EnviarComprobante(xmlFirmado []byte, environment int) (*RespuestaRecepcion, error) {
	// 1. Seleccionar URL según ambiente
	url := URLRecepcionPruebas
	if environment == 2 {
		url = URLRecepcionProduccion
	}

	// 2. Codificar XML a Base64
	xmlBase64 := base64.StdEncoding.EncodeToString(xmlFirmado)

	// 3. Construir Envelope SOAP
	soapEnvelope := fmt.Sprintf(`
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ec="http://ec.gob.sri.ws.recepcion">
   <soapenv:Header/>
   <soapenv:Body>
      <ec:validarComprobante>
         <xml>%s</xml>
      </ec:validarComprobante>
   </soapenv:Body>
</soapenv:Envelope>`, xmlBase64)

	return c.ejecutarRecepcion(url, soapEnvelope)
}

func (c *SoapClient) ejecutarRecepcion(url, soapEnvelope string) (*RespuestaRecepcion, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(soapEnvelope))
	if err != nil {
		return nil, fmt.Errorf("error creando petición SOAP: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; charset=UTF-8")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de conexión con el SRI: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("el SRI respondió con estado HTTP: %d", resp.StatusCode)
	}

	var envelope SoapRecepcionEnvelope
	if err := xml.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("error leyendo respuesta del SRI: %w", err)
	}

	return &envelope.Body.ValidarComprobanteResponse.RespuestaRecepcion, nil
}

// AutorizarComprobante consulta el estado de autorización usando la Clave de Acceso.
func (c *SoapClient) AutorizarComprobante(claveAcceso string, environment int) (*RespuestaAutorizacion, error) {
	url := URLAutorizacionPruebas
	if environment == 2 {
		url = URLAutorizacionProduccion
	}

	soapEnvelope := fmt.Sprintf(`
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ec="http://ec.gob.sri.ws.autorizacion">
   <soapenv:Header/>
   <soapenv:Body>
      <ec:autorizacionComprobante>
         <claveAccesoComprobante>%s</claveAccesoComprobante>
      </ec:autorizacionComprobante>
   </soapenv:Body>
</soapenv:Envelope>`, claveAcceso)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(soapEnvelope))
	if err != nil {
		return nil, fmt.Errorf("error creando petición Autorización: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; charset=UTF-8")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error conectando a Autorización SRI: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SRI Autorización error HTTP: %d", resp.StatusCode)
	}

	var envelope SoapAutorizacionEnvelope
	if err := xml.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta Autorización: %w", err)
	}

	return &envelope.Body.AutorizacionComprobanteResponse.RespuestaAutorizacion, nil
}