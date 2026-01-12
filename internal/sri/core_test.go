package sri

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAccessKey(t *testing.T) {
	t.Run("Generate Valid Key", func(t *testing.T) {
		date := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
		ruc := "1790012345001"
		env := 1
		estab := "001"
		ptoEmi := "002"
		seq := "000000001"
		code := "12345678"
		
		key := GenerateAccessKey(date, "01", ruc, env, estab, ptoEmi, seq, code, 1)
		
		assert.Len(t, key, 49)
		assert.Equal(t, "10012026", key[0:8])   // Fecha ddmmaaaa
		assert.Equal(t, "01", key[8:10])       // CodDoc
		assert.Equal(t, ruc, key[10:23])       // RUC
		assert.Equal(t, "1", key[23:24])       // Amb
	})

	t.Run("Modulo 11 Correctness", func(t *testing.T) {
		// Ejemplo manual
		key48 := "100120260117900123450011001002000000001123456781"
		verifier := computeMod11(key48)
		assert.True(t, verifier >= 0 && verifier <= 9)
	})
}

func TestMarshalFactura(t *testing.T) {
	t.Run("Should include XML declaration", func(t *testing.T) {
		f := &Factura{
			Version: "2.1.0",
		}
		f.InfoTributaria.RazonSocial = "Test Empresa"
		
		bytes, err := MarshalFactura(f)
		assert.NoError(t, err)
		
		xmlStr := string(bytes)
		assert.Contains(t, xmlStr, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		assert.Contains(t, xmlStr, "<factura id=\"comprobante\" version=\"2.1.0\">")
	})
}