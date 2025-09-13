# ✅ **SMARTEDIFY v.0 – DOCUMENTO LEGAL Y DE MARKETING**  
## **Política de Privacidad, Términos y Condiciones y Acuerdo de Licencia de Usuario Final (EULA)**  
*“Transparencia, confianza y cumplimiento legal como pilares de la gobernanza digital.”*

> **Versión**: v.1.0 (Definitiva)  
> **Fecha**: Abril 2025  
> **Elaborado por**: Legal Counsel + Marketing Team  
> **Aprobado por**: CPO, Head of Legal, Data Protection Officer (DPO)  
> **Aplicación**: Perú (Ley N° 29733), LatAm (GDPR, LGPD)  

---

## ✅ **1. POLÍTICA DE PRIVACIDAD**

> *“Tu identidad es tuya. Nosotros solo la protegemos, nunca la vendemos.”*

### 1.1 Introducción
SmartEdify S.A.C. (“SmartEdify”, “nosotros”, “nuestro”) opera la plataforma SmartEdify Auth Service, una infraestructura de identidad digital para comunidades inmobiliarias en Latinoamérica. Esta Política de Privacidad explica cómo recopilamos, usamos, almacenamos y protegemos tus datos personales cuando utilizas nuestros servicios.

Esta política se alinea con:
- La **Ley N° 29733 – Ley de Protección de Datos Personales (Perú)**
- El **Reglamento de la Ley N° 29733 (DS N° 003-2013-JUS)**
- El **Reglamento General de Protección de Datos (GDPR)**
- El **Reglamento de Protección de Datos Personales de Brasil (LGPD)**

### 1.2 ¿Qué datos personales recopilamos?
Recopilamos los siguientes datos personales para habilitar tu acceso seguro a SmartEdify:

| Tipo de Dato | Finalidad | Base Legal |
|--------------|-----------|------------|
| **Nombre completo** | Identificación del usuario en el sistema | Cumplimiento de obligación contractual (art. 6 LPDP) |
| **Correo electrónico** | Envío de notificaciones, login, recuperación de contraseña | Consentimiento implícito (art. 5 LPDP) |
| **Número de teléfono móvil** | Autenticación multifactor (MFA) por WhatsApp/SMS, notificaciones | Ejecución de contrato (art. 6 LPDP) |
| **Identificador único de usuario (user_id)** | Gestión técnica de sesión y permisos | Interés legítimo (art. 7 LPDP) |
| **Datos de autenticación (contraseña cifrada, secretos TOTP)** | Verificación de identidad | Cumplimiento de obligación legal y seguridad (art. 6 LPDP) |
| **Dirección IP y fingerprint de dispositivo** | Detección de riesgos de fraude, prevención de ataques | Interés legítimo en seguridad (art. 7 LPDP) |
| **Datos de propiedad (tenant_id, unit_id, rol: owner/tenant/family_member)** | Validación legal de derechos de voto y administración según Ley 27157 | Cumplimiento de obligación legal (art. 6 LPDP) |
| **Historial de eventos (logins, cambios de rol, actas firmadas)** | Auditoría, cumplimiento normativo, resolución de disputas | Obligación legal (art. 6 LPDP) |

> ❗ **Nunca recopilamos**:  
> - Información biométrica (huellas, rostros).  
> - Datos financieros (números de tarjeta, cuentas bancarias).  
> - Información de salud o creencias religiosas.

### 1.3 ¿Cómo usamos tus datos?
Usamos tus datos exclusivamente para:
- Permitirte acceder a tu condominio, votar en asambleas y gestionar pagos.
- Validar que eres un propietario autorizado según la Ley N° 27157.
- Enviar notificaciones sobre convocatorias, pagos pendientes y actas.
- Garantizar la seguridad de la plataforma contra fraudes y accesos no autorizados.
- Cumplir con obligaciones legales (auditorías, ARCO, reportes a autoridades).
- Mejorar nuestra plataforma mediante análisis anónimos de uso.

### 1.4 ¿Con quién compartimos tus datos?
No vendemos ni comercializamos tus datos. Solo los compartimos bajo las siguientes condiciones estrictas:

