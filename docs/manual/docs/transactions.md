# 📝 Transacciones Diarias

El registro de transacciones es la actividad principal en Verith. Aquí documentas cada evento financiero de tu negocio.

---

## ➕ Registrar una Transacción
1.  Haz clic en el botón **"Agregar Transacción"** en la barra superior.
2.  Completa el formulario:
    *   **Monto:** El valor total de la operación.
    *   **Fecha:** Por defecto es hoy, pero puedes seleccionar fechas pasadas.
    *   **Descripción:** Detalle claro del movimiento.
    *   **Cuenta:** De qué cuenta sale o a cuál entra el dinero.
    *   **Categoría:** Clasifica el movimiento.
    *   **Cliente (Opcional):** Si es un ingreso, selecciona un cliente para generar la factura electrónica.
3.  Pulsa **"Guardar"**.

## 📝 Editar una Transacción
1.  Busca la transacción en la tabla principal.
2.  Haz clic en el icono de **Editar** (lápiz ✏️).
3.  Modifica los datos necesarios y guarda.
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
