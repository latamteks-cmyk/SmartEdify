Estructura de monorepo propuesta (SaaS, PMV primero). Incluye `doc/`.

```
smartedify/
├─ apps/
│  ├─ web-admin/                # Next.js (SSR/ISR), UI Admin
│  ├─ web-user/                 # Next.js, UI Usuario
│  ├─ mobile/                   # Expo/React Native
│  └─ bff/
│     ├─ admin/                 # BFF Admin (OIDC PKCE, CSRF, cache corto)
│     ├─ app/                   # BFF Usuario
│     └─ mobile/                # BFF Móvil
├─ services/
│  ├─ pmv/                      # Línea 1 (cercanos al frontend)
│  │  ├─ assembly-service/
│  │  │  ├─ src/
│  │  │  ├─ openapi/            # contratos REST
│  │  │  ├─ db/migrations/
│  │  │  ├─ deployments/        # ecs/k8s manifests
│  │  │  └─ tests/              # unit/contract/e2e
│  │  ├─ reservation-service/
│  │  └─ maintenance-service/
│  ├─ support/                  # Línea 2 (soporte al PMV)
│  │  ├─ auth-service/
│  │  ├─ user-service/          # OPA policies en /policies/
│  │  ├─ tenants-service/
│  │  ├─ document-service/
│  │  ├─ communication-service/
│  │  └─ finance-service/
│  └─ complementary/            # Línea 3 (complementarios)
│     ├─ payments-service/
│     ├─ compliance-service/
│     ├─ payroll-service/
│     ├─ certification-service/
│     ├─ support-bot-service/
│     └─ facility-security-service/
├─ platform/
│  ├─ gateway/                  # rutas, WAF, domains, rate limits
│  ├─ mesh/                     # mTLS, S2S authz, retries, circuit breaking
│  ├─ events/                   # AsyncAPI, esquemas (JSON Schema/Avro), outbox
│  ├─ observability/            # Otel collectors, dashboards, alerts, SLOs
│  ├─ security/                 # OPA bundles, CSP/HSTS, secret policies, KMS
│  └─ shared/                   # libs comunes: tipos, SDK OpenAPI, utils, clients
├─ contracts/
│  ├─ openapi/                  # *.yaml por servicio
│  ├─ asyncapi/                 # eventos por dominio
│  └─ pacts/                    # tests consumidor-productor (BFF↔servicios)
├─ infra/
│  ├─ terraform/
│  │  ├─ modules/               # vpc, rds, redis, s3, cloudfront, waf, ecs, iam
│  │  └─ envs/                  # dev, stg, prod
│  └─ cicd/                     # pipelines, plantillas, imágenes base
├─ config/
│  ├─ dev/ stg/ prod/           # feature flags, env vars plantillas
│  └─ secrets/README.md         # gestión en KMS/Secrets Manager
├─ qa/
│  ├─ k6/                       # pruebas de carga
│  └─ chaos/                    # experimentos de resiliencia
├─ scripts/                     # make, db bootstrap, codegen, lint
├─ .github/workflows/           # CI (lint, tests, contract, build, deploy)
├─ doc/                         # ADRs, arquitectura, runbooks, DPIA, políticas
│  ├─ adr/
│  ├─ diagrams/                 # mermaid, drawio
│  ├─ runbooks/                 # oncall, incident, DR
│  ├─ security/                 # modelo de amenazas, controles Ley 29733
│  └─ product/                  # roadmaps, criterios PMV
└─ README.md
```

Convenciones clave:

* Un servicio = su propio Postgres, `openapi/`, `db/migrations/`, `deployments/`, `tests/`.
* Eventos versionados en `platform/events/`; contratos sincronizados con CI.
* BFF termina PKCE y oculta refresh. UIs consumen solo BFF.
* `platform/mesh` gobierna tráfico este-oeste; Gateway solo norte-sur.
* `doc/adr` registra decisiones; cambios breaking requieren ADR y versión de contrato.

### 2.3 Reglas operativas y Diagrama de Flujo

## Reglas operativas

Clientes (UI Web Admin, Web User, Móvil) → BFF por canal → API Gateway → servicios PMV.

Servicio→Servicio: solo por malla, nunca por Gateway. Autorización con OAuth2 m2m (client credentials) o JWT exchange corto.

Identidad multi-tenant: tenant_id en JWT del usuario; cada servicio revalida políticas con User/OPA y aplica RLS.

