# 👥 Gestión de Usuarios y Roles

Verith implementa un sistema de **Control de Acceso Basado en Roles (RBAC)** que garantiza que cada colaborador acceda únicamente a las herramientas que necesita para su trabajo.

---

## 🎭 Roles Disponibles

Al crear o editar un usuario, podrá elegir entre tres perfiles distintos:

### 🛡️ Administrador (Dueño / Gerente)
Es el usuario con control total sobre el sistema.
*   ✅ **Configuración SRI:** Único que puede cambiar datos legales y firmas.
*   ✅ **Gestión de Usuarios:** Puede crear, editar y eliminar otros accesos.
*   ✅ **Finanzas Totales:** Acceso a Dashboard, Cuentas, Reportes y Reconciliación.
*   ✅ **Auditoría:** Puede anular cualquier transacción.

### 📊 Supervisor (Contador / Auditor)
Enfocado en el control financiero y reporte de datos.
*   ✅ **Reportes:** Puede generar y exportar balances (PDF/CSV).
*   ✅ **Reconciliación:** Puede realizar cierres de caja y cuadres bancarios.
*   ✅ **Anulación:** Autorizado para anular facturas mediante Notas de Crédito.
*   🚫 **Restricción:** No puede ver la pestaña de Usuarios ni cambiar la configuración del SRI.

### 🛒 Cajero (Operativo / Ventas)
Interfaz simplificada para la operación diaria de facturación.
*   ✅ **Ventas:** Registro ágil de ingresos y facturación al SRI.
*   ✅ **Clientes:** Gestión de base de datos de contribuyentes.
*   🚫 **Privacidad:** No puede ver el Dashboard de ganancias ni los saldos de cuentas (`***`).
*   🚫 **Seguridad:** No puede anular facturas ni ver reportes globales.

---

## 🔐 Seguridad de Contraseñas

:::info Master Key
Si un administrador olvida su contraseña, el sistema permite recuperarla usando la **Clave de Licencia** comercial. Consulte la sección de [Licenciamiento](./licensing.md) para más detalles.
:::

:::warning Importante
Las contraseñas deben tener al menos **8 caracteres** para ser aceptadas por el sistema.
:::