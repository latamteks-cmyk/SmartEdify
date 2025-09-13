### Documento de Diseño y Arquitectura — **User Service** (SmartEdify)

---

# 1. Especificaciones de Diseño

## 1.1 Alcance

El **User Service** administra la identidad interna de usuarios en SmartEdify. No reemplaza a **Auth Service**, sino que provee información de perfil, roles, pertenencia a tenant y estados de usuario. Es la **fuente de verdad** para:

* Datos maestros de personas (propietarios, moderadores, secretarios, admins de condominio, soporte local, NOC).
* Relación usuario ↔ tenant (condominio).
* Roles y permisos de negocio (RBAC/ABAC).
* Estado de cuenta de usuario (activo, bloqueado, eliminado).
* Datos de contacto para notificaciones (email, teléfono, dirección física).
* Preferencias de comunicación.
* Auditoría de cambios de perfil.

## 1.2 Requisitos funcionales

* CRUD de usuarios bajo contexto `tenant_id`.
* Asignación y revocación de roles.
* Gestión de relaciones multi-tenant (usuario puede pertenecer a varios condominios).
* APIs para otros servicios: Assembly, Finance, Communication, etc.
* Búsqueda avanzada (por nombre, email, DNI, tenant).
* Importación y exportación de padrones.
* Desactivación/reactivación de usuarios con registro de motivo.
* Sincronización opcional con Auth Service (para login).

## 1.3 Requisitos no funcionales

* SLA 99.9%.
* Latencia de lectura p95 < 100 ms.
* Escalabilidad a 500k usuarios y 10k tenants.
* Auditoría inmutable.
* Cumplimiento LOPDP y GDPR (data minimization, derecho al olvido).
* API versionada, OpenAPI 3.1.

## 1.4 Modelo de datos lógico (simplificado)

* `users(id, global_id, nombre, apellido, email, phone, documento_tipo, documento_num, status, created_at, updated_at)`
* `tenants(id, nombre, direccion, ruc, created_at)`
* `user_tenants(user_id, tenant_id, rol, coeficiente, estado)`
* `roles(id, nombre, scope, descripcion)`
* `preferences(user_id, lang, notif_email, notif_sms, notif_push)`
* `audit_user(id, user_id, actor_id, action, changes_json, ts)`

Índices:

* `users(email)`, `users(documento_tipo, documento_num)` únicos.
* `user_tenants(user_id, tenant_id)` único.
* `audit_user(user_id, ts)`.

---

# 2. Arquitectura del Sistema

## 2.1 Vista de contenedores

```mermaid
flowchart LR
  subgraph Clients
    WEB[Web App (RBAC)]
    MOB[App Móvil]
    NOC[Web Soporte]
  end

  APIGW[API Gateway + WAF]
  USVC[User Service]
  AUTH[Auth Service]
  DOC[Document Service]
  COMM[Communication Service]
  BUS[Event Bus (Kafka/NATS)]
  DB[(PostgreSQL)]
  CACHE[(Redis)]
  AUDIT[(WORM Storage)]

  WEB --> APIGW --> USVC
  MOB --> APIGW --> USVC
  NOC --> APIGW --> USVC

  USVC --> DB
  USVC --> CACHE
  USVC --> AUTH
  USVC --> DOC
  USVC --> COMM
  USVC --> BUS
  USVC --> AUDIT
```

## 2.2 Componentes internos

* **API Layer**: controladores REST/gRPC, validación de entrada, scopes.
* **User Core**: lógica de negocio (creación, actualización, roles, estados).
* **Tenant Manager**: gestión multi-tenant, coeficientes, import/export.
* **Role Engine**: asignación de roles RBAC/ABAC.
* **Preferences**: notificaciones y settings de idioma.
* **Repositories**: Postgres, Redis (cache de perfiles frecuentes).
* **Audit Logger**: escribe en tabla + WORM.
* **Event Publisher**: dispara `user.created`, `user.updated`, `user.deactivated`.

## 2.3 Integraciones

* **Auth Service**: sincroniza claims (`roles`, `tenant_id`, `status`).
* **Document Service**: guarda documentos de identidad, poderes.
* **Communication Service**: usa datos de contacto y preferencias.
* **Event Bus**: publica cambios de usuarios para Assembly, Finance, Payroll.

## 2.4 Resiliencia

