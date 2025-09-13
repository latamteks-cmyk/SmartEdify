# ✅ **SMARTEDIFY v.0 – DOCUMENTO DE DESPLIEGUE Y MANTENIMIENTO**  
## **Auth Service — Guía de Implementación, Operación y Soporte para DevOps / SRE**

> **Versión**: v.1.0 (Definitiva)  
> **Fecha**: Abril 2025  
> **Autor**: DevOps / SRE Lead, SmartEdify  
> **Aprobado por**: Software Architect, Head of Engineering  

---

## ✅ **1. Guía de Implementación (Deployment Guide)**

> *“Cómo desplegar, configurar y validar Auth Service en entornos de desarrollo, staging y producción.”*

### 1.1 Requisitos Previos

| Componente | Requisito | Detalle |
|----------|---------|---------|
| **Infraestructura** | AWS Account | Cuenta con permisos para crear: VPC, EKS, RDS, CloudHSM, Secrets Manager, S3, Route53 |
| **Red** | VPC Privada | Subredes públicas (EKS) + privadas (DB, Redis, Kafka). NAT Gateway obligatorio. |
| **Seguridad** | HSM | AWS CloudHSM activado con claves RSA-256. Acceso restringido por IAM roles. |
| **DNS** | Dominio propio | `auth.smartedify.dev` (dev), `auth.smartedify.com` (prod) |
| **Secrets** | Vault / Secrets Manager | Almacenamiento seguro de: claves privadas JWT, claves de cifrado AES, tokens de WhatsApp, credenciales de Kafka. |
| **CI/CD** | GitHub Actions | Repositorio público `github.com/smartedify/auth-service` con workflows configurados. |
| **Monitoring** | Prometheus + Grafana + Jaeger | Instalados en cluster. Alertmanager configurado con Slack/email. |
| **Compliance** | Certificado TLS | Certificado SSL/TLS válido emitido por ACM (AWS Certificate Manager) |

### 1.2 Arquitectura de Despliegue (Producción)

```mermaid
graph LR
    subgraph "VPC Prod - us-east-1"
        A[API Gateway] --> B[Ingress Controller<br>(NGINX Ingress)]
        B --> C[EKS Cluster<br>6 Nodes m6i.large]
        C --> D[Auth Service Pod<br>(ReplicaSet: 4)]
        C --> E[Redis Cluster<br>3 shards, 1 replica]
        C --> F[PostgreSQL RDS<br>db.r6g.xlarge, Multi-AZ]
        C --> G[Kafka Cluster<br>3 brokers, 3 replicas]
        D --> H[AWS CloudHSM<br>Claves JWT & RSA]
        D --> I[AWS Secrets Manager<br>Configuraciones, Tokens]
        D --> J[IPFS Pinning<br>Pinata]
        D --> K[Event Bus<br>Kafka]
        L[CloudWatch] --> D
        M[Prometheus] --> D
        N[Jaeger] --> D
    end

    O[Client] -->|HTTPS + DPoP| A
    P[Partner App] -->|OAuth 2.1 + MTLS| A
```

### 1.3 Flujo de Despliegue Automatizado (CI/CD)

#### 🔄 Pipeline GitHub Actions (`ci.yml`)

```yaml
name: CI/CD Auth Service
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install dependencies
        run: npm ci
      - name: Run unit tests
        run: npm run test:unit
      - name: Run integration tests
        run: npm run test:integration
      - name: Check code coverage
        run: npx nyc report --reporter=text-lcov > coverage.lcov
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.lcov

  build-and-push:
    needs: test
    runs-on: ubuntu-latest
    environment: production
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - name: Login to ECR
        uses: aws-actions/amazon-ecr-login@v2
      - name: Build Docker image
        run: |
          docker build -t ${{ secrets.ECR_REPOSITORY }}:${{ github.sha }} .
      - name: Push to ECR
        run: |
          docker push ${{ secrets.ECR_REPOSITORY }}:${{ github.sha }}

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    environment: production
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Update Helm values
        run: |
          sed -i "s/image-tag: latest/image-tag: ${{ github.sha }}/g" helm/values-prod.yaml
      - name: Deploy to EKS
        uses: azure/k8s-deploy@v4
        with:
          action: deploy
          namespace: auth-service
          manifests: |
            helm/templates/deployment.yaml
            helm/templates/service.yaml
            helm/templates/ingress.yaml
          images: |
            ${{ secrets.ECR_REPOSITORY }}:${{ github.sha }}
      - name: Validate deployment
        run: |
          kubectl rollout status deployment/auth-service -n auth-service --timeout=300s
          curl -f https://auth.smartedify.com/.well-known/jwks.json
```