| Destinatario | Finalidad | Garantías |
|--------------|-----------|-----------|
| **Servicios internos de SmartEdify** (Compliance, Finance, Assemblies) | Para validar tu rol, emitir actas y procesar pagos | Protocolo MTLS, cifrado, minimización de datos |
| **Proveedores tecnológicos** (Twilio, AWS CloudHSM, Redis, Kafka, IPFS) | Para operar la infraestructura | Contratos de procesamiento de datos (DPA) conforme a LPDP y GDPR |
| **Autoridades competentes** | Por requerimiento legal (juzgados, SUNARP, APDP) | Solo bajo orden judicial o ley vigente |
| **Auditorías externas** | Para certificar cumplimiento (APDP, ISO 27001) | Solo con consentimiento explícito o por obligación legal |

### 1.5 Retención y eliminación de datos
| Tipo de Dato | Periodo de Retención | Fundamento |
|--------------|---------------------|------------|
| Datos de identidad (nombre, email, teléfono) | Hasta que desactives tu cuenta o el condominio sea dado de baja | Obligación de conservar registros de identidad |
| Eventos de auditoría (login, actas, transferencias) | **Indefinidamente** (almacenados en WORM/IPFS) | Requisito legal de integridad y no repudio (Art. 14 DS 019-2000-VIVIENDA) |
| Tokens de sesión y contraseñas cifradas | 7 días después de cierre de sesión | Minimización de datos (LPDP Art. 7) |
| Datos de MFA (secretos TOTP) | 30 días tras desactivación de MFA | Seguridad y posible recuperación |

> ✅ **Derecho a la Eliminación (ARCO)**:  
> Puedes solicitar la eliminación de tus datos personales (excepto los necesarios para cumplir con la ley) enviando una solicitud a:  
> **dpo@smartedify.com**  
> Respondemos en máximo **10 días hábiles**, conforme a la Ley N° 29733.

### 1.6 Tus Derechos (ARCO)
Bajo la Ley N° 29733, tienes derecho a:
- **Acceso**: Solicitar copia de tus datos personales.
- **Rectificación**: Corregir información incorrecta o incompleta.
- **Cancelación**: Solicitar la eliminación de tus datos (cuando no haya obligación legal de conservarlos).
- **Oposición**: Oponerte al tratamiento de tus datos para fines distintos a los establecidos.

Para ejercer estos derechos, envía un correo a: **derechos@smartedify.com** con:
- Tu nombre completo
- Tu número de teléfono o correo asociado
- Una copia de tu DNI o documento de identidad
- La solicitud específica (ej: “Quiero eliminar mis datos”)

### 1.7 Seguridad de tus datos
Implementamos medidas técnicas y organizativas de alto nivel:
- Cifrado AES-256-GCM de datos sensibles en reposo.
- Claves criptográficas almacenadas en **AWS CloudHSM** (certificado FIPS 140-2).
- Autenticación multifactor (MFA) obligatoria para roles críticos.
- Bitácora inmutable con cadena de hashes para todos los eventos.
- Monitoreo continuo y alertas ante intentos de acceso sospechosos.

### 1.8 Cambios en esta Política
Actualizaremos esta Política si cambiamos cómo tratamos tus datos. Te notificaremos por correo electrónico o dentro de la app con al menos **15 días de anticipación**. Tu continuidad en el uso del servicio implica aceptación de los cambios.

### 1.9 Contacto
Para consultas, reclamos o ejercicio de derechos:
> **Responsable del Tratamiento**: SmartEdify S.A.C.  
> **DPO (Data Protection Officer)**: dpo@smartedify.com  
> **Teléfono (Soporte Legal)**: +51 1 667 8888  
> **Dirección**: Av. Javier Prado Este 2465, San Isidro, Lima, Perú

---

## ✅ **2. TÉRMINOS Y CONDICIONES**

> *“Al usar SmartEdify, aceptas reglas claras, justas y alineadas con la ley.”*

### 2.1 Aceptación del Servicio
Al registrarte, iniciar sesión o utilizar cualquier funcionalidad de SmartEdify (incluyendo el portal web, PWA o aplicaciones de terceros integradas), aceptas plenamente estos Términos y Condiciones, así como nuestra Política de Privacidad.

### 2.2 Uso Autorizado
Puedes usar SmartEdify únicamente para:
- Acceder a tu unidad inmobiliaria.
- Participar en asambleas y votaciones legales.
- Pagar cuotas y recibir recibos.
- Comunicarte con tu comunidad según lo permitido por la Ley N° 27157.