Egresos externos: SMS/Email/Payments solo vía Egress Gateway con allowlist y DLP.

Eventos primero: comandos sincrónicos mínimos; todo lo demás, eventos con idempotencia.

### Diagrama acordado

```mermaid
flowchart LR
  subgraph UI[Frontend]
    A[Web Admin]-->BFFA[BFF Admin]
    U[Web User]-->BFFU[BFF App]
    M[Móvil]-->BFFM[BFF Móvil]
  end

  BFFA-->GW[API Gateway+WAF]
  BFFU-->GW
  BFFM-->GW

  subgraph L1[PMV · Línea 1]
    ASSEMBLY[Assembly Service]
    RES[Reservation Service]
    MAINT[Maintenance Service]
  end

  subgraph L2[Soporte PMV · Línea 2]
    AUTH[Auth Service]
    USER[User Service + OPA]
    TEN[Tenants Service]
    DOC[Document Service]
    COM[Communication Service]
    FIN[Finance Service]
  end

  subgraph L3[Complementarios · Línea 3]
    PAY[Payments]
    COMP[Compliance]
    PAYR[Payroll]
    CERT[Certification]
    BOT[Support-Bot]
    SEC[Facility Security]
  end

  %% Norte-sur
  GW-->AUTH
  GW-->ASSEMBLY
  GW-->RES
  GW-->MAINT
  GW-->USER
  GW-->TEN
  GW-->DOC
  GW-->COM
  GW-->FIN

  %% Este-oeste a través de malla (simplificado)
  ASSEMBLY-.mTLS/.->USER
  ASSEMBLY-.mTLS/.->TEN
  ASSEMBLY-.mTLS/.->DOC
  ASSEMBLY-.mTLS/.->COM
  ASSEMBLY-.mTLS/.->FIN

  RES-.mTLS/.->TEN
  RES-.mTLS/.->FIN
  RES-.mTLS/.->COM

  MAINT-.mTLS/.->DOC
  MAINT-.mTLS/.->FIN
  MAINT-.mTLS/.->COM

  FIN-.eventos/.->PAY
  ASSEMBLY-.reglas/.->COMP
```

Aquí tienes los flujos, nivel BA. Primero orquestación SmartEdify. Luego cada microservicio. Assembly Service al detalle.

# Orquestación SmartEdify (end-to-end)

```mermaid
sequenceDiagram
  autonumber
  actor Admin as Moderador/Administrador
  participant ASM as Assembly Service
  participant CMP as Compliance Service
  participant COM as Communication Service
  participant AUT as Auth Service
  participant FIN as Finance Service
  participant DOC as Document Service
  participant PAY as Payments Service
  participant MEET as Google Meet

  Admin->>ASM: Crear asamblea (tipo, fecha, agenda)
  ASM->>CMP: Validar convocatoria/agenda
  CMP-->>ASM: OK + reglas aplicables
  ASM->>COM: Generar convocatoria multicanal (meet_link)
  ASM->>MEET: Crear sala + grabación/captions
  COM-->>Propietarios: Avisos enviados

  Note over Admin,ASM: Fase “Antes”

  participant Pres as Asistente presencial
  participant Virt as Asistente virtual
  Pres->>ASM: Check-in QR + DNI
  Virt->>AUT: Login OIDC + MFA (cámara ON en acreditación)
  AUT-->>ASM: Token válido + identidad
  ASM->>FIN: Traer coeficientes y morosidad
  ASM-->>Admin: Quórum tiempo real

  Note over Admin,ASM: Fase “Durante”

  Admin->>ASM: Abrir Ítem 1
  ASM->>Virt: Apertura voto electrónico
  ASM->>Pres: Conteo presencial (boletas/hand)
  Virt-->>ASM: Votos electrónicos
  Admin->>ASM: Registro manual si aplica (boleta obligatoria)
  ASM->>FIN: Cálculo ponderado
  ASM-->>Todos: Resultado consolidado

  ASM->>DOC: Borrador de acta + evidencias
  ASM->>MEET: Cerrar grabación
  ASM->>DOC: Generar acta final (PDF)
  DOC->>Firma: Flujo de firma + TSA
  ASM->>COM: Publicar y notificar acta
  ASM->>DOC: Archivar WORM (expediente + hash raíz)
```

# Assembly Service — detalle por flujo

## 0) Ciclo de vida y estados