### 1.4 Configuración Clave por Entorno

| Parámetro | Desarrollo (dev) | Producción (prod) |
|----------|------------------|-------------------|
| **JWT TTL (access)** | 5 min | 1 h |
| **JWT TTL (refresh)** | 1 día | 7 días |
| **Argon2id Memory** | 32 MB | 64 MB |
| **Argon2id Iterations** | 2 | 3 |
| **Rate Limit** | 100 req/min/user | 10 req/min/user |
| **MFA por defecto** | TOTP | WhatsApp OTP |
| **DPoP** | Opcional | Obligatorio |
| **MTLS** | Desactivado | Obligatorio (entre servicios) |
| **JWKS Rotación** | Manual | Automática cada 90 días |
| **IPFS Pinning** | Local (Docker) | Pinata (servicio externo) |
| **HSM** | Simulado (Mock) | AWS CloudHSM |

> ✅ **Nota crítica**: En producción, **no se permite desactivar DPoP o MTLS**.

### 1.5 Procedimiento de Despliegue Manual (Emergencia)

> Solo usar si CI/CD falla.

```bash
# 1. Obtener la imagen del registro
docker pull <ECR_URL>/auth-service:v1.2.3

# 2. Actualizar el valor en Helm
helm upgrade auth-service ./helm \
  --namespace auth-service \
  --set image.tag=v1.2.3 \
  --set config.hsm.enabled=true \
  --set config.dpop.required=true \
  --values ./helm/values-prod-overrides.yaml

# 3. Validar rollout
kubectl get pods -n auth-service -w

# 4. Verificar endpoints
curl -k https://auth.smartedify.com/health
curl https://auth.smartedify.com/.well-known/jwks.json

# 5. Validar trazas en Jaeger
# Ir a: https://jaeger.smartedify.com/search?service=auth-service
```

---

## ✅ **2. Documentación del Usuario (User Manual)**

> *“Guía para administradores, ingenieros y operadores que interactúan con Auth Service.”*

### 2.1 Acceso al Sistema

| Recurso | URL | Acceso |
|--------|-----|--------|
| Dashboard de Monitorización | `https://grafana.smartedify.com` | SSO con Okta / LDAP |
| Tracing (Jaeger) | `https://jaeger.smartedify.com` | Mismo acceso que Grafana |
| Endpoint JWKS | `https://auth.smartedify.com/.well-known/jwks.json` | Público (sin autenticación) |
| API Docs (OpenAPI) | `https://docs.auth.smartedify.com` | Público |
| Sandbox | `https://sandbox.auth.smartedify.com` | Acceso con cuenta de prueba |
| Logs (CloudWatch) | `https://us-east-1.console.aws.amazon.com/cloudwatch/home` | Rol IAM necesario |

### 2.2 Gestión de Secretos

> ⚠️ **Nunca editar secretos directamente en AWS Secrets Manager sin auditoría.**

| Operación | Procedimiento |
|----------|---------------|
| **Cambiar clave JWT** | 1. Generar nueva clave en HSM.<br>2. Subir clave pública a Secrets Manager como `jwt-public-key-v2`.<br>3. Actualizar Helm: `config.jwt.keyVersion: v2`.<br>4. Desplegar. La clave antigua sigue vigente 7 días. |
| **Rotar token de WhatsApp** | 1. Generar nuevo token en Twilio/Mercado Pago.<br>2. Actualizar secreto: `whatsapp-api-token`.<br>3. Reiniciar pod de Auth Service. |
| **Revisar permisos IAM** | Usar AWS IAM Access Analyzer para detectar accesos excesivos al HSM o Secrets Manager. |

### 2.3 Monitoreo y Alertas (Grafana)

#### 🔍 Dashboards Clave

