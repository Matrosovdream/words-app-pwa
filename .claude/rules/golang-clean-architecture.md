# Go Clean Architecture — Project Rules

These rules are derived from https://github.com/asseph/golang-clean-architecture and are to be followed for **all Go code in this repo**. When writing, reviewing, or refactoring Go, apply these conventions strictly.

---

## 1. Layered architecture (dependency inversion)

Request flow moves through clearly separated layers. Each layer only depends on layers closer to the domain (inward).

```
External system
     │
     ▼
Delivery  ──►  UseCase  ──►  Repository  ──►  Database
                  │
                  └──►  Gateway  ──►  External system
```

**Rule:** never let an outer layer be imported by an inner layer.
- `usecase` may import `repository`, `entity`, `model`, `gateway`.
- `repository` may import `entity` only.
- `entity` imports nothing from the app (pure domain).
- `delivery` may import `usecase`, `model`.

## 2. Folder layout

```
/api                    # API specs (OpenAPI, proto files)
/cmd                    # Main entry points (one main.go per binary)
  /web                  # HTTP server
  /worker               # Background workers
/db/migrations          # SQL migrations
/internal               # Private app code (everything below)
  /config               # Viper config loading, app bootstrap
  /delivery             # HTTP handlers, gRPC, messaging consumers
    /http               # Fiber controllers, routes, middleware
    /messaging          # Kafka consumers
  /entity               # Domain structs (map to DB tables)
  /model                # Request/response DTOs + converters
    /converter          # entity ↔ model mapping functions
  /repository           # DB access (one repo per entity)
  /usecase              # Business logic (one usecase per entity)
  /gateway              # Outbound calls to external systems
    /messaging          # Kafka producers
/test                   # Unit/integration tests
config.json             # Runtime config
```

## 3. Naming conventions

| Layer | File name | Struct name | Constructor |
|---|---|---|---|
| UseCase | `user_usecase.go` | `UserUseCase` | `NewUserUseCase(...)` |
| Repository | `user_repository.go` | `UserRepository` | `NewUserRepository(...)` |
| Controller | `user_controller.go` | `UserController` | `NewUserController(...)` |
| Entity | `user.go` | `User` | — |
| Model (request) | `user_model.go` | `RegisterUserRequest`, `UpdateUserRequest` | — |
| Model (response) | `user_model.go` | `UserResponse`, `Auth` | — |
| Converter | `user_converter.go` | functions only: `UserToResponse`, `UserToEvent` | — |
| Producer | `user_producer.go` | `UserProducer` | `NewUserProducer(...)` |

- Files use `snake_case`. Types use `PascalCase`. All exported identifiers start uppercase.
- One entity per file group. Do **not** put multiple entities in the same file.

## 4. Struct & constructor pattern

Every service-like type embeds its dependencies and is created via a `New...` constructor.

```go
type UserUseCase struct {
    DB             *gorm.DB
    Log            *logrus.Logger
    Validate       *validator.Validate
    UserRepository *repository.UserRepository
    UserProducer   *messaging.UserProducer
}

func NewUserUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
    userRepository *repository.UserRepository, userProducer *messaging.UserProducer) *UserUseCase {
    return &UserUseCase{
        DB:             db,
        Log:            logger,
        Validate:       validate,
        UserRepository: userRepository,
        UserProducer:   userProducer,
    }
}
```

- Struct fields are exported (`DB`, `Log`, `Validate`) — used across layers.
- Constructor returns a pointer.
- Receiver convention: `(c *UserUseCase)` / `(c *UserController)` — always use `c`.

## 5. UseCase method contract

Every usecase method follows this shape:

```go
func (c *UserUseCase) Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body : %+v", err)
        return nil, fiber.ErrBadRequest
    }

    // ... business logic using entity + repository ...

    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed commit transaction : %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    return converter.UserToResponse(user), nil
}
```

**Rules:**
1. First parameter is always `ctx context.Context`.
2. Second parameter is a `*model.XxxRequest` pointer.
3. Begin a transaction at the top; `defer tx.Rollback()` immediately.
4. Validate the request **before** any DB work — return `fiber.ErrBadRequest` on failure.
5. Log every failure with `c.Log.Warnf("Failed to ... : %+v", err)` before returning.
6. Return a `*model.XxxResponse` via a converter function — never return entities directly.
7. Commit the transaction explicitly before returning success.
8. Map domain errors to `fiber.Err*` constants (`ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrInternalServerError`).

## 6. Delivery (HTTP controller) contract

