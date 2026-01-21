---
sidebar_position: 2
---

# Instalación y Configuración

## 1. Instalación en Windows

1.  Descarga el instalador oficial `Verith_Setup.exe`.
2.  Ejecuta el archivo. Es posible que Windows muestre una advertencia de "Editor desconocido"; haz clic en **"Más información"** y luego en **"Ejecutar de todas formas"**.
3.  El asistente instalará la aplicación en `Archivos de Programa` y configurará automáticamente el motor de base de datos **PostgreSQL**.
4.  Al finalizar, marca la casilla **"Ejecutar Verith"**.

:::tip Nota sobre permisos
Verith está diseñado para ejecutarse sin privilegios de administrador después de la instalación. Los datos y logs se guardan de forma segura en carpetas de usuario.
:::

---

## 2. Onboarding: Registro Inicial

Al abrir la aplicación por primera vez, Verith detectará que no existen usuarios y te presentará la pantalla de **Registro de Administrador**.

1.  **Nombre de Usuario:** Elige un nombre para tu cuenta (mínimo 3 caracteres).
2.  **Contraseña:** Define una clave segura (mínimo 8 caracteres).
3.  **Nombre y Apellido:** Ingresa tus datos reales para los registros de auditoría.
4.  Haz clic en **"Crear Cuenta"**.

Una vez creada, serás redirigido automáticamente a la pantalla de Inicio de Sesión.

![Onboarding - Registro inicial](https://via.placeholder.com/800x450?text=Captura+Pantalla+Registro+Inicial)

---

## 3. Configuración Inicial SRI

Inmediatamente después de tu primer inicio de sesión como administrador, aparecerá el **Asistente de Configuración Inicial**. Este paso es obligatorio para activar la facturación electrónica.

Deberás completar cuatro secciones clave:

1.  **Información Legal:** RUC, Razón Social y Direcciones tal como constan en el SRI.
2.  **Configuración de Emisión:** Selección de Ambiente (Pruebas/Producción) y Régimen RIMPE.
3.  **Firma Electrónica:** Carga de tu archivo `.p12` y su contraseña.
4.  **Identidad Visual:** Carga del logo de tu negocio para los PDFs.

:::info ¿Puedo omitir este paso?
Si cierras el asistente sin guardar, podrás volver a él desde la pestaña **"Configuración SRI"** en cualquier momento, pero no podrás emitir facturas hasta que los datos estén completos.
:::