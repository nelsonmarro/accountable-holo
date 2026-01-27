# 📝 Transacciones Diarias

El registro de transacciones es la actividad principal en Verith. Aquí documentas cada evento financiero de tu negocio.

---

## ➕ Registrar Movimientos

Verith diferencia entre tus ventas y tus gastos para ofrecerte la mejor experiencia en cada caso:

### 1. Registrar Venta (Ingresos)
Ideal para tu facturación diaria y movimientos que requieren detalle.
*   Haz clic en el botón **"Venta"** (Verde).
*   **Gestión de Ítems:** Permite agregar múltiples productos o servicios.
*   **IVA Automático:** Si configuraste un **IVA Predeterminado**, el sistema lo seleccionará por ti.
*   **Facturación SRI:** Selecciona un cliente para emitir el comprobante electrónico legal.

### 2. Registrar Gasto (Egresos)
Un flujo simplificado para tus compras y pagos rápidos.
*   Haz clic en el botón **"Gasto"** (Naranja).
*   **Registro Rápido:** Solo necesitas ingresar el monto total, categoría y descripción.
*   **Sin Impuestos:** Los gastos se registran como valores totales (Tarifa 0% interna) para agilizar la contabilidad.
*   **Recurrencia:** Puedes marcarlo como recurrente directamente desde aquí.

---

## ⚙️ Menú de Herramientas
Para mantener la interfaz limpia, hemos agrupado las funciones avanzadas en el icono de engranaje (**Herramientas**):
*   📊 **Reportes:** Generación de PDF/CSV.
*   🔄 **Reconciliar:** Cuadre de caja y bancos.
*   🕒 **Recurrentes:** Acceso al gestor de automatizaciones.
*   📤 **Cola SRI:** Ver estado de documentos pendientes.

---

## 📝 Editar una Transacción
1.  Busca la transacción en la tabla principal.
2.  Haz clic en el icono de **Editar** (lápiz ✏️).
3.  **Inteligencia de Diálogo:** El sistema abrirá automáticamente el diálogo simplificado si es un gasto, o el detallado si es una venta.
    *   **⚠️ Restricción:** Las transacciones que ya han sido **Autorizadas por el SRI** no pueden editarse por motivos legales. Si hubo un error, deberás anularla.

## 🚫 Anular (Eliminar) una Transacción
En Verith, para mantener un historial contable íntegro (pista de auditoría), **no se borran** las transacciones permanentemente. En su lugar, se **anulan**.

1.  Haz clic en el botón **Anular** (icono ❌ rojo) en la fila de la transacción.
2.  El sistema realizará lo siguiente:
    *   Creará una transacción de contrapartida (opuesta) para revertir el saldo.
    *   Marcará la original como "Anulada" (fila en color rojo suave).
    *   Si era una factura autorizada, emitirá automáticamente una **Nota de Crédito** al SRI.

---

:::info Adjuntos
No olvides adjuntar fotos de tus recibos o PDFs de transferencias en el campo de **Adjunto** al crear o editar para tener un respaldo digital.
:::