```mermaid
stateDiagram-v2
  [*] --> Draft
  Draft --> Validated: agenda.validated (Compliance OK)
  Validated --> Notified: call.published
  Notified --> CheckInOpen: session.opened
  CheckInOpen --> InSession: quorum.reached | session.started
  InSession --> Paused: session.paused
  Paused --> InSession: session.resumed
  InSession --> Voting: vote.opened(item)
  Voting --> InSession: vote.closed(item)
  InSession --> MinutesDraft: session.closed
  MinutesDraft --> Signed: minutes.signed
  Signed --> Published: minutes.published
  Published --> Archived: evidence.archived(WORM)
  Archived --> [*]
```

## 1) Creación y validación de asamblea

* Input: tipo, jurisdicción, fecha, agenda preliminar, reglas del reglamento.
* Pasos:

  1. `POST /assemblies` → estado `Draft`.
  2. `POST /assemblies/{id}/agenda/validate` → Compliance valida plazos, mayorías, quórum por ítem.
  3. `POST /assemblies/{id}/meet` → crea sala Meet, activa captions y esquema de grabación.
  4. `POST /assemblies/{id}/call/publish` → Communication envía convocatoria; Document guarda PDF con hash.
* Salidas: `agenda.validated`, `call.published`, `meet.created`.

## 2) Acreditación y check-in

* Presencial: escaneo QR, verificación DNI, device binding opcional.
* Virtual: OIDC + MFA; **cámara ON** durante acreditación.
* Poderes: carga y validación de tope.
* Datos guardados: `attendee.source`, coeficiente, canal, evidencias.
* Eventos: `attendee.checked_in`, `proxy.registered`.
* Reglas: deduplicación por persona; bloqueo si morosidad afecta voto.

## 3) Cómputo de quórum en vivo

* Motor consolida: presencial + virtual + representados.
* UI: tablero público espejo para sala y para Meet.
* KPI: p95 < 1 s para refresco.
* Evento: `quorum.updated`. Umbrales por ítem disponibles.

## 4) Moderación y órdenes del día

* Abrir/cerrar ítems secuenciales.
* Turnos de palabra: cola unificada.
* Incidencias: moción, objeción, pausa con sello de tiempo.
* Eventos: `item.opened`, `incident.logged`, `item.closed`.

## 5) Votación electrónica unificada

```mermaid
sequenceDiagram
  autonumber
  participant Admin
  participant ASM
  participant AUT as Auth
  participant FIN as Finance
  participant DOC as Document

  Admin->>ASM: Abrir voto (ítem N, regla)
  ASM->>AUT: Introspect token + step-up MFA si sensible
  ASM->>FIN: Traer coeficiente vigente
  ASM-->>Votantes: Ventana de voto abierta
  Votantes-->>ASM: Emiten voto (1 sola vez)
  ASM->>ASM: Anti doble voto + recibo cifrado
  Admin->>ASM: Registrar votos manuales (boleta obligatoria)
  ASM->>DOC: Anexar boletas manuales (hash)
  ASM->>FIN: Consolidar ponderado
  ASM-->>Todos: Resultado ítem N
  ASM->>DOC: Log de apertura/cierre + recibos
```

* Modos: nominal, secreto, coeficiente, delegados, bloque.
* Manual: solo moderador. Campo `source=manual`, `ballot_url` obligatorio.
* Seguridad: one-time vote token, replay guard, nonces, hashing de recibo.
* Evento: `vote.closed`, `vote.results_published`.

## 6) Redacción de acta en vivo

* Borrador incremental: resúmenes MPC + marcadores a clips.
* Sección fija “Registros manuales”.
* Evidencias: convocatoria, asistentes, poderes, logs voto, grabación, hashes.
* Salida: `draft.updated` → PDF provisional en Document.

## 7) Firma, publicación y archivo

* Firma digital cualificada + TSA.
* `minutes.signed` → `minutes.published` → notificación multicanal.
* Archivo WORM: expediente + **hash raíz** de manifiesto.
* Eventos: `minutes.signed`, `minutes.published`, `evidence.archived`.

## 8) Post: acuerdos, seguimiento, impugnaciones

* Plan de tareas por acuerdo (responsable, fecha).
* Ventana de impugnación según Compliance.
* Recordatorios y reporte de cumplimiento.

---

# Workflows por microservicio

## Auth Service