| Dashboard | Objetivo | Alertas Críticas |
|----------|----------|------------------|
| **Auth Service Overview** | Estado general del servicio | - Latencia > 1s<br>- Error rate > 5%<br>- Conexiones Redis caídas |
| **Authentication Metrics** | Login por canal | - >5 intentos fallidos/min desde misma IP<br>- Uso inusual de WebAuthn |
| **Token Health** | Validez y rotación | - JWKS no actualizado en 24h<br>- Token expirado pero aún usado |
| **ARCO Requests** | Cumplimiento legal | - Solicitudes ARCO sin validación MFA<br>- Intentos de borrado masivo |
| **Kafka Consumer Lag** | Eventos pendientes | - Lag > 1000 mensajes → alerta de sincronización rota |

#### 📢 Alertas Definidas (Alertmanager)

| Alerta | Condición | Acción |
|-------|-----------|--------|
| `AuthServiceDown` | HTTP 5xx > 5% en 5 min | Notificación por Slack + Email + SMS |
| `JWKSRotationFailed` | No hay cambio en JWKS en 89 días | Email a Security Team |
| `DPoPValidationFailures` | >100 rechazos DPoP/hora | Investigar posible ataque de replay |
| `RedisConnectionLost` | >5 conexiones perdidas en 1 min | Escalar a DBA |
| `HighPasswordResetRequests` | >50 solicitudes de recuperación/hora | Posible ataque de fuerza bruta → bloquear IPs |

### 2.4 Flujos de Soporte

| Caso | Procedimiento |
|------|---------------|
| **Usuario no puede iniciar sesión** | 1. Verificar en Grafana: ¿login fallido por “invalid password” o “MFA required”?<br>2. Si es MFA: confirmar que el número de teléfono está registrado.<br>3. Si es contraseña: usar endpoint `/v1/auth/forgot-password` para enviar nuevo link. |
| **Acta digital no se genera** | 1. Verificar logs de Auth Service: ¿error al llamar a Compliance Service?<br>2. Verificar si IPFS tiene espacio disponible.<br>3. Revisar certificado de firma en HSM. |
| **Solicitud ARCO no responde** | 1. Confirmar que el usuario está autenticado y tiene permiso.<br>2. Verificar que el evento `arco.requested` se emitió en Kafka.<br>3. Revisar bitácora WORM: ¿se registró la acción? |
| **Error 403 “insufficient_scope”** | 1. Validar que el token contenga los claims correctos (`tenant_id`, `unit_id`).<br>2. Verificar en PostgreSQL que el rol del usuario esté activo en `user_unit_roles`. |

---

## ✅ **3. Manuales de Operaciones**

> *“Procedimientos operativos para mantenimiento, escalado y recuperación ante desastres.”*

### 3.1 Mantenimiento Diario

| Tarea | Frecuencia | Herramienta | Responsable |
|-------|------------|-------------|-------------|
| Verificar estado de HSM | Diario | `aws cloudhsm describe-clusters` | SRE |
| Validar integridad de bitácora | Diario | Script Python `audit-chain-validator.py` | SRE |
| Limpiar sesiones expiradas en Redis | Diario | `redis-cli --scan --pattern "refresh:*" | xargs redis-cli DEL` | Automation |
| Revisar logs de seguridad | Diario | CloudWatch Insights | SRE |
| Verificar certificados TLS | Semanal | `openssl s_client -connect auth.smartedify.com:443` | SRE |

### 3.2 Escalado Automático

| Métrica | Umbral | Acción |
|--------|--------|--------|
| CPU > 70% durante 5 min | Scale out | Aumentar replicas de Auth Service de 4 a 6 |
| Memoria > 80% durante 5 min | Scale out | Aumentar replicas de Auth Service de 4 a 6 |
| Latencia de login > 1.2s | Scale out | Aumentar replicas de Auth Service de 4 a 6 |
| Redis connections > 80% | Scale out | Aumentar shards de Redis de 3 a 5 |
| Kafka lag > 5000 messages | Scale out | Aumentar consumidores de Kafka (repartir carga) |

> ✅ **No se escala horizontalmente por número de usuarios. Se escala por carga de procesamiento real.**

### 3.3 Plan de Recuperación ante Desastres (DRP)

#### ❗ Escenario 1: Fallo total del Auth Service

