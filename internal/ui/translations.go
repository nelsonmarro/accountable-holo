package ui

import (
	"encoding/json"
	"log"

	"fyne.io/fyne/v2/lang"
)

// LoadSpanishTranslations injects Spanish translations for Fyne widgets.
func LoadSpanishTranslations() {
	// Claves estándar que Fyne usa para el calendario y diálogos
	translations := map[string]string{
		// Meses
		"January":   "Enero",
		"February":  "Febrero",
		"March":     "Marzo",
		"April":     "Abril",
		"May":       "Mayo",
		"June":      "Junio",
		"July":      "Julio",
		"August":    "Agosto",
		"September": "Septiembre",
		"October":   "Octubre",
		"November":  "Noviembre",
		"December":  "Diciembre",

		// Meses Cortos (si aplica)
		"Jan": "Ene",
		"Feb": "Feb",
		"Mar": "Mar",
		"Apr": "Abr",
		"Jun": "Jun",
		"Jul": "Jul",
		"Aug": "Ago",
		"Sep": "Sep",
		"Oct": "Oct",
		"Nov": "Nov",
		"Dec": "Dic",

		// Días
		"Sunday":    "Domingo",
		"Monday":    "Lunes",
		"Tuesday":   "Martes",
		"Wednesday": "Miércoles",
		"Thursday":  "Jueves",
		"Friday":    "Viernes",
		"Saturday":  "Sábado",

		// Días Cortos
		"Sun": "Dom",
		"Mon": "Lun",
		"Tue": "Mar",
		"Wed": "Mié",
		"Thu": "Jue",
		"Fri": "Vie",
		"Sat": "Sáb",

		// Variantes en MAYÚSCULAS (por si Fyne busca keys en mayúsculas)
		"JAN": "ENE", "FEB": "FEB", "MAR": "MAR", "APR": "ABR", "MAY": "MAY", "JUN": "JUN",
		"JUL": "JUL", "AUG": "AGO", "SEP": "SEP", "OCT": "OCT", "NOV": "NOV", "DEC": "DIC",

		"SUN": "DOM", "MON": "LUN", "TUE": "MAR", "WED": "MIÉ", "THU": "JUE", "FRI": "VIE", "SAT": "SÁB",

		"JANUARY": "ENERO", "FEBRUARY": "FEBRERO", "MARCH": "MARZO", "APRIL": "ABRIL",
		"AUGUST": "AGOSTO", "SEPTEMBER": "SEPTIEMBRE", "OCTOBER": "OCTUBRE", "NOVEMBER": "NOVIEMBRE", "DECEMBER": "DICIEMBRE",

		// Botones Comunes
		"Cancel": "Cancelar",
		"OK":     "Aceptar",
		"Yes":    "Sí",
		"No":     "No",
		"Open":   "Abrir",
		"Save":   "Guardar",
	}

	data, err := json.Marshal(translations)
	if err != nil {
		log.Println("Error encoding translations:", err)
		return
	}

	// Inyectamos esto para el locale 'es'
	err = lang.AddTranslationsForLocale(data, "es")
	if err != nil {
		log.Println("Error adding translations:", err)
	}
}
