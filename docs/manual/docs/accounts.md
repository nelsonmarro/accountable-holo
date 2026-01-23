# 🏦 Gestión de Cuentas

Verith te permite manejar múltiples cuentas financieras (Caja Chica, Banco Pichincha, Banco Guayaquil, etc.) para tener una visión clara de dónde reside tu capital.

---

## ➕ Crear una Cuenta
1.  Ve a la pestaña **Cuentas**.
2.  Haz clic en **"Agregar Cuenta"**.
3.  Ingresa el nombre (ej: "Banco Principal"), el número de cuenta y el saldo inicial.

:::note Restricción de Rol
Por seguridad estructural, el botón de **"Agregar Cuenta"** solo es visible para usuarios con rol **Administrador**.
:::

## 📝 Editar una Cuenta
1.  En la tabla de cuentas, pulsa el icono de **Editar** (lápiz ✏️).
2.  Puedes actualizar el nombre o el número de la cuenta.

## 🗑️ Eliminar una Cuenta
1.  Haz clic en el icono de **Eliminar** (basurero 🗑️).
2.  **⚠️ Advertencia:** Solo los **Administradores** pueden eliminar cuentas. Esta acción es irreversible y requiere que la cuenta no tenga transacciones históricas para mantener la integridad de los balances.

---

## 📊 Ver Saldos
La tabla muestra el saldo actual calculado automáticamente. 

:::important Privacidad del Cajero
Los usuarios con rol **Cajero** pueden ver el listado de cuentas para registrar ventas, pero los montos de balance aparecerán ocultos o restringidos según la configuración de privacidad del sistema.
:::