* **Login**: `/oidc/authorize` → MFA (TOTP/WebAuthn) → token con `tenant_id` y scopes.
* **Step-up**: solicitar MFA para abrir votos sensibles.
* **Introspect/Revocar**: tokens rotados; eventos de seguridad auditados.

## Compliance Service

* **Validación**: entrada agenda + jurisdicción → reglas aplicables → dictamen.
* **Alertas**: cambios normativos → `compliance.rule.updated`.
* **Cálculo**: mayorías por ítem, plazos, requisitos de convocatoria y firma.

## Finance Service

* **Coeficientes**: padrón, alícuotas, morosidad.
* **Cobranzas**: conciliación si hay pagos de convocatoria o multas de asamblea.
* **Estados**: exposición de coeficiente vigente por persona.

## Payments Service

* **Intents**: cobro de derechos o servicios ligados a la asamblea.
* **Webhooks**: `payment_succeeded` → Finance concilia.

## Communication Service

* **Convocatoria**: plantillas, multicanal, acuse y rebote.
* **Sesión**: recordatorios, cambio de sala, emergencias.
* **Publicación**: distribución de acta y acuerdos.

## Document Service

* **Almacenamiento**: S3, versiones, OCR.
* **Firma**: flujo con TSA, evidencia LTV.
* **WORM**: expediente con índice y hashes; compendio de boletas.

## SupportBot Service

* **Onboarding**: guía paso a paso para asistentes.
* **FAQ**: micropolíticas de voto, quórum, soporte técnico.
* **Escalamiento**: integra con Communication si hay incidentes.

## FacilitySecurity Service

* **Perímetro**: monitoreo de cámaras durante eventos grandes.
* **Accesos**: registro de apertura/cierre si se usa control facial.
* **Privacidad**: solo eventos y metadatos al acta si procede.

## Reservation Service

* **Espacios**: bloqueo y logística del salón.
* **Calendario**: evitar choques con otras reservas.
* **Costos**: traspaso a Finance si aplica.

## Maintenance Service

* **Soporte**: equipos A/V, micrófonos, UPS, conectividad.
* **OTs**: instalación, prueba, contingencia.
* **Post**: correctivos si falló equipamiento.

## Payroll Service

* **Roles**: validación de moderador/secretario si son staff.
* **Trazabilidad**: asistencia de personal en evento.
* **Documentos**: export regulatorios si aplica.

## Certification Service

* **Cumplimientos**: seguridad del local, aforo, rutas de evacuación.
* **Inspecciones**: registros y hallazgos anexables al expediente.

---

# Entregables operativos rápidos

* Diagramas incluidos.
* Estados y eventos cerrados.
* Campos críticos definidos para legalidad: `source=manual`, hashes, TSA, cámara ON en acreditación, quórum público.


Arquitectura propuesta de **Assembly Service**. Objetivo: sesiones mixtas legales, auditables, resilientes. Estilo hexagonal, eventos primero, consistencia eventual controlada.

# Vista C4 (Container)

```mermaid
flowchart LR
  subgraph Client
    Web[Web Admin]
    Mobile[App Propietario]
  end
  APIGW[API Gateway + WAF\nJWT validate + rate limit]
  ASMSVC[Assembly Service\n'HTTP gRPC Events']
  AUTH[Auth Service]
  COMP[Compliance Service]
  FIN[Finance Service]
  COMM[Communication Service]
  DOC[Document Service]
  PAY[Payments Service]
  MEET[Google Meet API]
  BUS[Event Bus 'Kafka/NATS']
  CACHE[Redis]
  DB['PostgreSQL']
  OBJ['S3/WORM']
  TRACE[Observability\nOTel + Tempo/Jaeger + Prom]

  Web --> APIGW --> ASMSVC
  Mobile --> APIGW
  ASMSVC --> AUTH
  ASMSVC <--> COMP
  ASMSVC <--> FIN
  ASMSVC --> COMM
  ASMSVC --> DOC
  ASMSVC --> PAY
  ASMSVC --> MEET
  ASMSVC --> BUS
  ASMSVC --> CACHE
  ASMSVC --> DB
  DOC --> OBJ
  ASMSVC --> TRACE
```

# Descomposición interna (hexagonal)