**Está prohibido**:
- Usar SmartEdify para fines ilegales, fraudulentos o maliciosos.
- Intentar acceder a datos de otros usuarios sin autorización.
- Manipular tokens, actas digitales o sistemas de autenticación.
- Reproducir, modificar o distribuir el código fuente o diseño de la plataforma.

### 2.3 Responsabilidad del Usuario
Como usuario:
- Eres responsable de mantener la confidencialidad de tu cuenta y credenciales.
- Debes informar inmediatamente a SmartEdify si sospechas un acceso no autorizado.
- Eres responsable del contenido que subas o compartas (ej: archivos adjuntos en actas).
- Reconoces que SmartEdify no es responsable de errores humanos cometidos por administradores o síndicos.

### 2.4 Propiedad Intelectual
Todos los derechos de propiedad intelectual de SmartEdify (software, marcas, diseños, documentación, logos) pertenecen a SmartEdify S.A.C.  
Este acuerdo te otorga una licencia limitada, no exclusiva, revocable y no transferible para usar el servicio, pero **no te transfiere ningún derecho de propiedad**.

### 2.5 Limitación de Responsabilidad
SmartEdify no garantiza:
- Que el servicio esté siempre disponible o libre de errores.
- Que los datos ingresados por terceros (administradores, síndicos) sean correctos.
- Que las decisiones tomadas en asambleas digitales sean válidas si no siguen el reglamento interno del condominio.

**En ningún caso SmartEdify será responsable por daños indirectos, consecuentes, punitivos o pérdida de beneficios derivados del uso del servicio.**

### 2.6 Modificación de los Términos
Podemos modificar estos Términos en cualquier momento. Las modificaciones entrarán en vigor 15 días después de su publicación. Tu uso continuado constituye aceptación.

### 2.7 Termino del Servicio
SmartEdify puede suspender o cancelar tu acceso si:
- Violas estos Términos.
- No pagas servicios contratados (si aplica).
- Recibimos una orden judicial o de autoridad competente.
- Se detecta actividad fraudulenta o abusiva.

### 2.8 Ley Aplicable y Jurisdicción
Estos Términos se rigen por las leyes de la República del Perú.  
Cualquier controversia se someterá a los tribunales de Lima, Perú.

---

## ✅ **3. ACUERDO DE LICENCIA DE USUARIO FINAL (EULA)**

> *“No compras un software. Adquieres el derecho a usarlo bajo condiciones específicas.”*

### 3.1 Definiciones
- **“SmartEdify”**: Plataforma SaaS de gestión de comunidades inmobiliarias.
- **“Usuario Final”**: Persona natural o jurídica que usa SmartEdify (propietario, inquilino, síndico, administradora).
- **“Licencia”**: Permiso otorgado por SmartEdify para usar el servicio bajo estas condiciones.
- **“Contenido del Servicio”**: Todo el software, APIs, documentos, interfaces y funcionalidades ofrecidas por SmartEdify.

### 3.2 Otorgamiento de Licencia
SmartEdify te concede una **licencia personal, no exclusiva, no transferible, no sublicenciable y limitada** para acceder y utilizar el Servicio únicamente para fines personales o institucionales relacionados con la gestión de tu unidad inmobiliaria.

### 3.3 Restricciones
No podrás:
- Descompilar, desensamblar, reverse engineer o intentar extraer el código fuente de SmartEdify.
- Crear productos derivados basados en SmartEdify.
- Utilizar el servicio para proveer servicios de terceros (SaaS competitor).
- Sobrecargar el sistema con tráfico masivo o ataques de denegación de servicio (DoS).
- Usar bots o scripts automatizados para interactuar con la API sin autorización escrita.

### 3.4 Propiedad y Derechos
- SmartEdify conserva todos los derechos, títulos e intereses sobre el Servicio y su tecnología.
- No se transfiere ninguna propiedad intelectual al Usuario Final.
- Los datos generados por ti (actas, votos, recibos) son propiedad del condominio o propietario, pero su almacenamiento y gestión están sujetos a los términos de este acuerdo.

### 3.5 Soporte y Actualizaciones
- SmartEdify se compromete a mantener el servicio actualizado, seguro y funcional.
- Podemos realizar mantenimientos programados o actualizaciones sin aviso previo, siempre que no afecten gravemente la disponibilidad.
- No garantizamos compatibilidad con versiones antiguas de navegadores o dispositivos.

