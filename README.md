# SmartEdify Auth Service

[![Build Status](https://github.com/smartedify/smartedify-monorepo/workflows/auth-service/badge.svg)](https://github.com/smartedify/smartedify-monorepo/actions)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-90%2B-green.svg)](https://github.com/smartedify/smartedify-monorepo)

Servicio de autenticación y autorización para la plataforma SmartEdify. Proporciona autenticación JWT segura, gestión de sesiones, y soporte multi-tenant para condominios.
## 🆕 Cambios Recientes (2025)

- **Mocks centralizados:** Todos los tests de integración y unitarios usan un mock repository único y alineado con los requisitos de la interfaz.
- **Tests de integración habilitados:** Se corrigieron errores de lógica y visibilidad en los tests del repositorio, ahora todos los tests pasan.
- **Fixes de compilación:** Se agregaron métodos faltantes en interfaces y se alinearon las firmas de funciones según la documentación .kiro/specs.
- **Limpieza de archivos obsoletos:** Se identificaron y listaron archivos y carpetas para eliminación segura, mejorando la higiene del proyecto.
- **Comandos PowerShell para limpieza:** Se prepararon scripts para eliminar archivos basura y automatizar el mantenimiento en Windows.
- **Documentación alineada:** El README y los endpoints reflejan el estado real del código y las tareas completadas.

## 🚀 Características

### ✅ Implementadas

- **Autenticación JWT**: Tokens RS256 con rotación automática
- **Multi-tenant**: Aislamiento completo por condominio
- **Session Management**: Gestión de sesiones con Redis
- **Rate Limiting**: Protección contra ataques de fuerza bruta
- **OpenID Connect**: Compliance con estándares OIDC
- **Password Security**: Hashing bcrypt + políticas de seguridad
- **Account Lockout**: Bloqueo automático tras intentos fallidos
- **Health Checks**: Endpoints de salud para Kubernetes
- **Metrics**: Métricas Prometheus integradas
- **Structured Logging**: Logs estructurados con correlation IDs

### 🧪 Testing y Mocks

### Ejecución de Tests
```powershell
# Tests unitarios y de integración
 go test ./...

# Tests con coverage
 go test -cover ./...

# Tests de integración
 go test -tags=integration ./...

# Benchmark tests
 go test -bench=. ./...
```

### Generación y uso de mocks
```powershell
 go generate ./...
```

### Limpieza de archivos obsoletos (Windows)
```powershell
# Ejecutar desde el directorio raíz del proyecto
 .\scripts\cleanup-project.ps1
```

## 🏗️ Arquitectura

```
auth-service/
├── cmd/
│   └── main.go              # ← Entry point
├── internal/
│   ├── api/                 # ← HTTP handlers
│   ├── config/              # ← Configuration management
│   ├── database/            # ← Database connection & migrations
│   ├── errors/              # ← Custom error types
│   ├── handlers/            # ← Business logic handlers
│   ├── middleware/          # ← HTTP middleware
│   ├── models/              # ← Data models
│   ├── repository/          # ← Data access layer
│   ├── service/             # ← Business logic
│   └── utils/               # ← Utilities
├── migrations/              # ← Database migrations
├── scripts/                 # ← Deployment scripts
├── .kiro/specs/            # ← Specifications & tasks
└── docker-compose.yml       # ← Local development
```

## 🛠️ Tecnologías

- **Framework**: Go Fiber v2
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **JWT**: RS256 with key rotation
- **Monitoring**: Prometheus + Jaeger
- **Testing**: Testify + Mocks
- **Migration**: golang-migrate
- **Validation**: go-playground/validator

## 🚀 Inicio Rápido

### Prerrequisitos

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker (opcional)

### Instalación Local

```bash
# Navegar al directorio del servicio
cd packages/auth-service

# Instalar dependencias
go mod download

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tus configuraciones

# Ejecutar migraciones
go run cmd/migrate/main.go up

# Ejecutar el servicio
go run cmd/main.go
```

### Con Docker

```bash
# Desde el directorio auth-service
docker-compose up -d

# Ver logs
docker-compose logs -f auth-service
```

## ⚙️ Configuración

### Variables de Entorno

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=smartedify_auth
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSL_MODE=require

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_ISSUER=smartedify-auth-service
JWT_AUDIENCE=smartedify-api
JWT_ACCESS_TOKEN_TTL=900    # 15 minutes
JWT_REFRESH_TOKEN_TTL=604800 # 7 days

# Security
ENCRYPTION_KEY=your-32-character-encryption-key
MAX_LOGIN_ATTEMPTS=5
BLOCK_DURATION=1800 # 30 minutes

# Rate Limiting
RATE_LIMIT_PER_IP=1000
RATE_LIMIT_PER_USER=100

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://smartedify.com
```

Ver [.env.example](.env.example) para configuración completa.

## 📡 API Endpoints

### Autenticación

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
POST /api/v1/auth/validate
GET  /api/v1/auth/session
POST /api/v1/auth/reset-password
GET  /api/v1/auth/president         # Nuevo: obtener presidente del condominio
POST /api/v1/auth/lock-user         # Nuevo: bloqueo manual de usuario
GET  /api/v1/auth/tenants           # Nuevo: listado de tenants disponibles
GET  /api/v1/auth/units             # Nuevo: listado de unidades por tenant
```

### OpenID Connect

```http
GET  /.well-known/openid-configuration
GET  /.well-known/jwks.json
POST /oauth/token
GET  /oauth/userinfo
POST /oauth/revoke                  # Nuevo: revocación de token OAuth
```

### Health & Monitoring

```http
GET /health
GET /health/ready
GET /health/live
GET /metrics
GET /audit/logs                     # Nuevo: consulta de logs de auditoría

### 📁 Directorio y Función de Archivos

**cmd/**
- `main.go`: Punto de entrada principal del servicio.

**internal/api/**
- Handlers HTTP: Definen los endpoints y validan la entrada/salida.

**internal/config/**
- Configuración de variables de entorno y parámetros globales.

**internal/database/**
- Conexión a PostgreSQL y gestión de migraciones.

**internal/errors/**
- Tipos de error personalizados y manejo centralizado de errores.

**internal/handlers/**
- Lógica de negocio para cada endpoint (registro, login, bloqueo, etc).

**internal/middleware/**
- Middleware HTTP: autenticación, logging, rate limiting, CORS, etc.

**internal/models/**
- Estructuras de datos y modelos de entidades (User, Tenant, Session, etc).

**internal/repository/**
- Acceso a datos, queries SQL, mock repository para tests.

**internal/service/**
- Lógica de negocio principal, validaciones, reglas y orquestación de procesos.

**internal/utils/**
- Utilidades generales: helpers, validadores, formateadores.

**migrations/**
- Scripts SQL para crear y modificar la base de datos.

**scripts/**
- Scripts de despliegue, migración y limpieza (PowerShell y Bash).

**.kiro/specs/**
- Documentación de requisitos, diseño y tareas.

**docker-compose.yml**
- Orquestación de servicios para desarrollo local.
```

### Ejemplos de Uso

#### Registro de Usuario

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "firstName": "John",
    "lastName": "Doe",
    "tenantId": "condo-123",
    "unitId": "apt-456"
  }'
```

#### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "tenantId": "condo-123",
    "unitId": "apt-456"
  }'
```

#### Validar Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🧪 Testing

```bash
# Tests unitarios
go test ./...

# Tests con coverage
go test -cover ./...

# Tests de integración
go test -tags=integration ./...

# Benchmark tests
go test -bench=. ./...
```

### Coverage Actual

- **Config Package**: 90.8%
- **Repository Package**: Mocks + Unit tests
- **Database Package**: 59.3%
- **Overall Target**: >90%

## 🔒 Seguridad

### Características de Seguridad

- **Password Hashing**: bcrypt con cost factor 12
- **JWT Security**: RS256, rotación de keys, short-lived tokens
- **Rate Limiting**: 5 intentos de login por minuto por IP
- **Account Lockout**: Bloqueo tras 5 intentos fallidos
- **Session Security**: Tokens en Redis con TTL
- **Multi-tenant Isolation**: Aislamiento completo por tenant
- **CORS Protection**: Configuración restrictiva
- **Security Headers**: HSTS, CSP, X-Frame-Options

### Políticas de Contraseña

- Mínimo 8 caracteres
- Al menos 1 mayúscula
- Al menos 1 minúscula  
- Al menos 1 número
- Al menos 1 carácter especial
- No reutilización de últimas 5 contraseñas

## 📊 Monitoreo

### Métricas Prometheus

- `auth_requests_total`: Total de requests por endpoint
- `auth_request_duration_seconds`: Duración de requests
- `auth_active_sessions`: Sesiones activas
- `auth_failed_logins_total`: Intentos de login fallidos
- `auth_tokens_issued_total`: Tokens emitidos

### Health Checks

- `/health`: Estado general del servicio
- `/health/ready`: Readiness probe (K8s)
- `/health/live`: Liveness probe (K8s)

### Logs Estructurados

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "User login successful",
  "correlationId": "req-123-456",
  "userId": "user-789",
  "tenantId": "condo-123",
  "ipAddress": "192.168.1.1",
  "duration": 150
}
```

## 🚀 Deployment

### Docker

```bash
# Build image
docker build -t smartedify/auth-service:latest .

