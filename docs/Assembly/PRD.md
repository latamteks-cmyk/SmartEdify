### Documento de Planificación y Requisitos — Assembly Service (SmartEdify)

---

# 1. Documento de Requisitos del Producto (PRD)

## 1.1 Visión del producto

El **Assembly Service** permite a condominios, juntas y asociaciones ejecutar **asambleas mixtas (presenciales + virtuales)** de forma **legal, auditable y transparente**. Se integra en el ecosistema SmartEdify y orquesta convocatorias, acreditación, quórum, votación unificada, actas y archivo.

## 1.2 Objetivos principales

* Cumplir con la **Ley N° 27157** y modificatorias (D. Leg. 1527).
* Garantizar **igualdad de derechos** entre participantes presenciales y virtuales.
* Automatizar generación de **actas firmadas digitalmente** con trazabilidad.
* Integrar módulos existentes: Auth, Compliance, Finance, Document, Communication, Payments.
* Exponer interfaces para Web App (RBAC), App Móvil (user) y Web Soporte (NOC).

## 1.3 Público objetivo

* **Propietarios**: acceden vía Web App o App Móvil para participar, votar y descargar actas.
* **Moderadores/Administradores**: gestionan convocatoria, agenda, acreditaciones, votos manuales, acta y publicación.
* **Secretarios**: co-firman actas y registran observaciones.
* **Soporte Local**: asiste en acreditación presencial y control técnico.
* **Soporte SmartEdify (NOC)**: visualiza métricas, auditoría y salud del servicio.

## 1.4 Alcance funcional

* Creación, validación y publicación de convocatorias.
* Acreditación presencial y virtual con verificación de identidad.
* Cómputo de quórum consolidado en tiempo real.
* Gestión de poderes de representación.
* Apertura/cierre de ítems de agenda.
* Votación electrónica con token 1-uso y anti-doble voto.
* Registro manual de asistencias y votos con boleta obligatoria.
* Consolidación de resultados ponderados.
* Generación de actas firmadas digitalmente y archivadas en WORM.
* Notificación multicanal de resultados y actas.
* Integración con Google Meet (grabación y captions).

## 1.5 Restricciones

* Solo Google Meet como plataforma de video.
* Firma digital conforme a estándares de SUNAT y RENIEC (Perú).
* Almacenamiento WORM obligatorio para evidencias.
* Tokens de voto deben tener TTL < 5 minutos.
* App Móvil solo para propietarios (no para moderadores).

## 1.6 Requisitos de éxito

* Latencia de voto p95 < 200 ms.
* Acta firmada y publicada ≤ 24h post sesión.
* > 98% entregabilidad de convocatorias y notificaciones.
* 0 incidencias de doble voto.
* Auditoría legal completa ante impugnaciones.

---

# 2. Análisis de Requisitos

## 2.1 Funcionales (FR)

1. **Convocatoria**

   * FR1.1: Crear asamblea con agenda y jurisdicción.
   * FR1.2: Validar convocatoria con Compliance.
   * FR1.3: Publicar convocatoria por múltiples canales.
   * FR1.4: Guardar PDF de convocatoria en Document con hash.

2. **Acreditación**

   * FR2.1: Registrar asistencia presencial con DNI/QR.
   * FR2.2: Registrar asistencia virtual con login + MFA y cámara ON.
   * FR2.3: Asociar poderes y coeficientes.
   * FR2.4: Emitir tablero de quórum SSE.

3. **Sesión**

   * FR3.1: Abrir/cerrar ítems de agenda.
   * FR3.2: Pausar/reanudar sesión con sellos de tiempo.
   * FR3.3: Registrar incidentes y mociones.

4. **Votación**

   * FR4.1: Abrir ventana de voto con step-up MFA opcional.
   * FR4.2: Emitir votos electrónicos con token 1-uso (JTI).
   * FR4.3: Bloquear doble voto por índice único.
   * FR4.4: Registrar votos manuales con boleta obligatoria.
   * FR4.5: Consolidar votos ponderados vía Finance.
   * FR4.6: Publicar resultados en vivo.