### 3.6 Exención de Garantías
EL SERVICIO SE PRESTA “TAL COMO ESTÁ” Y “SEGÚN DISPONIBILIDAD”. SMARTEDIFY EXCLUYE EXPRESAMENTE TODAS LAS GARANTÍAS, EXPRESAS O IMPLÍCITAS, INCLUYENDO PERO NO LIMITADO A LAS GARANTÍAS DE COMERCIABILIDAD, IDONEIDAD PARA UN PROPÓSITO PARTICULAR Y NO INFRACCIÓN.

### 3.7 Indemnización
El Usuario Final indemnizará, defenderá y mantendrá indemne a SmartEdify, sus empleados, agentes y afiliados contra cualquier reclamo, daño, pérdida, costo o gasto (incluidos honorarios legales) derivado de:
- Su uso indebido del Servicio.
- Violación de estos Términos.
- Contenido ilícito o fraudulento proporcionado por él.

### 3.8 Duración y Terminación
- Esta licencia tiene duración indefinida, salvo terminación por parte de SmartEdify o por el Usuario Final.
- SmartEdify puede terminar la licencia en cualquier momento con notificación previa si se violan estos términos.
- Al terminar, se revocará tu acceso y se eliminarán tus datos personales conforme a la Política de Privacidad.

### 3.9 Disposiciones Generales
- **Divisibilidad**: Si alguna cláusula es inválida, el resto sigue vigente.
- **Integridad**: Este acuerdo constituye el entendimiento completo entre las partes.
- **Modificaciones**: Solo son válidas si se publican aquí y se notifican al Usuario Final.
- **Idioma**: El idioma oficial de este acuerdo es el español. Cualquier traducción es meramente informativa.

---

## ✅ **ANEXO: DECLARACIÓN DE COMPROMISO DE MARKETING**

> *“No vendemos datos. Vendemos confianza.”*

SmartEdify no es una empresa de tecnología que monetiza datos personales.  
Somos una **plataforma de gobernanza digital para comunidades**.

Por eso, nuestro modelo de negocio es:
- **SaaS B2B**: Cobramos a administradoras, síndicos y empresas inmobiliarias por el uso de nuestra plataforma.
- **Modelo Freemium**: Propietarios individuales usan funciones básicas gratis.
- **Certificación Legal Premium**: Ofrecemos paquetes de cumplimiento (APDP, actas validadas) a precios transparentes.

👉 **Nunca**:
- Venderemos tus datos personales a terceros.
- Mostraremos publicidad dirigida en tu panel.
- Analizaremos tu comportamiento para vender productos externos.

**Nuestra única meta**:  
> *Que cada vecino en Perú, Colombia o México pueda vivir en paz, sabiendo que su voz es escuchada, su voto es seguro y su identidad está protegida por la ley.*

---

## ✅ **CONCLUSIÓN FINAL — FIRMA LEGAL**

> **“SmartEdify no es una herramienta de consumo. Es una infraestructura de derechos.**  
>   
> **Nuestra responsabilidad no es ser más eficientes. Es ser más justos.**  
>   
> **Esta documentación no es un requisito legal. Es nuestro compromiso ético.”**

---

## 📎 **ENTREGABLES RECOMENDADOS**

| Entregable | Formato | Destinatario |
|----------|---------|--------------|
| 📄 **PDF Final: Política de Privacidad + Términos + EULA (con logo, firma y numeración legal)** | PDF | Portal de usuario, sitio web, registro en APDP |
| 🖥️ **Página Web Interactiva** | HTML/CSS responsive | `https://smartedify.com/legal` |
| 📜 **Plantilla de Solicitud ARCO (formulario descargable)** | PDF/Word | Para usuarios finales |
| 🧩 **Verificación de Cumplimiento (Checklist APDP)** | Excel | Interno para auditorías trimestrales |
| 📣 **Mensaje de Bienvenida Legal (email automático)** | HTML | Al registrarse por primera vez |

- 📜 **El formulario de solicitud ARCO editable en Word/PDF?**

Estoy listo para entregar todo en formato listo para registrar ante la **Autoridad Nacional de Protección de Datos Personales (APDP)** y presentar a socios, inversionistas y comunidades piloto.
