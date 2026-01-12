package transaction

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// ItemsListManager maneja la visualización y edición de la lista de ítems en memoria.
type ItemsListManager struct {
	items     []domain.TransactionItem
	container *fyne.Container
	parent    fyne.Window
	onUpdate  func(items []domain.TransactionItem) // Callback cuando cambia la lista
}

func NewItemsListManager(parent fyne.Window, onUpdate func([]domain.TransactionItem)) *ItemsListManager {
	m := &ItemsListManager{
		items:    make([]domain.TransactionItem, 0),
		parent:   parent,
		onUpdate: onUpdate,
	}
	m.container = container.NewVBox()
	return m
}

func (m *ItemsListManager) GetContent() fyne.CanvasObject {
	header := container.NewGridWithColumns(4,
		widget.NewLabelWithStyle("Descripción", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Cant.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Precio", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Acción", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	addBtn := widget.NewButtonWithIcon("Agregar Ítem", theme.ContentAddIcon(), m.showAddItemDialog)
	addBtn.Importance = widget.SuccessImportance

	// Envolvemos la lista en un scroll por si hay muchos ítems
	scrollList := container.NewVScroll(m.container)
	scrollList.SetMinSize(fyne.NewSize(0, 200))

	return container.NewBorder(header, addBtn, nil, nil, scrollList)
}

func (m *ItemsListManager) SetItems(items []domain.TransactionItem) {
	m.items = items
	m.refreshList()
}

func (m *ItemsListManager) refreshList() {
	m.container.Objects = nil

	for i := range m.items {
		index := i
		item := m.items[i]

		desc := widget.NewLabel(item.Description)
		desc.Truncation = fyne.TextTruncateEllipsis

		qty := widget.NewLabel(fmt.Sprintf("%.2f", item.Quantity))
		qty.Alignment = fyne.TextAlignCenter

		price := widget.NewLabel(fmt.Sprintf("$%.2f", item.UnitPrice))
		price.Alignment = fyne.TextAlignTrailing

		deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			m.removeItem(index)
		})
		deleteBtn.Importance = widget.DangerImportance

		row := container.NewGridWithColumns(4, desc, qty, price, deleteBtn)
		m.container.Add(row)
	}

	m.container.Refresh()
	if m.onUpdate != nil {
		m.onUpdate(m.items)
	}
}

func (m *ItemsListManager) removeItem(index int) {
	m.items = append(m.items[:index], m.items[index+1:]...)
	m.refreshList()
}

func (m *ItemsListManager) showAddItemDialog() {
	d := NewItemDialog(m.parent, func(item domain.TransactionItem) {
		m.items = append(m.items, item)
		m.refreshList()
	})
	d.Show()
}