```go
type UserController struct {
    Log     *logrus.Logger
    UseCase *usecase.UserUseCase
}

func NewUserController(useCase *usecase.UserUseCase, logger *logrus.Logger) *UserController {
    return &UserController{Log: logger, UseCase: useCase}
}

func (c *UserController) Register(ctx *fiber.Ctx) error {
    request := new(model.RegisterUserRequest)
    if err := ctx.BodyParser(request); err != nil {
        c.Log.Warnf("Failed to parse request body : %+v", err)
        return fiber.ErrBadRequest
    }
    response, err := c.UseCase.Create(ctx.UserContext(), request)
    if err != nil {
        return err  // usecase already returned a fiber.Err*
    }
    return ctx.JSON(model.WebResponse[*model.UserResponse]{Data: response})
}
```

**Rules:**
- Controllers are thin — only parse request, call usecase, wrap response.
- Never put business logic in a controller.
- Always wrap responses in the generic `model.WebResponse[T]{Data: ...}`.
- For protected endpoints, get the current user via `middleware.GetUser(ctx)`.

## 7. Repository contract

Repositories accept a `*gorm.DB` (or `*gorm.DB` transaction) as first param — they don't own the transaction.

```go
type UserRepository struct {
    Log *logrus.Logger
}

func NewUserRepository(log *logrus.Logger) *UserRepository {
    return &UserRepository{Log: log}
}

func (r *UserRepository) FindById(db *gorm.DB, user *entity.User, id string) error {
    return db.Where("id = ?", id).First(user).Error
}

func (r *UserRepository) Create(db *gorm.DB, user *entity.User) error {
    return db.Create(user).Error
}
```

- First argument is always the `db` handle (from usecase's `tx`).
- Returns only `error` (or `(int64, error)` for counts).
- No business logic, no validation — just DB operations.

## 8. Model & converter pattern

- **Request models** carry `validate:"..."` tags for go-playground validator.
- **Response models** are the outward shape (no sensitive fields).
- **Converters** are pure functions in `model/converter/`:

```go
func UserToResponse(user *entity.User) *model.UserResponse { ... }
func UserToTokenResponse(user *entity.User) *model.UserResponse { ... }
func UserToEvent(user *entity.User) *model.UserEvent { ... }
```

Never expose an `entity.*` type via HTTP or Kafka. Always convert.

## 9. Library stack (use these, not alternatives)

| Concern | Library |
|---|---|
| HTTP | `github.com/gofiber/fiber/v2` |
| ORM | `gorm.io/gorm` |
| Config | `github.com/spf13/viper` |
| Validation | `github.com/go-playground/validator/v10` |
| Logging | `github.com/sirupsen/logrus` |
| Migrations | `github.com/golang-migrate/migrate` |
| Kafka | `github.com/IBM/sarama` |
| UUID | `github.com/google/uuid` |
| Password hashing | `golang.org/x/crypto/bcrypt` |

Default to these. If a different library is actually needed, flag it explicitly and justify.

## 10. Logging discipline

- Inject `*logrus.Logger` as a field, never use the package-level logger.
- Log **warnings on every failure path** before returning: `c.Log.Warnf("Failed to X : %+v", err)`.
- Log info when publishing events: `c.Log.Info("Publishing user created event")`.
- Use `%+v` for errors to get stack-friendly output.

## 11. Transactions

- Usecases own the transaction lifecycle (`Begin`, `Rollback` via defer, `Commit`).
- Repositories **receive** the transaction `*gorm.DB`, never start their own.
- Always `defer tx.Rollback()` right after `tx := db.Begin()` — it's a no-op after commit.

## 12. Events / messaging

- Kafka producers are nullable — always guard with `if c.UserProducer != nil`.
- If disabled, log: `c.Log.Info("Kafka producer is disabled, skipping ...")`.
- Convert entity to event via `converter.UserToEvent(user)` before sending.
- Publishing failure returns `fiber.ErrInternalServerError` from the usecase.

## 13. Configuration

- All runtime config in `config.json` at project root.
- Loaded via Viper in `internal/config`.
- Never hard-code URLs, credentials, or feature flags — read from config.

## 14. Commands

```bash
go test -v ./test/          # Run tests
go run cmd/web/main.go      # Start HTTP server
go run cmd/worker/main.go   # Start Kafka worker
```

---

## Application when modifying this repo

**Note:** the current `backend/` directory in this project does **not** yet follow this architecture — it's a flat `main.go` serving a small JSON API. When the user asks for expansion (new entities, DB, auth, events), **migrate toward this layout incrementally**:

1. Move routes + handlers into `internal/delivery/http/`.
2. Extract business logic into `internal/usecase/`.
3. Introduce `internal/entity/`, `internal/model/`, `internal/repository/` as data/persistence appears.
4. Move `main.go` to `cmd/web/main.go`.

Do not do this migration until the user asks for it. When they do, follow the rules above verbatim.
