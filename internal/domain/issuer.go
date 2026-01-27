package domain

type Issuer struct {
	BaseEntity
	RUC                  string `db:"ruc"`
	BusinessName         string `db:"business_name"`
	TradeName            string `db:"trade_name"`
	MainAddress          string `db:"main_address"`
	EstablishmentAddress string `db:"establishment_address"`
	EstablishmentCode    string `db:"establishment_code"`
	EmissionPointCode    string `db:"emission_point_code"`
	ContributionClass    string `db:"contribution_class"`
	WithholdingAgent     string `db:"withholding_agent"`
	RimpeType            string `db:"rimpe_type"`
	Environment          int    `db:"environment"`
	KeepAccounting       bool   `db:"keep_accounting"`
	SignaturePath        string `db:"signature_path"`
	LogoPath             string `db:"logo_path"`
	IsActive             bool   `db:"is_active"`
	DefaultTaxRate       int    `db:"default_tax_rate"` // 0=None, 2=0%, 4=15%, 6=Exempt

	// Configuración de Correo (SMTP)
	SMTPServer   *string `db:"smtp_server"`
	SMTPPort     *int    `db:"smtp_port"`
	SMTPUser     *string `db:"smtp_user"`
	SMTPPassword *string `db:"smtp_password"`
	SMTPSSL      bool    `db:"smtp_ssl"`
}