| Paso | Acción |
|------|--------|
| 1. Activar DRP | Notificar a equipo de emergencia. |
| 2. Verificar estado de dependencias | ¿Redis está caído? ¿HSM no responde? ¿Kafka fuera de línea? |
| 3. Restaurar base de datos | Restaurar RDS PostgreSQL desde backup más reciente (últimas 24h). |
| 4. Restaurar Redis | Recrear cluster desde snapshot. |
| 5. Restaurar HSM | Recuperar claves desde copia segura (fuera de AWS). |
| 6. Desplegar versión estable | Forzar rollback a última versión funcional (v1.1.0). |
| 7. Redirigir tráfico | Cambiar DNS de `auth.smartedify.com` a IP de cluster de respaldo (si existe). |
| 8. Comunicar incidente | Notificar a todos los tenants: “Estamos restaurando el servicio. Por favor, no intente volver a loguearse.” |

#### ❗ Escenario 2: Exposición de clave privada JWT

| Paso | Acción |
|------|--------|
| 1. Detectar | Alerta de anomalía en JAEGGER o análisis de logs. |
| 2. Revocar inmediato | Eliminar clave comprometida de HSM. |
| 3. Emitir nueva clave | Generar nueva clave RSA en HSM. |
| 4. Publicar nueva JWKS | Actualizar endpoint `.well-known/jwks.json` con nueva clave. |
| 5. Revocar todos los tokens | Ejecutar script: `revoke-all-tokens-by-key-id <old-kid>` → borra todos los refresh tokens asociados. |
| 6. Notificar a usuarios | Enviar mensaje: “Su sesión ha sido cerrada por seguridad. Inicie sesión nuevamente.” |
| 7. Auditoría forense | Analizar logs: ¿Quién accedió a la clave? ¿Desde dónde? |

#### ❗ Escenario 3: Ataque de DPoP Replay

| Paso | Acción |
|------|--------|
| 1. Detectar | Alerta: `DPoPValidationFailures > 100/hour` |
| 2. Bloquear IPs sospechosas | Agregar IPs a lista negra en WAF. |
| 3. Verificar origen | ¿Son bots? ¿Son clientes maliciosos? |
| 4. Revisar implementación cliente | ¿Los partners están usando nonce correctamente? |
| 5. Reforzar política | Cambiar TTL de DPoP proof de 5 min a 1 min. |
| 6. Notificar a partners | Enviar email: “Su integración presenta intentos de replay. Por favor, revise su implementación de DPoP.” |

### 3.4 Copias de Seguridad y Retención

| Recurso | Frecuencia | Retención | Método |
|--------|------------|-----------|--------|
| PostgreSQL (RDS) | Diaria | 35 días | Snapshot automático |
| Redis | Diaria | 7 días | Backup RDB en S3 |
| HSM Keys | Mensual | 1 año | Exportación criptográfica segura (fuera de AWS) |
| Bitácora Inmutable | Continua | Indefinida | IPFS + WORM DB (PostgreSQL con trigger) |
| Logs (CloudWatch) | Continua | 180 días | Archivado en Glacier Deep Archive |

> ✅ **Prueba de restauración**: Realizada trimestralmente.  
> ✅ **Resultado esperado**: Restauración completa en ≤ 45 minutos.

### 3.5 Checklist de Cierre de Cambio (Change Freeze)

Antes de cualquier despliegue:

| Verificación | Estado |
|--------------|--------|
| ✅ El cambio está documentado en Jira | ☐ |
| ✅ Todas las pruebas pasaron (unit, e2e, security) | ☐ |
| ✅ No hay cambios en dependencias externas (HSM, WhatsApp API) | ☐ |
| ✅ Se notificó a stakeholders (Legal, Product) | ☐ |
| ✅ Hay un plan de rollback definido | ☐ |
| ✅ El cambio se despliega en horario de baja actividad (02:00–04:00 UTC) | ☐ |
| ✅ Se monitorea durante 1 hora post-despliegue | ☐ |

> ✅ **Firma del responsable**: _________________________  
> **Fecha**: ___/___/2025

---

## ✅ **CONCLUSIÓN FINAL — DECLARACIÓN DEL DEVOPS / SRE**

> “No gestionamos servidores. Gestionamos confianza.  
>   
> Auth Service es la puerta de entrada a la ley en LatAm.  
>   
> Si falla, las comunidades pierden su voz.  
>   
> Por eso, cada línea de código, cada alerta, cada backup, cada rotación de clave…  
>   
> …es una promesa cumplida.  
>   
> Este documento no es un manual. Es un juramento técnico.”

---