* Idempotencia en `user.create` por `global_id` (ej.: DNI).
* Circuit breakers hacia Auth y Comm.
* Auditoría WORM evita borrado accidental de historial.

---

# 3. Documentación de la API — User Service

## Endpoints principales

Base URL: `https://api.smartedify.com/api/user/v1`

### Usuarios

* `POST /users`
  Crea usuario.
  Body: `{nombre, apellido, email, phone, documento_tipo, documento_num}`
* `GET /users/{id}`
  Devuelve usuario.
* `GET /users`
  Filtros: `tenantId, email, documento_num, status`.
* `PATCH /users/{id}`
  Actualiza datos de perfil.
* `DELETE /users/{id}`
  Desactiva usuario (soft delete).

### Tenants y membresías

* `POST /users/{id}/tenants`
  Asigna usuario a tenant con rol.
* `GET /users/{id}/tenants`
  Lista membresías.
* `PATCH /users/{id}/tenants/{tenantId}`
  Cambia rol/coeficiente.
* `DELETE /users/{id}/tenants/{tenantId}`
  Revoca membresía.

### Roles

* `GET /roles`
  Lista de roles soportados.
* `POST /users/{id}/roles`
  Asigna rol adicional.
* `DELETE /users/{id}/roles/{role}`
  Revoca rol.

### Preferencias

* `GET /users/{id}/preferences`
* `PATCH /users/{id}/preferences`

### Auditoría

* `GET /users/{id}/audit`
  Devuelve historial de cambios.

### Import/Export

* `POST /tenants/{tenantId}/users/import` (sube CSV/Excel → Document Service).
* `GET /tenants/{tenantId}/users/export` (CSV).

---

## OpenAPI 3.1 (extracto clave)

```yaml
openapi: 3.1.0
info:
  title: User Service API
  version: "1.0.0"
servers:
  - url: https://api.smartedify.com/api/user/v1
paths:
  /users:
    post:
      summary: Crear usuario
      security: [{oauth2: [user:write]}]
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UserCreate'
      responses:
        "201": {description: Creado, content: {application/json: {schema: {$ref: '#/components/schemas/User'}}}}
    get:
      summary: Listar usuarios
      security: [{oauth2: [user:read]}]
      parameters:
        - {in: query, name: tenantId, schema: {type: string}}
        - {in: query, name: email, schema: {type: string}}
        - {in: query, name: documento_num, schema: {type: string}}
        - {in: query, name: status, schema: {type: string, enum: [active, inactive, deleted]}}
      responses:
        "200": {description: OK, content: {application/json: {schema: {$ref: '#/components/schemas/PageUsers'}}}}
  /users/{id}:
    get:
      summary: Obtener usuario
      parameters: [{$ref: '#/components/parameters/userId'}]
      responses:
        "200": {description: OK, content: {application/json: {schema: {$ref: '#/components/schemas/User'}}}}
    patch:
      summary: Actualizar usuario
      parameters: [{$ref: '#/components/parameters/userId'}]
      requestBody:
        content:
          application/json:
            schema: {$ref: '#/components/schemas/UserUpdate'}
      responses: {"200": {description: OK}}
    delete:
      summary: Desactivar usuario
      parameters: [{$ref: '#/components/parameters/userId'}]
      responses: {"204": {description: Desactivado}}
components:
  parameters:
    userId: {in: path, name: id, required: true, schema: {type: string}}
  schemas:
    UserCreate:
      type: object
      required: [nombre, apellido, email, documento_tipo, documento_num]
      properties:
        nombre: {type: string}
        apellido: {type: string}
        email: {type: string, format: email}
        phone: {type: string}
        documento_tipo: {type: string, enum: [dni, ce, pasaporte]}
        documento_num: {type: string}
    UserUpdate:
      type: object
      properties:
        nombre: {type: string}
        apellido: {type: string}
        phone: {type: string}
        status: {type: string, enum: [active, inactive]}
    User:
      type: object
      properties:
        id: {type: string}
        nombre: {type: string}
        apellido: {type: string}
        email: {type: string}
        phone: {type: string}
        documento_tipo: {type: string}
        documento_num: {type: string}
        status: {type: string}
        created_at: {type: string, format: date-time}
        updated_at: {type: string, format: date-time}
    PageUsers:
      type: object
      properties:
        items: {type: array, items: {$ref: '#/components/schemas/User'}}
        total: {type: integer}
```
