package transaction

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
)

type RecurringForm struct {
	window          fyne.Window
	service         RecurringTransactionService
	accountService  AccountService
	categoryService CategoryService
	item            *domain.RecurringTransaction // Nil if new
	OnSaved         func()

	descEntry      *widget.Entry
	amountEntry    *widget.Entry
	intervalEntry  *widget.Select
	startDate      *widget.Entry // YYYY-MM-DD
	accountSelect  *widget.Select
	categorySelect *widget.Select
	activeCheck    *widget.Check

	accounts   []domain.Account
	categories []domain.Category
}

func NewRecurringForm(
	parent fyne.Window,
	item *domain.RecurringTransaction,
	service RecurringTransactionService,
	accountService AccountService,
	categoryService CategoryService,
) *RecurringForm {
	return &RecurringForm{
		window:          parent,
		item:            item,
		service:         service,
		accountService:  accountService,
		categoryService: categoryService,
	}
}

func (f *RecurringForm) Show() {
	f.descEntry = widget.NewEntry()
	f.amountEntry = widget.NewEntry()
	f.startDate = widget.NewEntry()
	f.startDate.SetPlaceHolder("YYYY-MM-DD")
	f.activeCheck = widget.NewCheck("Activo", nil)
	f.activeCheck.SetChecked(true)

	// Options in Spanish
	f.intervalEntry = widget.NewSelect([]string{"SEMANAL", "MENSUAL"}, nil)
	f.accountSelect = widget.NewSelect(nil, nil)
	f.categorySelect = widget.NewSelect(nil, nil)

	// Load Data
	f.loadOptions()

	// Fill if editing
	if f.item != nil {
		f.descEntry.SetText(f.item.Description)
		f.amountEntry.SetText(fmt.Sprintf("%.2f", f.item.Amount))

		// Map from DB to UI Spanish
		if f.item.Interval == domain.IntervalWeekly {
			f.intervalEntry.SetSelected("SEMANAL")
		} else {
			f.intervalEntry.SetSelected("MENSUAL")
		}

		f.startDate.SetText(f.item.NextRunDate.Format("2006-01-02"))
		f.activeCheck.SetChecked(f.item.IsActive)

		// Find selected account/category names
		for _, acc := range f.accounts {
			if acc.ID == f.item.AccountID {
				f.accountSelect.SetSelected(acc.Name)
				break
			}
		}
		for _, cat := range f.categories {
			if cat.ID == f.item.CategoryID {
				f.categorySelect.SetSelected(cat.Name)
				break
			}
		}
	} else {
		// Defaults
		f.startDate.SetText(time.Now().Format("2006-01-02"))
		f.intervalEntry.SetSelected("MENSUAL")
	}

	items := []*widget.FormItem{
		{Text: "Descripción", Widget: f.descEntry},
		{Text: "Monto", Widget: f.amountEntry},
		{Text: "Cuenta", Widget: f.accountSelect},
		{Text: "Categoría", Widget: f.categorySelect},
		{Text: "Frecuencia", Widget: f.intervalEntry},
		{Text: "Próxima Ejecución", Widget: f.startDate},
		{Text: "Estado", Widget: f.activeCheck},
	}

	d := dialog.NewForm("Regla Recurrente", "Guardar", "Cancelar", items, f.onSave, f.window)
	d.Resize(fyne.NewSize(600, 500))
	d.Show()
}

func (f *RecurringForm) loadOptions() {
	// Accounts
	accs, err := f.accountService.GetAllAccounts(context.Background())
	if err == nil {
		f.accounts = accs
		opts := make([]string, len(accs))
		for i, a := range accs {
			opts[i] = a.Name
		}
		f.accountSelect.Options = opts
	}

	// Categories - Filter: Only Outcome (Egreso) and EXCLUDE system categories
	allCats, err := f.categoryService.GetAllCategories(context.Background())
	if err == nil {
		var outcomeCats []domain.Category
		var opts []string
		for _, c := range allCats {
			// Rule: Only Outcome AND not system categories (Ajuste/Anulación)
			if c.Type == domain.Outcome &&
				!strings.Contains(c.Name, "Ajuste") &&
				!strings.Contains(c.Name, "Anular") {
				outcomeCats = append(outcomeCats, c)
				opts = append(opts, c.Name)
			}
		}
		f.categories = outcomeCats
		f.categorySelect.Options = opts
	}
}

func (f *RecurringForm) onSave(ok bool) {
	if !ok {
		return
	}

	// Validation
	amount, err := strconv.ParseFloat(f.amountEntry.Text, 64)
	if err != nil {
		dialog.ShowError(fmt.Errorf("monto inválido"), f.window)
		return
	}

	date, err := time.Parse("2006-01-02", f.startDate.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("fecha inválida (use YYYY-MM-DD)"), f.window)
		return
	}

	var accID int
	for _, a := range f.accounts {
		if a.Name == f.accountSelect.Selected {
			accID = a.ID
			break
		}
	}
	if accID == 0 {
		dialog.ShowError(fmt.Errorf("seleccione una cuenta"), f.window)
		return
	}

	var catID int
	for _, c := range f.categories {
		if c.Name == f.categorySelect.Selected {
			catID = c.ID
			break
		}
	}
	if catID == 0 {
		dialog.ShowError(fmt.Errorf("seleccione una categoría"), f.window)
		return
	}

	// Create or Update
	isNew := f.item == nil
	if isNew {
		f.item = &domain.RecurringTransaction{}
		f.item.StartDate = date // Original start date
	}

	f.item.Description = f.descEntry.Text
	f.item.Amount = amount
	f.item.AccountID = accID
	f.item.CategoryID = catID

	// Map Spanish UI to Domain
	if f.intervalEntry.Selected == "SEMANAL" {
		f.item.Interval = domain.IntervalWeekly
	} else {
		f.item.Interval = domain.IntervalMonthly
	}

	f.item.NextRunDate = date
	f.item.IsActive = f.activeCheck.Checked

	ctx := context.Background()
	if isNew {
		err = f.service.Create(ctx, f.item)
	} else {
		err = f.service.Update(ctx, f.item)
	}

	if err != nil {
		dialog.ShowError(err, f.window)
		return
	}

	if f.OnSaved != nil {
		f.OnSaved()
	}
}