```mermaid
flowchart TB
  subgraph Application
    CMD[Command Handlers]
    QRY[Query Handlers]
    SAGA[Sagas/Process Manager]
    POL[Policies/Domain Rules]
  end
  subgraph Domain
    AGG[Aggregates:\nAssembly, AgendaItem, VoteBatch, Quorum, Minutes]
    EVT[Domain Events]
    ACL[Access Control (roles, powers)]
  end
  subgraph Ports
    Repo[Repositories]
    MeetPort[MeetPort]
    DocPort[DocPort]
    CompPort[CompliancePort]
    FinPort[FinancePort]
    CommPort[CommPort]
    IdPPort[AuthPort]
    BusPort[EventPublisher]
  end
  INFRA[(Adapters: Postgres, Redis, Kafka, gRPC/REST, S3, JWKS)]
  CMD --> AGG --> EVT --> SAGA
  QRY --> Repo
  Ports --> INFRA
  POL --> AGG
```

# Módulos

* **Assemblies**: ciclo de vida, estados, check-in, quórum.
* **Voting**: ventanas, anti-doble voto, ponderación.
* **ManualRecords**: alta manual con boleta obligatoria.
* **Minutes**: borrador, firma, publicación, archivo.
* **Integration**: Meet, Compliance, Finance, Document, Communication, Payments.
* **Audit**: eventos, hashes, manifiesto de evidencias.
* **Access**: enforcement de roles y poderes.

# Datos (relacional normalizado)

* `assemblies(id, tenant_id, tipo, modalidad, fecha, estado, meet_id, meet_link, compliance_validation_id, hash_convocatoria, created_at)`
* `agenda_items(id, assembly_id, titulo, tipo_decision, mayoria, norma_ref, orden, estado)`
* `attendees(id, assembly_id, persona_id, rol, canal, coeficiente, present, source ENUM['auto','manual'], manual_reason, manual_by, manual_evidence_id, created_at)`
* `proxies(id, assembly_id, otorgante_id, apoderado_id, limite, vigencia, evidencia_doc_id)`
* `votes(id, item_id, voter_id, mode, value, coef_aplicado, receipt_hash, source ENUM['auto','manual'], manual_reason, ballot_doc_id, overridden_vote_id, ts)`
* `quorum_snapshots(id, assembly_id, ts, coef_presentes, coef_virtuales, coef_poderes)`
* `minutes(id, assembly_id, status, url_pdf, hash, tsa_token, annex_boletas_manifest_id)`
* `evidence(id, assembly_id, tipo, doc_id, hash, ts)`
* `outbox(id, aggregate, event_type, payload, status)`  // transactional outbox

Índices: `tenant_id`, `(assembly_id, estado)`, `(item_id, voter_id UNIQUE WHERE source='auto')`, `receipt_hash UNIQUE`.

# API (REST + gRPC internos)

**Prefix** `/api/assembly/v1/*`

* Assemblies: `POST /assemblies`, `GET /assemblies/{id}`, `POST /assemblies/{id}/agenda/validate`, `POST /assemblies/{id}/call/publish`, `POST /assemblies/{id}/session/open|close|pause|resume`, `GET /assemblies/{id}/quorum/stream` (SSE).
* Meet: `POST /assemblies/{id}/meet` (crear sala + captions + grabación), `POST /assemblies/{id}/meet/start|stop-recording`.
* Attendees: `POST /assemblies/{id}/attendees/checkin`, `POST /assemblies/{id}/proxies`, `GET /assemblies/{id}/attendees`.
* Voting: `POST /items/{itemId}/vote/open|close`, `POST /items/{itemId}/vote` (token de voto 1-uso), `GET /items/{itemId}/results`.
* Manual: `POST /attendees/manual`, `POST /items/{itemId}/votes/manual`, `POST /votes/{voteId}/override`, `GET /assemblies/{id}/manual-records`.
* Minutes: `POST /assemblies/{id}/minutes/draft`, `POST /assemblies/{id}/minutes/sign`, `POST /assemblies/{id}/minutes/publish`, `GET /assemblies/{id}/minutes`.

Scopes: `assembly:read|write|admin`. Step-up MFA para `vote.open`, `minutes.sign`, `manual.*`.

# Eventos (Kafka/NATS)

* `assembly.created`, `agenda.validated`, `call.published`, `session.started|paused|resumed|closed`
* `attendee.checked_in`, `proxy.registered`, `quorum.updated`
* `vote.opened`, `vote.closed`, `vote.results_published`, `manual.attendee.added`, `manual.vote.recorded`, `vote.overridden`
* `minutes.draft.updated`, `minutes.signed`, `minutes.published`, `evidence.archived`

