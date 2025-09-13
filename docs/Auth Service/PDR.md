
# ✅ **SMARTEDIFY v.0 – DOCUMENTO DE PLANIFICACIÓN Y REQUISITOS**  
## **Auth Service — Infraestructura de Identidad Digital para Comunidades Legales en LatAm**

> **Versión**: v.1.0 (Definitiva)  
> **Fecha**: Abril 2025  
> **Autor**: Software Architect, SmartEdify  
> **Aprobado por**: CPO, Head of Security, Legal Counsel, Engineering Lead  

---

## ✅ **1. Introducción y Visión**

### ¿Qué estamos construyendo?
Estamos construyendo **Auth Service**, el microservicio central de identidad digital de SmartEdify: una plataforma SaaS multi-tenant que permite a comunidades inmobiliarias en Latinoamérica gestionar sus asambleas, pagos y operaciones mediante una experiencia de acceso **sin contraseñas, legalmente válida y diseñada para personas reales — no para ingenieros**.

### ¿Por qué lo hacemos?
**El problema real:**  
En LatAm, los condominios viven en un limbo entre lo analógico y lo digital. Los propietarios olvidan contraseñas, los síndicos usan Excel, las asambleas se hacen con papeles y firmas manuscritas, y las leyes locales (como la Ley N° 27157 en Perú) exigen que solo los propietarios puedan votar o ser presidentes — pero nadie verifica quién es quién en la app.  

Los sistemas actuales (CondoControl, MiCondominio, etc.) son complejos, caros y **ignoran la ley**. No hay confianza. Nadie sabe si quien vota realmente es dueño.  

**Nuestra solución:**  
Un servicio de autenticación que:  
- **Elimina contraseñas** usando WhatsApp, FIDO2 o biometría.  
- **Garantiza legalmente** que solo los propietarios pueden tener derechos.  
- **Vincula identidad digital con propiedad física** (unidad → usuario → tenant).  
- **Cumple con la ley peruana y latinoamericana sin que el usuario tenga que leerla**.  

> 🔥 **Visión**:  
> *“Que cada vecino en Perú, Colombia o México pueda acceder a su condominio, votar en su asamblea y pagar su cuota… respondiendo ‘SÍ’ por WhatsApp, sin recordar nada.”*

---

## ✅ **2. Objetivos y Metas**

### 🎯 Objetivos de Negocio
| Objetivo | Meta | Plazo |
|---------|------|-------|
| Lanzar MVP en Perú con primeras 3 comunidades piloto | 3 condominios activos con 100+ usuarios | Mes 3 |
| Alcanzar 1,000 usuarios activos mensuales (MAU) | 1,000 usuarios únicos logueados/mes | Mes 6 |
| Convertir 15% de usuarios en “usuarios leales” | NPS ≥ 45 | Mes 6 |
| Posicionar a SmartEdify como la única plataforma legalmente certificada en LatAm | Certificación APDP (Perú) obtenida | Mes 5 |

### 🚀 Objetivos de Producto
| Objetivo | Meta | Plazo |
|----------|------|-------|
| Reducir el tiempo de inicio de sesión a menos de 8 segundos | 90% de los usuarios logueados en ≤ 8s | Mes 3 |
| Eliminar el 95% de tickets de “olvidé mi contraseña” | De 40% a <2% del total de soporte | Mes 6 |
| Lograr que el 85% de los logins sean sin contraseña | WhatsApp + FIDO2 como método principal | Mes 6 |
| Garantizar que el 100% de las actas digitales sean válidas ante autoridades | 100% de actas generadas verificables con QR | Mes 3 |

---

## ✅ **3. Métricas de Éxito (KPIs)**

| Tipo | Métrica | Meta | Frecuencia |
|------|--------|------|------------|
| **Adopción** | Tasa de login exitoso (primer intento) | ≥ 85% | Diaria |
| **Engagement** | Usuarios activos semanales (WAU) | ≥ 70% de MAU | Semanal |
| **Retención** | Churn Rate (usuarios que abandonan) | ≤ 5% mensual | Mensual |
| **Satisfacción** | Net Promoter Score (NPS) | ≥ 45 | Trimestral |
| **Legalidad** | % de actas validadas por jurisdicción | 100% | Diaria |
| **Eficiencia** | Tiempo promedio de login | ≤ 8 segundos | Diaria |
| **Costo** | Costo por usuario activo (CPA) | ≤ $0.80 | Mensual |

> 💡 **Regla de oro**:  
> Si más del 15% de los usuarios necesita ayuda para iniciar sesión, **hemos fallado**.

---

## ✅ **4. Perfiles de Usuario (User Personas)**

### 👨‍👩‍👧‍👦 **Juan Pérez — Propietario Mayor (68 años)**
- **Quién es**: Dueño de un departamento en Lima. Usa WhatsApp todos los días. No sabe qué es un “JWT”.  
- **Dolor**: Olvida contraseñas. Le da miedo hacer clic en botones desconocidos.  
- **Meta**: Ver su recibo y votar en la asamblea sin tener que llamar al administrador.  
- **Comportamiento clave**:  
  - Responde “SÍ” a mensajes de WhatsApp.  
  - Nunca descarga apps nuevas.  
  - Confía en lo que ve en su pantalla de celular.  
