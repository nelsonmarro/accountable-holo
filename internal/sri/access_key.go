// Package sri provides logic to talk to SRI.
package sri

import (
	"fmt"
	"strconv"
	"time"
)

func GenerateAccessKey(
	date time.Time,
	codDoc string,
	ruc string,
	environment int,
	establishment string,
	emissionPoint string,
	sequential string,
	numericCode string,
	emissionType int,
) string {
	// Formato Fecha: ddmmaaaa
	dateStr := date.Format("02012006")

	// Concatenar los primeros 48 dígitos
	key48 := fmt.Sprintf("%s%s%s%d%s%s%s%s%d",
		dateStr,
		codDoc,
		ruc,
		environment,
		establishment,
		emissionPoint,
		sequential,
		numericCode,
		emissionType,
	)

	verifier := computeMod11(key48)

	return key48 + strconv.Itoa(verifier)
}

func computeMod11(digits string) int {
	sum := 0
	factor := 2
	for i := len(digits) - 1; i >= 0; i-- {
		num, _ := strconv.Atoi(string(digits[i]))
		sum += num * factor
		factor++
		if factor > 7 {
			factor = 2
		}
	}

	res := 11 - (sum % 11)

	if res == 11 {
		return 0
	}

	if res == 10 {
		return 1
	}

	return res
}