Diseño **outbox+inbox** para entrega al menos una vez. Idempotencia por `event_id`.

# Integraciones

* **Auth**: OIDC introspection, JWKS caché, WebAuthn, TOTP. Device binding opcional.
* **Compliance**: sync validate; cache con TTL por jurisdicción.
* **Finance**: coeficientes, morosidad, ponderación.
* **Document**: subida de boletas, PDF de acta, manifiesto de evidencias, WORM, TSA.
* **Communication**: convocatorias, recordatorios, publicación de acta.
* **Payments**: cobros asociados si aplica.
* **Google Meet**: creación de sala, start/stop recording, captions; almacenar IDs y hashes.

# Seguridad

* JWT validado en Gateway y revalidado en servicio. `tenant_id` obligatorio.
* RBAC por rol y **poderes**; políticas a nivel de ítem.
* **Anti-doble voto**: token de voto 1-uso (JTI + nonce) + unique index `(item_id, voter_id)` para `source='auto'`.
* **Registros manuales**: requieran archivo en Document; bloqueo sin boleta.
* **Evidencias**: hashes SHA-256, hash raíz de manifiesto, TSA en acta.
* PII minimizada. Cifrado en tránsito (TLS) y en reposo (PG crypto-at-rest + S3 SSE-KMS).

# Escalabilidad y rendimiento

* Read-heavy: **CQRS light**. Queries denormalizadas en vistas materializadas (`assembly_view`, `results_view`), cache Redis.
* WebSockets/SSE para quórum y resultados en tiempo real.
* p95 < 200 ms en `vote.open/close`; >10k votos/min con partición por `tenant_id`.
* Sharding por `tenant_id` en Kafka y claves en Redis.
* Workers asíncronos para generación de actas y archivado.

# Resiliencia y consistencia

* **Sagas** para: convocatoria, sesión, votación por ítem, acta. Compensaciones: reenvío de comunicaciones, re-cierre de voto, reintento de firma, rearchivo WORM.
* Retries exponenciales, circuit breakers a externos.
* Modo degradado: si Meet falla, registrar fallback y permitir reanudación.

# Observabilidad

* OTel traces con `tenant_id`, `assembly_id`, `item_id`.
* Métricas: TPS votos, latencia open/close, quorum drift, fallos manuales sin boleta, tiempo a publicación de acta.
* Logs firmados y tamper-evident.

# Tecnología sugerida

* **Runtime**: Kotlin + Spring Boot o Go + chi/fx.
* **DB**: PostgreSQL 15 + pgcrypto + logical decoding (future CDC).
* **Cache**: Redis 7.
* **Bus**: Kafka o NATS JetStream.
* **Docs**: S3 compatible + Glacier; WORM.
* **Infra**: Kubernetes, HPA por QPS y lag de cola.
* **API**: OpenAPI 3.1 + gRPC internos.
* **AuthN**: OIDC/OAuth2 provider externo (Auth Service).

# Decisiones clave

* Google Meet como único VC.
* Manuales marcados e indisociables de boleta.
* Legalidad: cámara ON en acreditación virtual, quórum público, resultados consolidados.
* Outbox para fiabilidad de eventos.
* CQRS light para UX en vivo.

# Backlog técnico inmediato

1. Esquema SQL y migraciones.
2. OpenAPI por módulo.
3. Adapters: MeetPort, DocPort, CompliancePort.
4. Vistas materializadas y SSE para quórum/resultados.
5. Sagas y outbox.
6. Pruebas de carga de voto y latencia.

Arquitectura propuesta de **Assembly Service**. Objetivo: sesiones mixtas legales, auditables, resilientes. Estilo hexagonal, eventos primero, consistencia eventual controlada.

# Vista C4 (Container)

```mermaid
flowchart LR
  subgraph Client
    Web[Web Admin]
    Mobile[App Propietario]
  end
  APIGW[API Gateway + WAF\nJWT validate + rate limit]
  ASMSVC[Assembly Service\n'HTTP gRPC Events']
  AUTH[Auth Service]
  COMP[Compliance Service]
  FIN[Finance Service]
  COMM[Communication Service]
  DOC[Document Service]
  PAY[Payments Service]
  MEET[Google Meet API]
  BUS[Event Bus 'Kafka/NATS']
  CACHE[Redis]
  DB['PostgreSQL']
  OBJ['S3/WORM']
  TRACE[Observability\nOTel + Tempo/Jaeger + Prom]

  Web --> APIGW --> ASMSVC
  Mobile --> APIGW
  ASMSVC --> AUTH
  ASMSVC <--> COMP
  ASMSVC <--> FIN
  ASMSVC --> COMM
  ASMSVC --> DOC
  ASMSVC --> PAY
  ASMSVC --> MEET
  ASMSVC --> BUS
  ASMSVC --> CACHE
  ASMSVC --> DB
  DOC --> OBJ
  ASMSVC --> TRACE
```