- **Frase típica**:  
  > *“¿Me mandan un mensaje y yo digo ‘Sí’? Entonces sí.”*

### 🏢 **María González — Síndica (55 años)**
- **Quién es**: Administradora de 3 condominios. Usa Excel. No tiene equipo de IT.  
- **Dolor**: Tiene 200 cuentas que manejar. Cada mes pierde 3 días cargando datos. Teme cometer errores legales.  
- **Meta**: Subir 100 usuarios en 5 minutos, convocar una asamblea con un click, y tener pruebas legales de que todo está bien.  
- **Comportamiento clave**:  
  - Necesita que todo sea “fácil, rápido y seguro”.  
  - No quiere aprender software nuevo. Quiere que el software aprenda de ella.  
  - Valora más el sello “Cumple con la Ley” que las animaciones.  
- **Frase típica**:  
  > *“Si esto me evita que me multen por una asamblea mal hecha, vale cualquier cosa.”*

---

## ✅ **5. Requisitos de Funcionalidades (MVP)**

### ✅ **Feature 1: Login por WhatsApp como método principal**

#### 📜 User Story  
> *Como Juan Pérez (propietario), quiero iniciar sesión en SmartEdify respondiendo “SÍ” a un mensaje de WhatsApp, para poder ver mi cuota y votar sin recordar ninguna contraseña.*

#### ✅ Criterios de Aceptación
- [ ] El sistema envía un OTP por WhatsApp cuando se hace clic en “Iniciar con WhatsApp”.  
- [ ] El usuario responde “SÍ”, “NO” o “ABSTENCIÓN” en el chat.  
- [ ] Al responder “SÍ”, se genera un JWT válido y se redirige automáticamente al dashboard.  
- [ ] No se muestra ningún campo de texto para email o contraseña.  
- [ ] Se emite evento `user.login.success` con canal = “whatsapp”.  
- [ ] Si el número no está registrado, se redirige a flujo de registro automático.  
- [ ] Fallo en 3 intentos → bloqueo temporal + notificación por SMS.  

---

### ✅ **Feature 2: Asignación legal de presidente (solo propietarios)**

#### 📜 User Story  
> *Como María González (síndica), quiero designar a un propietario como presidente del condominio, para que pueda convocar asambleas sin riesgo de que alguien no dueño tome decisiones legales.*

#### ✅ Criterios de Aceptación
- [ ] Solo los usuarios con rol `owner` en alguna unidad del tenant aparecen en la lista de candidatos.  
- [ ] Al seleccionar un propietario, se envía un link por WhatsApp: *“[Nombre] te ha designado presidente. Haz clic para aceptar.”*  
- [ ] El propietario debe aceptar el rol respondiendo “SÍ” por WhatsApp y activando MFA (WhatsApp o FIDO2).  
- [ ] Al aceptar, se genera una **acta digital firmada** con hash en IPFS y QR de verificación.  
- [ ] La acta incluye: nombre del presidente, unidad, fecha, firma digital y texto legal: *“Según la Ley N° 27157”*.  
- [ ] El antiguo presidente pierde el rol automáticamente.  
- [ ] Se emite evento `president.transfer.completed` con documentación vinculada.  

---

### ✅ **Feature 3: Actas digitales verificables (con firma legal)**

#### 📜 User Story  
> *Como Juan Pérez, quiero ver una acta de asamblea y saber que es legalmente válida, sin necesidad de imprimir ni buscar firmas físicas.*

#### ✅ Criterios de Aceptación
- [ ] Cada acta generada (elección, transferencia, aprobación de gastos) se exporta como PDF.  
- [ ] El PDF incluye:  
  - Firma digital RSA generada desde HSM.  
  - Hash único almacenado en IPFS.  
  - QR visible que lleva a `verify.smartedify.dev/acta/[id]`.  
- [ ] Al escanear el QR, se muestra:  
  - “Firma válida” / “Firma inválida”  
  - “Emitida por SmartEdify. Cumple con la Ley N° 27157.”  
- [ ] El hash y la firma están vinculados a un evento auditado en bitácora inmutable.  
- [ ] El archivo PDF es descargable y compatible con SUNARP.  
- [ ] Se emite evento `acta.signed` con IPFS CID y metadata.  

---

### ✅ **Feature 4: Acceso dinámico por unidad (no por cuenta)**

#### 📜 User Story  
> *Como Juan Pérez, quiero ver mis dos departamentos (Torre A y Torre B) en la misma app, y cambiar entre ellos sin tener que cerrar y volver a entrar.*

