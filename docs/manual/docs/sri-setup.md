# Configuración SRI

Antes de comenzar a facturar electrónicamente, es obligatorio configurar los datos de su empresa o perfil profesional. Estos datos son los que el SRI utiliza para validar sus comprobantes. 

:::info Asistente Inicial
Si es la primera vez que usa Verith, el sistema le mostrará automáticamente un **Asistente de Configuración Inicial** tras su primer inicio de sesión para facilitarle este proceso.
:::

---

## 1. Información Legal y Matriz
Esta sección contiene los datos base que aparecerán en la cabecera de todos sus documentos.

*   **RUC:** Registro Único de Contribuyentes (13 dígitos).
*   **Razón Social:** Nombre legal completo del contribuyente.
*   **Nombre Comercial:** Nombre de fantasía del negocio (si aplica).
*   **Dir. Matriz:** Dirección de la oficina principal registrada en el RUC.
*   **Dir. Establecimiento:** Dirección física del punto de venta desde donde emite actualmente.

## 2. Configuración de Facturación
Define los parámetros operativos para la generación de archivos XML.

*   **Cod. Establecimiento:** Código de 3 dígitos (ej: `001`) asignado por el SRI a su local.
*   **Punto de Emisión:** Código de 3 dígitos (ej: `001`) que identifica la caja o terminal actual.
*   **Régimen RIMPE:** Selección del régimen tributario actual (Ninguno, Negocio Popular o Emprendedor).
*   **Ambiente SRI:** 
    *   *Pruebas:* Para realizar tests de conexión sin validez legal.
    *   *Producción:* Para emisión real de documentos con validez tributaria.
*   **Nro. Resolución:** Número de resolución administrativa (requerido si es Contribuyente Especial o Agente de Retención).
*   **Obligado a llevar contabilidad:** Casilla de verificación según su obligación legal.

## 3. Firma Electrónica
Es el componente que otorga validez legal a sus documentos electrónicos.

*   **Archivo Firma:** Seleccione su archivo de certificado digital con extensión `.p12` o `.pfx`.
*   **Contraseña:** La clave de seguridad de su firma electrónica. 
    *   *Seguridad:* Verith almacena esta clave de forma encriptada en el llavero seguro de su sistema operativo.

## 4. Identidad Visual
Configuración estética para la representación impresa (PDF).

*   **Logo Empresa:** Cargue una imagen en formato `.png`, `.jpg` o `.jpeg`. Este logo se visualizará en la parte superior de sus facturas y notas de crédito enviadas por correo.

---

### Guardar Cambios
Una vez completados los datos, pulse el botón **"Guardar Cambios"**. Verith verificará la integridad de los datos y guardará su configuración de forma segura.