# Descomposición interna (hexagonal)

```mermaid
flowchart TB
  subgraph Application
    CMD[Command Handlers]
    QRY[Query Handlers]
    SAGA[Sagas/Process Manager]
    POL[Policies/Domain Rules]
  end
  subgraph Domain
    AGG[Aggregates:\nAssembly, AgendaItem, VoteBatch, Quorum, Minutes]
    EVT[Domain Events]
    ACL[Access Control (roles, powers)]
  end
  subgraph Ports
    Repo[Repositories]
    MeetPort[MeetPort]
    DocPort[DocPort]
    CompPort[CompliancePort]
    FinPort[FinancePort]
    CommPort[CommPort]
    IdPPort[AuthPort]
    BusPort[EventPublisher]
  end
  INFRA[(Adapters: Postgres, Redis, Kafka, gRPC/REST, S3, JWKS)]
  CMD --> AGG --> EVT --> SAGA
  QRY --> Repo
  Ports --> INFRA
  POL --> AGG
```

# Módulos

* **Assemblies**: ciclo de vida, estados, check-in, quórum.
* **Voting**: ventanas, anti-doble voto, ponderación.
* **ManualRecords**: alta manual con boleta obligatoria.
* **Minutes**: borrador, firma, publicación, archivo.
* **Integration**: Meet, Compliance, Finance, Document, Communication, Payments.
* **Audit**: eventos, hashes, manifiesto de evidencias.
* **Access**: enforcement de roles y poderes.

# Datos (relacional normalizado)

* `assemblies(id, tenant_id, tipo, modalidad, fecha, estado, meet_id, meet_link, compliance_validation_id, hash_convocatoria, created_at)`
* `agenda_items(id, assembly_id, titulo, tipo_decision, mayoria, norma_ref, orden, estado)`
* `attendees(id, assembly_id, persona_id, rol, canal, coeficiente, present, source ENUM['auto','manual'], manual_reason, manual_by, manual_evidence_id, created_at)`
* `proxies(id, assembly_id, otorgante_id, apoderado_id, limite, vigencia, evidencia_doc_id)`
* `votes(id, item_id, voter_id, mode, value, coef_aplicado, receipt_hash, source ENUM['auto','manual'], manual_reason, ballot_doc_id, overridden_vote_id, ts)`
* `quorum_snapshots(id, assembly_id, ts, coef_presentes, coef_virtuales, coef_poderes)`
* `minutes(id, assembly_id, status, url_pdf, hash, tsa_token, annex_boletas_manifest_id)`
* `evidence(id, assembly_id, tipo, doc_id, hash, ts)`
* `outbox(id, aggregate, event_type, payload, status)`  // transactional outbox

Índices: `tenant_id`, `(assembly_id, estado)`, `(item_id, voter_id UNIQUE WHERE source='auto')`, `receipt_hash UNIQUE`.

# API (REST + gRPC internos)

**Prefix** `/api/assembly/v1/*`

* Assemblies: `POST /assemblies`, `GET /assemblies/{id}`, `POST /assemblies/{id}/agenda/validate`, `POST /assemblies/{id}/call/publish`, `POST /assemblies/{id}/session/open|close|pause|resume`, `GET /assemblies/{id}/quorum/stream` (SSE).
* Meet: `POST /assemblies/{id}/meet` (crear sala + captions + grabación), `POST /assemblies/{id}/meet/start|stop-recording`.
* Attendees: `POST /assemblies/{id}/attendees/checkin`, `POST /assemblies/{id}/proxies`, `GET /assemblies/{id}/attendees`.
* Voting: `POST /items/{itemId}/vote/open|close`, `POST /items/{itemId}/vote` (token de voto 1-uso), `GET /items/{itemId}/results`.
* Manual: `POST /attendees/manual`, `POST /items/{itemId}/votes/manual`, `POST /votes/{voteId}/override`, `GET /assemblies/{id}/manual-records`.
* Minutes: `POST /assemblies/{id}/minutes/draft`, `POST /assemblies/{id}/minutes/sign`, `POST /assemblies/{id}/minutes/publish`, `GET /assemblies/{id}/minutes`.