# Run container
docker run -p 8080:8080 \
  -e DB_HOST=postgres \
  -e REDIS_HOST=redis \
  smartedify/auth-service:latest
```

### Kubernetes

```bash
# Deploy to K8s
kubectl apply -f k8s/

# Check status
kubectl get pods -l app=auth-service

# View logs
kubectl logs -f deployment/auth-service
```

### Helm Chart

```bash
# Install with Helm
helm install auth-service ./helm/auth-service \
  --set image.tag=v1.0.0 \
  --set database.host=postgres.default.svc.cluster.local
```

## 🔧 Desarrollo

### Estructura de Código

- **Handlers**: Lógica HTTP y validación de entrada
- **Services**: Lógica de negocio
- **Repositories**: Acceso a datos
- **Models**: Estructuras de datos
- **Middleware**: Funcionalidad transversal

### Convenciones

- **Naming**: camelCase para JSON, snake_case para DB
- **Errors**: Errores tipados con códigos específicos
- **Logging**: Structured logging con correlation IDs
- **Testing**: Table-driven tests, mocks para dependencias

### Comandos Útiles

```bash
# Generar mocks
go generate ./...

# Linting
golangci-lint run

# Format code
go fmt ./...

# Update dependencies
go mod tidy

# Security scan
gosec ./...
```

## 📚 Documentación

- [Especificaciones](.kiro/specs/smartedify-auth-service/requirements.md)
- [Diseño de Arquitectura](.kiro/specs/smartedify-auth-service/design.md)
- [Plan de Implementación](.kiro/specs/smartedify-auth-service/tasks.md)
- [API Documentation](docs/api.md)
- [Database Schema](docs/schema.md)

### Referencias de limpieza y mantenimiento
- [Plan de limpieza](../../CLEANUP_PLAN.md)
- [Estado final de limpieza](../../FINAL_CLEANUP_STATUS.md)
- [Resumen de reorganización](../../REORGANIZATION_SUMMARY.md)

## 🤝 Contribución

1. Crear feature branch desde `main`
2. Implementar cambios con tests
3. Asegurar >90% coverage
4. Ejecutar linting y security checks
5. Crear PR con descripción detallada

## 🗒️ Notas
- El proyecto está alineado con los requisitos y tareas de `.kiro/specs`.
- Todos los endpoints y lógica de negocio están cubiertos por tests y mocks actualizados.
- La documentación y scripts de mantenimiento están actualizados para facilitar la gestión y evolución del servicio.

## 📄 Licencia

MIT License - ver [LICENSE](../../LICENSE) para detalles.

---

**Auth Service** - Autenticación segura y escalable para SmartEdify 🔐