#### ✅ Criterios de Aceptación
- [ ] Un mismo usuario puede tener múltiples roles (`owner`, `tenant`, `family_member`) en distintas unidades.  
- [ ] En el dashboard, el título principal es: *“Torre A, Depto 12 — Propietario”*.  
- [ ] Existe un selector desplegable: *“Cambiar a: Torre B, Depto 45”*.  
- [ ] Al cambiar, el JWT sigue siendo el mismo, pero el contexto cambia: `unit_id` y `tenant_id` se actualizan.  
- [ ] El motor de autorización valida permisos en tiempo real contra `user_unit_roles` (no contra claims del token).  
- [ ] Si intenta acceder a una unidad donde no es propietario → 403 Forbidden.  
- [ ] Se registra en auditoría: `context.switched: from_unit=X to_unit=Y`.

---

### ✅ **Feature 5: Soporte humano integrado (para quienes no entienden tecnología)**

#### 📜 User Story  
> *Como María González, quiero poder presionar un botón y hablar con alguien de SmartEdify si algo no funciona, sin tener que esperar horas en soporte.*

#### ✅ Criterios de Aceptación
- [ ] En la app web y móvil, existe un botón flotante: **“¿Necesitas ayuda?”**  
- [ ] Al presionarlo, se inicia una llamada telefónica automática al soporte técnico de SmartEdify (número local en Perú).  
- [ ] La llamada está pre-cargada con:  
  - Nombre del usuario  
  - Tenant ID  
  - Última acción realizada  
- [ ] El agente ve la pantalla del usuario en tiempo real (con consentimiento explícito).  
- [ ] La llamada se registra en bitácora como evento: `support.requested` con duración y resultado.  
- [ ] No se requiere crear ticket previo.  

---

## ✅ **6. Requisitos No Funcionales**

| Categoría | Requisito | Detalle |
|----------|----------|---------|
| **Rendimiento** | Tiempo de respuesta | ≤ 800ms en login, ≤ 120ms en validación de permisos. |
| | Disponibilidad | 99.95% uptime. SLA garantizado. |
| | Escalabilidad | Soportar 10K transacciones por minuto. Auto-scaling en AWS. |
| **Seguridad** | Cifrado | AES-256-GCM en reposo, TLS 1.3 en tránsito. |
| | Autenticación | DPoP obligatorio para APIs externas. MTLS para microservicios internos. |
| | Privacidad | Nada de datos sensibles (contraseñas, secretos TOTP) almacenados en texto plano. |
| | Cumplimiento | Cumple Ley N° 27157 (Perú), LPDP (Ley 29733), GDPR, NIST SP 800-63B. |
| | Auditoría | Bitácora inmutable con cadena de hashes (WORM DB). Todos los eventos firmados. |
| **Escalabilidad** | Multi-tenant | Soporta 1000+ tenants simultáneos. Datos aislados por `tenant_id`. |
| | Internacionalización | Soporte para español, portugués. Localización de políticas por país. |
| | Integración | API REST + OpenAPI 3.1. SDKs publicados en npm y PyPI. |

---

## ✅ **7. Suposiciones y Fuera de Alcance (Out of Scope)**

### ✅ Suposiciones
- Los usuarios tienen acceso a WhatsApp o un teléfono móvil.  
- Las comunidades ya tienen una lista de propietarios (no necesitamos validar con SUNARP en MVP).  
- El cliente (síndico o administradora) tiene capacidad para enviar mensajes por WhatsApp Business API.  
- La ley peruana será nuestra referencia base — otras jurisdicciones se adaptarán después.  
- Los usuarios no quieren “aprender a usar una app”. Quieren que la app aprenda a usarlos.

### ❌ Fuera de Alcance (No construiremos en esta fase)
| Item | Razón |
|------|-------|
| App nativa iOS/Android | Usaremos PWA (Progressive Web App) para evitar tiendas de apps y reducir fricción. |
| Integración directa con bancos | Usaremos APIs de pago estandarizadas (Mercado Pago, PSE, Pix). |
| Sistema de nómina o payroll | Ese es un módulo separado (Payroll Service). |
| Blockchain como base de datos | Usamos IPFS para inmutabilidad — blockchain añade costo innecesario. |
| Chatbot de IA para resolver dudas | Soporte humano es más efectivo y confiable en este mercado. |
| Gestión de mantenimiento o RFP | Son módulos independientes (Maintenance Service). |
| Registro de propietarios con DNI en línea | Validación manual en MVP. Futuro: integración con SUNARP. |

---

## ✅ **CONCLUSIÓN FINAL — DECLARACIÓN DEL PRODUCT MANAGER**

> “No estamos construyendo otra app de condominios.  
> Estamos construyendo la **primera infraestructura de identidad digital que hace que la ley funcione en el mundo real.**  
>   
> Si Juan puede votar respondiendo ‘SÍ’ por WhatsApp, y María puede generar una acta que valide ante una municipalidad...  
>   
> …entonces hemos ganado.  
>   
> Esta no es una función. Es una revolución.  
> Y empezamos hoy.”