Scopes: `assembly:read|write|admin`. Step-up MFA para `vote.open`, `minutes.sign`, `manual.*`.

# Eventos (Kafka/NATS)

* `assembly.created`, `agenda.validated`, `call.published`, `session.started|paused|resumed|closed`
* `attendee.checked_in`, `proxy.registered`, `quorum.updated`
* `vote.opened`, `vote.closed`, `vote.results_published`, `manual.attendee.added`, `manual.vote.recorded`, `vote.overridden`
* `minutes.draft.updated`, `minutes.signed`, `minutes.published`, `evidence.archived`

Diseño **outbox+inbox** para entrega al menos una vez. Idempotencia por `event_id`.

# Integraciones

* **Auth**: OIDC introspection, JWKS caché, WebAuthn, TOTP. Device binding opcional.
* **Compliance**: sync validate; cache con TTL por jurisdicción.
* **Finance**: coeficientes, morosidad, ponderación.
* **Document**: subida de boletas, PDF de acta, manifiesto de evidencias, WORM, TSA.
* **Communication**: convocatorias, recordatorios, publicación de acta.
* **Payments**: cobros asociados si aplica.
* **Google Meet**: creación de sala, start/stop recording, captions; almacenar IDs y hashes.

# Seguridad

* JWT validado en Gateway y revalidado en servicio. `tenant_id` obligatorio.
* RBAC por rol y **poderes**; políticas a nivel de ítem.
* **Anti-doble voto**: token de voto 1-uso (JTI + nonce) + unique index `(item_id, voter_id)` para `source='auto'`.
* **Registros manuales**: requieran archivo en Document; bloqueo sin boleta.
* **Evidencias**: hashes SHA-256, hash raíz de manifiesto, TSA en acta.
* PII minimizada. Cifrado en tránsito (TLS) y en reposo (PG crypto-at-rest + S3 SSE-KMS).

# Escalabilidad y rendimiento

* Read-heavy: **CQRS light**. Queries denormalizadas en vistas materializadas (`assembly_view`, `results_view`), cache Redis.
* WebSockets/SSE para quórum y resultados en tiempo real.
* p95 < 200 ms en `vote.open/close`; >10k votos/min con partición por `tenant_id`.
* Sharding por `tenant_id` en Kafka y claves en Redis.
* Workers asíncronos para generación de actas y archivado.

# Resiliencia y consistencia

* **Sagas** para: convocatoria, sesión, votación por ítem, acta. Compensaciones: reenvío de comunicaciones, re-cierre de voto, reintento de firma, rearchivo WORM.
* Retries exponenciales, circuit breakers a externos.
* Modo degradado: si Meet falla, registrar fallback y permitir reanudación.

# Observabilidad

* OTel traces con `tenant_id`, `assembly_id`, `item_id`.
* Métricas: TPS votos, latencia open/close, quorum drift, fallos manuales sin boleta, tiempo a publicación de acta.
* Logs firmados y tamper-evident.

# Tecnología sugerida

* **Runtime**: Kotlin + Spring Boot o Go + chi/fx.
* **DB**: PostgreSQL 15 + pgcrypto + logical decoding (future CDC).
* **Cache**: Redis 7.
* **Bus**: Kafka o NATS JetStream.
* **Docs**: S3 compatible + Glacier; WORM.
* **Infra**: Kubernetes, HPA por QPS y lag de cola.
* **API**: OpenAPI 3.1 + gRPC internos.
* **AuthN**: OIDC/OAuth2 provider externo (Auth Service).

# Decisiones clave

* Google Meet como único VC.
* Manuales marcados e indisociables de boleta.
* Legalidad: cámara ON en acreditación virtual, quórum público, resultados consolidados.
* Outbox para fiabilidad de eventos.
* CQRS light para UX en vivo.

# Backlog técnico inmediato

1. Esquema SQL y migraciones.
2. OpenAPI por módulo.
3. Adapters: MeetPort, DocPort, CompliancePort.
4. Vistas materializadas y SSE para quórum/resultados.
5. Sagas y outbox.
6. Pruebas de carga de voto y latencia.
