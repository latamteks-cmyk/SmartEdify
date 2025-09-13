Estructura monorepo y premisas. Objetivo: entrega rápida, calidad constante, auditoría simple.

# 1) Estructura de carpetas (top-level)

```
smartedify/
├─ apps/                     # Ejecutables (front y servicios)
│  ├─ web-app/               # Web App (RBAC único)
│  ├─ web-soporte/           # NOC/Helpdesk
│  ├─ mobile-app/            # iOS/Android (owner-only)
│  └─ services/              # Microservicios
│     ├─ assembly-service/
│     ├─ auth-service/
│     ├─ user-service/
│     ├─ finance-service/
│     ├─ document-service/
│     ├─ communication-service/
│     ├─ payments-service/
│     ├─ compliance-service/
│     ├─ reservation-service/
│     ├─ maintenance-service/
│     ├─ payroll-service/
│     ├─ certification-service/
│     └─ facilitysecurity-service/
├─ packages/                 # Librerías compartidas (no ejecutables)
│  ├─ core-domain/           # DDD, tipos, errores comunes
│  ├─ security/              # JWT, JWKS, WebAuthn, TOTP helpers
│  ├─ http-kit/              # Middlewares, client, retry, tracing
│  ├─ event-bus/             # Kafka/NATS SDK + outbox/inbox
│  ├─ persistence/           # Repos genéricos, migraciones helpers
│  ├─ validation/            # Esquemas Zod/JSON-Schema
│  ├─ i18n/                  # Mensajes y plantillas
│  └─ ui-kit/                # Componentes UI compartidos (web)
├─ api/                      # Contratos externos
│  ├─ openapi/               # *.yaml por servicio
│  └─ proto/                 # *.proto para gRPC internos
├─ db/                       # Migraciones y seeds
│  ├─ assembly/
│  ├─ auth/
│  └─ ...
├─ infra/                    # Infraestructura declarativa
│  ├─ terraform/             # VPC, KMS, RDS, S3/WORM, CDN
│  ├─ k8s/                   # Helm charts/overlays (dev,stg,prod)
│  ├─ docker/                # Dockerfiles base + compose local
│  └─ gateway/               # Reglas API Gateway/WAF, OIDC
├─ ops/                      # Operaciones y runbooks
│  ├─ runbooks/
│  ├─ sre/                   # Alertas, SLO, dashboards
│  └─ playbooks/             # Respuesta a incidentes
├─ docs/                     # Documentación viva
│  ├─ prd/                   # PRD por servicio
│  ├─ design/                # ADR, diagramas C4/BPMN/Mermaid
│  ├─ api/                   # Docs HTML generadas de OpenAPI
│  └─ legal/                 # Plantillas actas, checklist legal
├─ tools/                    # CLI internas, generadores, linters
├─ .github/                  # CI/CD (Actions), CODEOWNERS, templates
├─ scripts/                  # make, task runners, dev tooling
├─ Makefile                  # or Taskfile.yml
├─ CODEOWNERS
├─ LICENSE
└─ README.md
```

# 2) Plantilla de servicio (apps/services/\*-service)

```
*-service/
├─ cmd/
│  └─ server/                # main.go / main.kt
├─ internal/
│  ├─ app/                   # commands, queries, sagas
│  ├─ domain/                # aggregates, events, policies
│  ├─ adapters/
│  │  ├─ http/               # handlers, routers, dto
│  │  ├─ grpc/               # opcional
│  │  ├─ repo/               # postgres, redis
│  │  ├─ bus/                # kafka/nats
│  │  └─ ext/                # clientes a otros servicios
│  └─ config/                # carga de env, flags
├─ pkg/                      # utilidades específicas del servicio
├─ migrations/               # sql/atlas/flyway
├─ tests/
│  ├─ unit/
│  └─ integration/
├─ api/
│  ├─ openapi.yaml
│  └─ proto/
├─ Dockerfile
├─ helm/                     # chart del servicio
├─ k8s/                      # kustomize overlays
├─ .env.example
└─ README.md
```

# 3) Frontends

```
apps/web-app/                # Monorepo JS/TS (pnpm)
├─ src/
├─ public/
├─ vite.config.ts
└─ package.json

apps/web-soporte/
apps/mobile-app/             # React Native/Flutter
```

# 4) Premisas de creación de archivos

## Naming y layout

* Kebab-case para carpetas (`assembly-service`), PascalCase para tipos, snake\_case en SQL.
* `cmd/server/main.*` como entrypoint único.
* Un handler por archivo. Máx 300 líneas por archivo objetivo.
* DTOs en `adapters/http/dto/*`. No exponer entidades de dominio.