5. **Acta**

   * FR5.1: Generar borrador con MPC y evidencias.
   * FR5.2: Firmar digitalmente con TSA.
   * FR5.3: Publicar acta y notificar a participantes.
   * FR5.4: Archivar expediente en WORM con hash raíz.

6. **Auditoría y soporte**

   * FR6.1: Exponer API para NOC con métricas y logs.
   * FR6.2: Mantener manifiesto de evidencias con SHA-256.
   * FR6.3: Proveer trazabilidad de cada voto, asistencia y poder.

## 2.2 No funcionales (NFR)

* **Rendimiento:** soportar 10k votos/minuto.
* **Disponibilidad:** SLA 99.9% durante asambleas.
* **Escalabilidad:** auto-escalado horizontal en Kubernetes.
* **Seguridad:** MFA obligatorio, TLS 1.3, cifrado en reposo, RBAC + ABAC.
* **Legalidad:** cumplimiento con normativa peruana y LPH.
* **Usabilidad:** UI accesible (WCAG 2.1 AA), subtítulos en tiempo real.
* **Compatibilidad:** navegadores modernos y apps móviles iOS/Android.
* **Observabilidad:** métricas, logs y traces OTel con correlación `assembly_id`.

## 2.3 Integraciones

* **Auth Service**: OIDC, MFA, introspection.
* **Compliance Service**: validación normativa.
* **Finance Service**: coeficientes, morosidad, consolidación ponderada.
* **Document Service**: almacenamiento, firma TSA, WORM.
* **Communication Service**: notificaciones multicanal.
* **Payments Service**: cobros relacionados (si aplica).
* **Google Meet**: creación de sala, grabación, captions.

## 2.4 Requisitos legales específicos

* Convocatoria debe indicar modalidad *mixta*.
* Identificación virtual debe ser fehaciente (login + MFA + cámara ON).
* Quórum debe consolidar presencial + virtual en tiempo real.
* Votos manuales deben constar en acta con boleta escaneada.
* Acta debe reflejar modalidad, quórum, acuerdos y anexar evidencias.
* Acta debe ser firmada digitalmente y archivada en repositorio inmutable.

## 2.5 Riesgos identificados

* R1: Fallo de conectividad en sala → mitigación: redundancia de red y fallback de grabación local.
* R2: Impugnación legal por identificación débil → mitigación: MFA + cámara ON + logs.
* R3: Ataque de doble voto → mitigación: token JTI 1-uso + índice único.
* R4: Pérdida de evidencias → mitigación: archivo WORM con hash raíz y TSA.
* R5: Fallo de Meet API → mitigación: reanudación manual + evidencia de incidencia.

## 2.6 Priorización (MoSCoW)

* **Must Have**: convocatoria, acreditación, quórum, votación unificada, acta firmada, archivo WORM, cumplimiento legal.
* **Should Have**: soporte multicanal en notificaciones, MPC para borrador de acta.
* **Could Have**: votación por bloques/delegados, OCR automático en boletas.
* **Won’t Have (MVP)**: integración con plataformas de VC distintas a Meet.

---
### Roadmap de Implementación — Assembly Service (SmartEdify)
---

# 1. Fase 0 — Preparación

**Objetivo:** bases técnicas y legales.
**Entregables:**

* Revisión legal final de D. Leg. 1527 y Ley 27157.
* Setup de repositorios, CI/CD, Kubernetes y observabilidad (OTel, Prometheus, Grafana).
* Diseño detallado de esquemas DB y colas (Postgres, Kafka/NATS).
* Acuerdo de contratos OpenAPI entre microservicios.

---

# 2. Fase 1 — Núcleo de Asamblea

**Objetivo:** ciclo de vida básico de asamblea.
**Entregables:**

