1. issuers (Emisores)
   Función: Es el "Pasaporte Tributario" de tu cliente.

- ¿Qué guarda? Toda la información legal de quien usa el software: su RUC,
  Razón Social, Dirección Matriz, si es Contribuyente Especial, y la ruta a su
  firma electrónica (.p12).
- ¿Por qué es necesaria? Cada vez que se genera una factura, el SRI exige que
  estos datos vayan incrustados en el XML. Sin esta tabla, el sistema no sabría
  "quién" está facturando. Además, como es una app de escritorio, esta tabla
  permite que el usuario configure sus datos una sola vez y no los repita en
  cada venta.

1. emission_points (Puntos de Emisión)
   Función: Es el "Talonario de Facturas" digital.

- ¿Qué guarda? La configuración de las cajas o puntos de venta. Define el
  código del establecimiento (ej. 001) y el punto de emisión (ej. 002), y lo
  más importante: lleva la cuenta del secuencial actual (ej. "La última factura
  fue la 154, la siguiente es la 155").
- ¿Por qué es necesaria? El SRI es muy estricto con la numeración. No puedes
  saltarte números ni repetirlos. Esta tabla asegura que cada factura tenga un
  número único y consecutivo, bloqueando el registro para evitar duplicados si
  hubiera concurrencia.

1. tax_payers (Contribuyentes / Clientes)
   Función: Es la "Agenda de Contactos" fiscal.

- ¿Qué guarda? Los datos de las personas o empresas a las que les vendes:
  RUC/Cédula, Nombre, Correo Electrónico y Dirección.
- ¿Por qué es necesaria? Para emitir una factura válida, necesitas identificar
  al comprador. Guardar estos datos permite que, si un cliente vuelve, no
  tengas que pedirle sus datos de nuevo; solo buscas por cédula y el sistema
  rellena todo. Además, el campo email es vital para enviarles la factura
  electrónica automáticamente.

1. electronic_receipts (Comprobantes Electrónicos)
   Función: Es el "Archivo Legal" o la caja fuerte.

- ¿Qué guarda? Es la tabla más importante del módulo. Conecta una transacción
  de tu sistema (transaction_id) con el mundo del SRI. Guarda:
  - La Clave de Acceso (el ID único de 49 dígitos).
  - El XML firmado completo (la evidencia legal).
  - El Estado del SRI (¿Fue autorizada? ¿Fue rechazada?).
  - Los mensajes de error si falló.
- ¿Por qué es necesaria? La ley exige guardar estos documentos por 7 años. Si
  el SRI te hace una auditoría, de esta tabla salen las pruebas. También sirve
  para saber qué facturas se enviaron bien y cuáles hay que corregir y
  reenviar.

Resumen:

- issuers: Quién vende.
- tax_payers: A quién le vende.
- emission_points: Qué número de factura usa.
- electronic_receipts: La factura legal en sí (el XML y su estado).