## Contratos primero

* PRs que cambian API deben actualizar `api/openapi/*.yaml` y ejemplos.
* Generar SDKs cliente desde OpenAPI/proto en CI y publicar en `packages/*-sdk`.

## Configuración

* Solo variables env con prefijo por servicio: `ASM_`, `AUTH_`, etc.
* `internal/config/defaults.go|kt` con valores por defecto.
* Plantilla `.env.example` obligatoria.

## Seguridad

* Sin secretos en repo. Usar secretos de CI y vault.
* TLS obligatorio. JWT verificado en gateway y servicio.
* Logs sin PII. Redactar tokens y documentos.

## Persistencia

* Migraciones versionadas en `migrations/`.
* Una transacción por caso de uso.
* Patrón outbox para eventos externos.
* Índices declarados en migraciones.

## Testing

* Cobertura mínima 80% en `internal/app` y `domain`.
* Tests de contrato para HTTP/gRPC con snapshots.
* Pruebas de migraciones en CI.

## Observabilidad

* Tracing OTel con `tenant_id`, `service`, `assembly_id|user_id` cuando aplique.
* Métricas de negocio: votos/min, quorum drift, etc.
* Estructura logs JSON.

## Documentación

* README por servicio con: run local, env, puertos, dependencias.
* ADR en `docs/design/adr/yyyymmdd-title.md`.
* Diagramas Mermaid en `docs/design/diagrams/*.md`.

## Calidad

* Lint y format en pre-commit (`golangci-lint` / `eslint` / `ktlint`).
* Convenciones de commit: Conventional Commits.
* Revisiones obligatorias por CODEOWNERS.

## Versionado y releases

* SemVer por servicio.
* Tags por servicio: `assembly-service/v1.2.3`.
* Changelogs autogenerados desde commits.

## Branching

* `main` protegida.
* `release/*` estabiliza.
* `feat/*`, `fix/*`, `chore/*` por tarea.

## CI/CD

* Workflows por servicio en `.github/workflows/*-service.yml`.
* Jobs: lint → test → build → scan → image → helm-lint → deploy(dev).
* Promoción a `stg`/`prod` con aprobación manual.
* SBOM + SCA (Trivy/Grype).

## Seguridad supply-chain

* Firmar imágenes (cosign).
* Policy admission (Kyverno): no-run-as-root, readOnlyRootFs.
* Escaneo dependencias semanal.

# 5) Makefile (targets estándar)

```
make bootstrap        # instala toolchains locales
make gen              # genera SDKs desde openapi/proto
make lint             # linters todos los paquetes
make test             # unit + integration
make build            # binarios
make docker           # build imagen local
make migrate-up       # migraciones
make run              # docker compose local
make docs             # compila docs API a HTML
```

# 6) Docker Compose local (extracto)

```
services:
  postgres:
    image: postgres:16
    env_file: .env
  redis:
    image: redis:7
  nats:
    image: nats:2
  assembly-service:
    build: ./apps/services/assembly-service
    env_file: apps/services/assembly-service/.env.example
    depends_on: [postgres, redis, nats]
```

# 7) CODEOWNERS (ejemplo)

```
/apps/services/assembly-service/   @team-assembly
/apps/services/auth-service/       @team-auth
/api/openapi/assembly.yaml         @team-assembly @platform
/infra/**                          @platform
```

# 8) Plantillas mínimas

**ADR**

```
# ADR-YYYYMMDD: Título
Contexto
Decisión
Alternativas
Consecuencias
Estado
```

**README servicio**

```
# Assembly Service
Run local, variables, endpoints, decisiones, SLO, contacto equipo.
```

**PR plantilla**

```
Objetivo
Cambios
Checklist: [ ] OpenAPI actualizado [ ] Tests [ ] Migraciones
Riesgos
```

# 9) Línea base por servicio (carpetas obligatorias)

* `api/`, `migrations/`, `internal/app|domain|adapters|config/`, `tests/`, `helm/`, `k8s/`.
* OpenAPI válido, ejemplos en `docs/api/examples/`.
* Alertas SRE definidas en `ops/sre/alerts/*.yaml`.

# 10) Reglas de integración entre servicios

* Comunicación sincrónica vía HTTP/gRPC solo en lectura o validaciones rápidas.
* Escritura y orquestación por eventos (Kafka/NATS) con outbox.
* Idempotencia por `x-request-id` y `event-id`.
* Retries exponenciales, DLQ por servicio.
