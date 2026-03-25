# Ecommerce2

E-commerce completo construido con **Go** (backend) y **React + TypeScript** (frontend), usando [go-without-magic](https://github.com/JoX23/go-without-magic) como base del microservicio.

---

## Por qué go-without-magic como base del backend

Construir el backend de un e-commerce implica decisiones arquitectónicas que se pagan (o se cobran) durante toda la vida del proyecto. La elección de [go-without-magic](https://github.com/JoX23/go-without-magic) no fue casual — tiene ventajas concretas que se sintieron desde el primer día de desarrollo:

### 1. Clean Architecture lista para usar, sin lock-in

El template implementa la arquitectura en capas correctamente desde el arranque:

```
Domain → Service → Handler → Repository
```

Cada entidad del e-commerce (User, Product, Order) siguió exactamente este flujo. Agregar lógica de negocio compleja como validar stock antes de crear una orden, o persistir un snapshot del nombre del producto al momento de la compra, se hizo en la capa `service` sin tocar los handlers ni los repositorios.

### 2. Codegen: de YAML a 9 archivos sin boilerplate

La herramienta más valiosa del proyecto. Definir una entidad en YAML y ejecutar el generador produce en segundos todo el scaffolding necesario:

```bash
go run ./tools/codegen/ generate --schema product.yaml
```

Para este e-commerce se generaron **User**, **Product** y **Order** con sus respectivos:
- Entidad de dominio con invariantes
- Interfaz de repositorio
- Tipos de error del dominio
- Servicio base
- Handler HTTP
- Repositorio en memoria (desarrollo)
- Repositorio PostgreSQL (producción)
- Stub gRPC

Sin el codegen, esto serían ~685 líneas de boilerplate repetitivo por entidad. Con él, el tiempo se invirtió en la lógica real del e-commerce.

### 3. Repositorio dual: memoria en desarrollo, Postgres en producción

El template separa limpiamente las implementaciones `memory` y `postgres` de cada repositorio. Durante el desarrollo y los tests se usa el repositorio en memoria (sin depender de Docker ni bases de datos). En producción se activa PostgreSQL cambiando una variable de entorno. Esta separación permitió escribir y pasar los 17 tests de servicio sin infraestructura externa.

### 4. Middleware composable listo para producción

El stack de middlewares incluye RequestID, logging estructurado (Zap), tracing (OpenTelemetry), métricas (Prometheus) y recovery de panics. Se extendió con un middleware JWT propio sin modificar nada del sistema base — simplemente se agregó al chain.

### 5. Wiring explícito — sin magia de inyección de dependencias

Toda la construcción del grafo de dependencias es visible en `main.go`:

```go
pool       := postgres.NewPool(ctx, cfg.Database)
userRepo   := memory.NewUserRepository()
tokenSvc   := service.NewTokenService(cfg.Auth.JWTSecret)
userSvc    := service.NewUserService(userRepo, logger)
authHandler := handler.NewAuthHandler(userSvc, tokenSvc, logger)
authHandler.RegisterRoutes(mux)
```

Cuando algo falla en producción, no hay framework internals que debuggear. El código es exactamente lo que se lee.

### 6. Producción-ready desde el día 1

Características que normalmente toman semanas implementar vinieron incluidas: graceful shutdown, health checks (`/healthz`), Dockerfile multi-stage, Docker Compose, GitHub Actions CI, y configuración 12-factor vía Viper. Para este proyecto se extendió con migraciones versionadas (`golang-migrate`), cookie HttpOnly para JWT, y CORS configurable.

---

## Stack

| Capa | Tecnología |
|---|---|
| Backend | Go 1.22, `net/http` (sin framework externo) |
| Base de dominio | [go-without-magic](https://github.com/JoX23/go-without-magic) |
| Autenticación | JWT (HS256) + cookie HttpOnly |
| Base de datos | PostgreSQL 15 via `pgx/v5` |
| Migraciones | `golang-migrate` |
| Frontend | React 18, TypeScript (strict), Vite |
| Estilos | Tailwind CSS |
| Routing frontend | React Router v6 |
| Containerización | Docker + Docker Compose |

---

## Funcionalidades

- Registro y login de usuarios con JWT (cookie HttpOnly + Bearer token)
- Catálogo de productos con SKU, precio, stock y estado (`draft`, `published`, `archived`)
- Paginación en listados (`?page=1&limit=20`)
- Carrito de compras persistido en localStorage
- Creación de órdenes con validación de stock y estado del producto
- Snapshot del nombre del producto en cada línea de orden
- Historial de órdenes por usuario
- Redirección automática a la orden recién creada tras el checkout
- CORS restringido a orígenes configurados

---

## Endpoints

| Método | Ruta | Auth | Descripción |
|---|---|---|---|
| `POST` | `/auth/register` | — | Registro de usuario |
| `POST` | `/auth/login` | — | Login, retorna JWT |
| `POST` | `/auth/logout` | — | Borra la cookie de sesión |
| `GET` | `/auth/me` | ✓ | Perfil del usuario autenticado |
| `GET` | `/products` | — | Listar productos publicados (paginado) |
| `GET` | `/products/{id}` | — | Detalle de producto |
| `POST` | `/products` | ✓ | Crear producto |
| `PUT` | `/products/{id}` | ✓ | Actualizar producto |
| `POST` | `/orders` | ✓ | Crear orden desde items del carrito |
| `GET` | `/orders` | ✓ | Listar órdenes del usuario (paginado) |
| `GET` | `/orders/{id}` | ✓ | Detalle de orden |

---

## Estructura del proyecto

```
ecommerce2/
├── backend/                        # Go backend (basado en go-without-magic)
│   ├── cmd/server/main.go          # Entrypoint + wiring explícito
│   ├── internal/
│   │   ├── domain/                 # Entidades, errores, interfaces, paginación
│   │   ├── service/                # Lógica de negocio + TokenService
│   │   ├── handler/http/           # Handlers HTTP (auth, products, orders)
│   │   ├── repository/
│   │   │   ├── memory/             # Repositorios en memoria (desarrollo)
│   │   │   └── postgres/           # Repositorios PostgreSQL (producción)
│   │   ├── middleware/             # JWT auth, CORS, logging, tracing
│   │   └── config/                 # Configuración 12-factor (Viper)
│   ├── migrations/                 # Migraciones SQL versionadas (golang-migrate)
│   ├── tools/codegen/              # Generador de entidades desde YAML
│   ├── user.yaml                   # Schema de la entidad User
│   ├── product.yaml                # Schema de la entidad Product
│   ├── order.yaml                  # Schema de la entidad Order
│   └── Dockerfile
├── frontend/                       # React + TypeScript + Tailwind
│   ├── src/
│   │   ├── api/client.ts           # Fetch wrapper con Bearer token automático
│   │   ├── context/
│   │   │   ├── AuthContext.tsx     # Estado de auth + localStorage
│   │   │   └── CartContext.tsx     # Carrito en memoria + localStorage
│   │   ├── pages/                  # Login, Register, Products, Cart, Orders
│   │   ├── components/             # Navbar, ProductCard, ProtectedRoute
│   │   ├── types/index.ts          # Interfaces TypeScript
│   │   └── utils/orderStatus.ts   # Colores de estado de orden
│   └── Dockerfile
└── docker-compose.yml              # PostgreSQL + backend + frontend
```

---

## Levantar el stack

### Opción 1 — Local (desarrollo)

**Requisitos:** Go 1.22+, Node 20+

```bash
# Backend
cd backend
APP_AUTH_JWT_SECRET=tu_secreto go run ./cmd/server/
# Corre en http://localhost:8080

# Frontend (otra terminal)
cd frontend
npm install
npm run dev
# Corre en http://localhost:5173
```

### Opción 2 — Docker Compose (producción)

**Requisitos:** Docker + Docker Compose

```bash
# Crear archivo de variables
cp backend/.env.example .env
# Editar .env con tus valores reales

docker-compose up --build
```

| Servicio | URL |
|---|---|
| Frontend | http://localhost:5173 |
| Backend | http://localhost:8080 |
| PostgreSQL | localhost:5432 |

### Demo de la API

```bash
cd backend
chmod +x examples/demo.sh
./examples/demo.sh
```

---

## Variables de entorno

| Variable | Default | Descripción |
|---|---|---|
| `APP_AUTH_JWT_SECRET` | *requerido* | Secreto para firmar JWT |
| `APP_DATABASE_DSN` | — | DSN de PostgreSQL (omitir para usar memoria) |
| `APP_SERVER_HTTP_PORT` | `8080` | Puerto del servidor HTTP |
| `APP_REPOSITORY_DRIVER` | `memory` | `memory` o `postgres` |

Para producción con PostgreSQL, copiar `.env.example` a `.env` y completar los valores.

---

## Tests

```bash
cd backend
go test ./...          # Todos los tests
go test -race ./...    # Con detector de race conditions
```

Cobertura actual: 17 tests en la capa de servicio cubriendo User, Product y Order.

---

## Repositorio base

Este proyecto usa [go-without-magic](https://github.com/JoX23/go-without-magic) como template del backend.
Si el template te resulta útil, dale una ⭐ al repositorio original.