* CRUD de asambleas con estados: Draft → Validated → Notified.
* Validación de agenda y convocatoria con Compliance Service.
* Creación de sala Google Meet (API).
* Generación de convocatoria PDF y almacenamiento en Document Service.
* Publicación de convocatoria vía Communication Service.
* Exposición de APIs `assemblies/*`, `agenda/validate`, `call/publish`, `meet/*`.
  **Éxito:** convocatorias legales y trazables emitidas desde SmartEdify.

---

# 3. Fase 2 — Acreditación y Quórum

**Objetivo:** identificación fehaciente y tablero en vivo.
**Entregables:**

* Check-in presencial (DNI/QR) y virtual (login + MFA + cámara ON).
* Registro de poderes y asociación de coeficientes (Finance).
* Tablero de quórum consolidado con SSE/WebSocket.
* APIs `attendees/checkin`, `proxies`, `quorum/stream`.
* Logs de acreditación vinculados a identidad y canal.
  **Éxito:** quórum legalmente válido, visible en sala y en remoto.

---

# 4. Fase 3 — Votación Unificada

**Objetivo:** garantizar voto único, ponderado y auditable.
**Entregables:**

* Apertura/cierre de ítems de agenda.
* Emisión de tokens JTI 1-uso para votos.
* Registro de votos electrónicos (web/móvil) con anti-doble voto.
* Registro manual de votos con boleta obligatoria.
* Consolidación de resultados ponderados (Finance).
* Publicación de resultados en vivo (SSE).
* APIs `vote/open`, `vote`, `vote/close`, `votes/manual`, `results`.
  **Éxito:** resultados consolidados en tiempo real, sin riesgo de doble voto.

---

# 5. Fase 4 — Acta, Firma y Archivo

**Objetivo:** generación de actas legales e inmutables.
**Entregables:**

* Borrador incremental de acta con MPC y logs de sesión.
* Firma digital + TSA.
* Publicación y notificación de acta.
* Archivo WORM en Document Service con hash raíz.
* APIs `minutes/draft`, `minutes/sign`, `minutes/publish`.
  **Éxito:** actas firmadas digitalmente, archivadas y notificadas ≤ 24h.

---

# 6. Fase 5 — Auditoría, Observabilidad y Soporte

**Objetivo:** resiliencia, métricas y soporte.
**Entregables:**

* Exposición de APIs de auditoría y manifiesto de evidencias.
* Dashboard de métricas en NOC (Web Soporte).
* Alertas proactivas (quórum bajo, latencia alta, fallos de Meet API).
* Documentación de auditoría para impugnaciones.
  **Éxito:** soporte técnico puede monitorear, diagnosticar y auditar cada asamblea.

---

# 7. Fase 6 — Optimización y Extensiones

**Objetivo:** mejorar experiencia y ampliar alcance.
**Entregables:**

* Subtítulos automáticos integrados en acta.
* OCR de boletas manuales para indexación.
* Modos de votación extendidos (bloques/delegados).
* Escalado automático validado con pruebas de carga (≥ 10k votos/min).
  **Éxito:** sistema robusto y listo para certificaciones adicionales.

---

# Cronograma Resumido

| Fase                   | Semanas | Hitos principales     |
| ---------------------- | ------- | --------------------- |
| 0. Preparación         |         | Infraestructura lista |
| 1. Núcleo Asamblea     |         | Convocatorias legales |
| 2. Acreditación/Quórum |         | Tablero en vivo       |
| 3. Votación Unificada  |         | Resultados auditables |
| 4. Acta/Firma/Archivo  |         | Actas firmadas WORM   |
| 5. Auditoría/Soporte   |         | NOC con métricas      |
| 6. Extensiones         |         | Subtítulos/OCR/escala |

---

# Dependencias críticas

* **Auth Service** debe estar operativo con OIDC y MFA antes de Fase 2.
* **Finance Service** debe exponer coeficientes antes de Fase 2.
* **Document Service** con firma TSA/WORM debe estar listo antes de Fase 4.
* **Communication Service** operativo desde Fase 1 para convocatorias.
* **Meet API** integración validada antes de Fase 1.

