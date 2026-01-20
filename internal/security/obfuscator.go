// Package security proporciona funciones para ofuscar y desofuscar datos sensibles.
package security

import (
	"encoding/base64"
)

// InternalKey es la llave usada para el XOR.
// En un escenario real, esto debería ser algo único y no obvio.
const InternalKey = "fc0f9971-3a7e-4610-b20e-f3789b014f89"

// SimpleXOR realiza una operación XOR entre el input y la llave.
// Funciona tanto para cifrar como para descifrar (es simétrico).
func SimpleXOR(input, key string) string {
	output := make([]byte, len(input))
	for i := 0; i < len(input); i++ {
		output[i] = input[i] ^ key[i%len(key)]
	}
	return string(output)
}

// DecodeSMTPPassword toma una cadena en Base64 (inyectada al compilar),
// la decodifica y luego aplica XOR para recuperar el texto plano.
func DecodeSMTPPassword(encodedInput string) (string, error) {
	if encodedInput == "" {
		return "", nil
	}

	// 1. Decodificar Base64
	data, err := base64.StdEncoding.DecodeString(encodedInput)
	if err != nil {
		return "", err
	}

	// 2. Descifrar XOR
	return SimpleXOR(string(data), InternalKey), nil
}
