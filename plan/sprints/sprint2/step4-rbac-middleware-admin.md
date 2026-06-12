# Sprint 2 — Step 4: RBAC, Signature Middleware & Admin APIs

> **Mục tiêu:** Bảo vệ toàn bộ API bằng RBAC middleware; thêm Request Signature Middleware chống request giả mạo cho Desktop; xây dựng các API quản trị User/Role/Permission.
> **Phụ thuộc:** Step 2 (JWT issue/verify hoạt động), Step 1 (Role entity, RoleRepository).
> **Output:** Mọi route protected đều check permission; request từ Desktop không có signature hợp lệ bị block.

---

## Nền tảng Sprint 1 + bước trước sử dụng

| Package | Dùng để |
|---------|---------|
| `pkg/auth/jwt.go` — `VerifyAccessToken` | Extract claims từ JWT trong middleware |
| `internal/identity/domain/Role, Permission` | Kiểm tra permission trong claims |
| `internal/identity/infrastructure/RoleRepositoryPG` | `UpdatePermissions` |
| `pkg/errors/errors.go` | `ErrUnauthorized (401)`, `ErrForbidden (403)` |
| `pkg/middleware/recover.go` | Pattern middleware để viết RBAC tương tự |
| `go-common/redis` | Cache permission list (TTL 5m để giảm DB hit) |

---

## 1. JWT Auth Middleware — `pkg/middleware/jwt_auth.go`

- [ ] `pkg/middleware/jwt_auth.go`:
  ```go
  // JWTAuth extract và verify JWT từ Authorization header.
  // Attach Claims vào fiber.Ctx locals để các handler downstream dùng.
  func JWTAuth(signingKey, encKey []byte) fiber.Handler
  ```
  - Extract `Authorization: Bearer <token>`
  - Gọi `auth.VerifyAccessToken(token, signingKey, encKey)`
  - Nếu lỗi → trả `ErrUnauthorized`
  - Lưu claims vào context: `c.Locals("claims", claims)`
  - Helper: `GetClaims(c *fiber.Ctx) (auth.Claims, bool)`

---

## 2. RBAC Middleware — `pkg/middleware/rbac.go`

- [ ] `pkg/middleware/rbac.go`:
  ```go
  // RequirePermission trả về middleware kiểm tra claims có chứa permission yêu cầu không.
  // Dùng cùng với JWTAuth (phải gọi JWTAuth trước trong chain).
  func RequirePermission(permission string) fiber.Handler

  // RequireRole kiểm tra ít nhất 1 trong roles được phép.
  func RequireRole(roles ...string) fiber.Handler
  ```
  - Lấy `claims` từ `c.Locals("claims")`
  - Check `claims.Permissions` chứa permission yêu cầu
  - Nếu thiếu → `ErrForbidden (403)`

- [ ] Cách dùng trên route:
  ```go
  users := api.Group("/api/v1/users", jwtAuth, rbac.RequireRole("admin"))
  users.Get("/",    userHandler.List)
  users.Post("/",   userHandler.Create)
  users.Put("/:id/roles", rbac.RequirePermission("user:assign_role"), userHandler.AssignRoles)
  ```

---

## 3. Request Signature Middleware — `pkg/middleware/signature.go`

> Áp dụng cho các route nhạy cảm của Desktop (không áp dụng cho Web).

- [ ] `pkg/middleware/signature.go`:
  ```go
  // RequestSignature verify chữ ký của request từ Desktop client.
  // Dùng sau JWTAuth (cần claims.cnf.jkt để lấy public key hash).
  func RequestSignature(deviceRepo DeviceRepository) fiber.Handler
  ```
  - Đọc Headers: `X-Timestamp`, `X-Signature`
  - Check `X-Timestamp`: nếu chênh lệch > 5 phút với server time → reject (chống replay)
  - Lấy `cnf.jkt` (public key hash) từ JWT claims
  - Lookup Device trong DB bằng `cnf.jkt` → lấy `PublicKeyPEM`
  - Tạo message: `SHA256(Method + URL + X-Timestamp + Body)`
  - Verify ECDSA-P256 signature bằng PublicKeyPEM (`crypto/ecdsa`, `elliptic.P256()`)
  - Nếu fail → `401 INVALID_SIGNATURE`

