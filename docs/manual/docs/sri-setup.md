# Configuración SRI

Antes de comenzar a facturar electrónicamente, es obligatorio configurar los datos de su empresa o perfil profesional. Estos datos son los que el SRI utiliza para validar sus comprobantes. 

:::info Asistente Inicial
Si es la primera vez que usa Verith, el sistema le mostrará automáticamente un **Asistente de Configuración Inicial** tras su primer inicio de sesión para facilitarle este proceso.
:::

:::caution Acceso Exclusivo
Solo los usuarios con rol de **Administrador** tienen permiso para ver y modificar esta configuración. Si usted es Cajero o Supervisor y necesita realizar cambios, contacte al administrador del sistema.
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



---



## 🚀 Migración desde otro Sistema



Si usted ya emitía facturas electrónicas con otro software y desea empezar a usar **Verith** manteniendo su numeración actual, debe seguir este proceso de migración para evitar rechazos del SRI por secuenciales duplicados.



### 1. Preparación

Antes de configurar Verith, emita su última factura en su sistema anterior y anote el número (ejemplo: `001-001-000001500`).



### 2. Configurar el Emisor

En esta pestaña de **Configuración SRI**, asegúrese de haber guardado sus **Datos Legales** y **Códigos de Emisión** (Establecimiento y Punto de Emisión) antes de proceder al ajuste de secuenciales.



### 3. Ajuste de Secuenciales

Haga clic en el botón **"MIGRAR / AJUSTAR SECUENCIALES"**. Se abrirá un diálogo con los registros de sus puntos de emisión.



Para cada tipo de documento (Factura o Nota de Crédito), haga clic en el icono de editar y configure los siguientes campos:



*   **Secuencial Actual:** Ingrese el número del **último documento emitido con éxito** en su sistema anterior.

    *   *Ejemplo:* Si su última factura fue la **1500**, ingrese `1500`. Verith generará la siguiente como la `1501`.

*   **Secuencial Inicial:** Ingrese el número con el que **desea que Verith empiece su historial**.

    *   *Ejemplo:* Ingrese `1501`. Este campo es solo para referencia de auditoría interna.



:::danger Advertencia de Seguridad

Al guardar un cambio en el secuencial, Verith le solicitará una confirmación. **Reducir el número secuencial** es altamente peligroso, ya que el SRI rechazará cualquier factura con un número que ya haya sido autorizado previamente.

:::



### 4. Verificación

Una vez guardado, cierre el diálogo y proceda a realizar su primera venta. Verith tomará automáticamente el "Secuencial Actual" que usted ingresó, le sumará 1, y emitirá el comprobante con la numeración correcta.