- [ ] Unit test:
  - Valid signature → pass
  - Expired timestamp (> 5 phút) → 401
  - Tampered body → 401
  - Unknown public key hash → 401

---

## 4. Queries — `internal/identity/application/query/`

- [ ] `get_user_by_id.go`:
  ```go
  type GetUserByIDQuery struct { ID uuid.UUID }
  type GetUserByIDResult struct { User *User }
  ```

- [ ] `list_users.go`:
  ```go
  type ListUsersQuery struct {
      Page   int
      Limit  int
      Role   string   // optional filter
      Search string   // optional: search by username
  }
  type ListUsersResult struct {
      Users []*User
      Total int64
  }
  ```

- [ ] `get_role_permissions.go`:
  ```go
  type GetRolePermissionsQuery struct { RoleID uuid.UUID }
  type GetRolePermissionsResult struct { Role *Role }
  ```

---

## 5. Admin API Endpoints — `internal/api/admin/`

### Route Setup

- [ ] Đăng ký admin routes (tất cả require `JWTAuth` + `RequireRole("admin")`):
  ```go
  adminUsers := api.Group("/api/v1/users", jwtAuth, rbac.RequireRole("admin"))
  adminUsers.Get("/",                userHandler.List)
  adminUsers.Post("/",               userHandler.Create)
  adminUsers.Get("/:id",             userHandler.GetByID)
  adminUsers.Put("/:id",             userHandler.Update)
  adminUsers.Put("/:id/deactivate",  userHandler.Deactivate)
  adminUsers.Put("/:id/roles",       userHandler.AssignRoles)

  adminRoles := api.Group("/api/v1/roles", jwtAuth, rbac.RequireRole("admin"))
  adminRoles.Get("/",                      roleHandler.List)
  adminRoles.Get("/:id/permissions",       roleHandler.GetPermissions)
  adminRoles.Put("/:id/permissions",       roleHandler.UpdatePermissions)

  adminDepts := api.Group("/api/v1/departments", jwtAuth, rbac.RequireRole("admin"))
  adminDepts.Get("/",   deptHandler.List)
  adminDepts.Post("/",  deptHandler.Create)
  ```

### Handler Implementations

- [ ] `GET /api/v1/users` — paginated, filter by role, search by username
- [ ] `POST /api/v1/users` — tạo staff account (không phải patient):
  ```
  Request: { "username": "drnguyen", "email": "dr@hospital.vn", "role_ids": ["uuid"], "department_id": "uuid" }
  Response: { "id": "uuid", "username": "drnguyen", ... }
  ```
  - Generate temporary password (12 chars random)
  - Enqueue job gửi email thông tin đăng nhập
- [ ] `PUT /api/v1/users/:id/deactivate` — set `is_active = false`, xoá refresh tokens khỏi Redis
- [ ] `PUT /api/v1/users/:id/roles` — update `user_roles` table
- [ ] `GET /api/v1/roles` — list roles với permissions
- [ ] `PUT /api/v1/roles/:id/permissions` — bulk update permissions (delete + insert batch)
- [ ] `GET /api/v1/departments` — list departments
- [ ] `POST /api/v1/departments` — tạo department mới

---

## 6. Permission Cache (Redis)

- [ ] Cache permissions của role trong Redis (TTL 5 phút) để giảm DB query:
  ```go
  // Key: "perm:{role_name}" → JSON([]string{"patient:read", "appointment:write"})
  // Khi UpdatePermissions: xoá key cache
  ```

---

## Definition of Done (Step 4)

- [ ] `GET /api/v1/users` không có JWT → 401
- [ ] `GET /api/v1/users` với JWT role=doctor → 403 (chỉ admin được phép)
- [ ] `GET /api/v1/users` với JWT role=admin → 200 + danh sách users
- [ ] `POST /api/v1/users` tạo staff account thành công, email notification enqueue
- [ ] `PUT /api/v1/users/:id/deactivate` → user không login được nữa
- [ ] `PUT /api/v1/roles/:id/permissions` → permissions cập nhật trong DB và cache bị xoá
- [ ] Signature Middleware: request Desktop không có `X-Signature` → 401
- [ ] Signature Middleware: timestamp cũ > 5 phút → 401
- [ ] Signature Middleware: body bị tamper → 401
- [ ] Unit test signature middleware: 4 case trên đều pass
- [ ] Swagger docs hiển thị tất cả admin endpoints
