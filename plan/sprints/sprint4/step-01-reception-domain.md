# Sprint 4 — Step 1: Reception Domain + WebSocket Queue API (Backend)

## Mục tiêu
Xây dựng nền tảng Backend cho module `internal/reception`: domain entity, và toàn bộ Queue API (check-in, gọi số, bỏ qua, hoàn thành).
Đặc biệt, tích hợp thư viện `github.com/thanhbvha/go-common/websocket` để quản lý WebSocket realtime hiệu năng cao.

## Phân tích thư viện `go-common/websocket`
Thư viện đã được tải về trong dự án cung cấp sẵn cấu trúc rất mạnh mẽ:
1. **Core Lifecycle**: Có sẵn `core.Manager` (quản lý connection pool qua Shards), `core.Connection` (xử lý read/write pump, ping/pong tự động). Không cần tự viết lại `hub.go` và `client.go`.
2. **PubSub**: Tích hợp Redis PubSub để scale đa node (nếu chạy nhiều instance API).
3. **Adapter Fiber**: Thư viện cung cấp sẵn `adapter.fiber.Handler`. Tuy nhiên, `Handler` mặc định chặn (block) ngay sau khi upgrade để chạy `readPump()`, **không có hook `OnConnect`** để chúng ta gửi `current queue state` lúc vừa kết nối.
4. **Quyết định**: Chúng ta sẽ **sử dụng `core`** của thư viện, nhưng sẽ **viết mới một Custom Fiber Adapter** (`pkg/ws/adapter.go`) dựa trên adapter gốc, có bổ sung thêm callback `OnConnect` để đáp ứng đúng yêu cầu của hệ thống (gửi trạng thái queue ngay khi connect).

## Files cần tạo / cập nhật
```
backend/
├── pkg/ws/
│   ├── adapter.go       -- Custom Fiber Adapter (bọc core manager + hỗ trợ OnConnect)
│   └── event.go         -- Khai báo cấu trúc WSEvent (JSON payloads)
├── internal/reception/
│   ├── domain/
│   │   ├── queue_entry.go    -- Entity QueueEntry + status machine
│   │   └── repository.go     -- QueueRepository interface
│   ├── application/
│   │   ├── commands/
│   │   │   ├── checkin.go
│   │   │   ├── call_queue.go
│   │   │   ├── skip_queue.go
│   │   │   └── complete_queue.go
│   │   └── queries/
│   │       ├── get_current_queue.go
│   │       └── get_queue_stats.go
│   ├── infrastructure/
│   │   └── queue_repository_pg.go
│   ├── handlers/
│   │   └── queue_handler.go   -- HTTP + config Custom WS Adapter
│   └── bootstrap/
│       ├── module.go
│       └── router.go
```

## Nhiệm vụ chi tiết

### 1. Custom WebSocket Adapter (`pkg/ws/`)

**`pkg/ws/adapter.go`:**
Dựa trên `fiber/handler.go` của thư viện, ta viết một Handler mới cho phép truyền vào `OnConnect func(userID string, sendJSON func(interface{}))`:
```go
package ws

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	libadapter "github.com/thanhbvha/go-common/websocket/adapter/fiber"
	"github.com/thanhbvha/go-common/websocket/core"
)

type CustomWSHandler struct {
	Authenticate func(c *fiber.Ctx) (string, error)
	OnConnect    func(userID string, sendJSON func(interface{}))
}

func (h *CustomWSHandler) HandleUpgrade(c *fiber.Ctx) error {
    // 1. Authenticate lấy userID từ query ?token=
    // 2. websocket.New upgrade
    // 3. Trong block websocket.New:
    //    - Tạo core.Conn (qua libadapter.NewConnAdapter)
    //    - Tạo shard, tạo core.Connection
    //    - Gọi h.OnConnect(userID, connection.SendJSON) ĐỂ GỬI QUEUE STATE BAN ĐẦU
    //    - Chạy go connection.WritePump()
    //    - Block với connection.ReadPump()
}
```

**`pkg/ws/event.go`:**
```go
package ws

// WSEvent defines the payload format broadcasted to clients
type WSEvent struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}

// Event types
const (
    EventQueueUpdated  = "queue.updated"
    EventQueueCalled   = "queue.called"
    EventQueueCompleted = "queue.completed"
)

// Helper broadcast
func BroadcastToAll(eventType string, payload interface{}) {
    manager := core.GetGlobalManager()
    event := WSEvent{Type: eventType, Payload: payload}
    // Convert to JSON byte array and use manager.BroadcastToAll()
}
```

### 2. Domain Layer (`internal/reception/domain/`)

**`queue_entry.go`:**
```go
type QueueStatus string
const (
    StatusWaiting    QueueStatus = "WAITING"
    StatusCalled     QueueStatus = "CALLED"
    StatusInProgress QueueStatus = "IN_PROGRESS"
    StatusDone       QueueStatus = "DONE"
    StatusSkipped    QueueStatus = "SKIPPED"
)

type QueueEntry struct {
    ID            uuid.UUID
    PatientID     uuid.UUID
    VisitID       *uuid.UUID
    AppointmentID *uuid.UUID
    ServiceType   string       // "GENERAL", "LAB", "RADIOLOGY", ...
    QueueNumber   string       // "KB001", "XN001"
    Status        QueueStatus
    CalledAt      *time.Time
    CompletedAt   *time.Time
    CreatedAt     time.Time
}

func (q *QueueEntry) Call() error
func (q *QueueEntry) Skip() error
func (q *QueueEntry) Complete() error
```

**`repository.go`:**
```go
type QueueRepository interface {
    Save(ctx, *QueueEntry) error
    FindByID(ctx, uuid.UUID) (*QueueEntry, error)
    FindTodayQueue(ctx, serviceType string) ([]*QueueEntry, error)
    GetNextSequence(ctx, prefix string) (int, error)
    UpdateStatus(ctx, id uuid.UUID, status QueueStatus) error
    GetStats(ctx) (*QueueStats, error)
}
```

### 3. Application Layer

**Commands:**
- `CheckInCommand` → tạo `QueueEntry` → publish Redis stream `HIS.VISIT.QueueCheckedIn` → dùng `ws.BroadcastToAll(ws.EventQueueUpdated, queueData)`
- `CallQueueCommand` → update status `CALLED` → `ws.BroadcastToAll(ws.EventQueueCalled, queueData)`
- `SkipQueueCommand` → update status `SKIPPED` → `ws.BroadcastToAll(ws.EventQueueUpdated, queueData)`
- `CompleteQueueCommand` → update status `DONE` → `ws.BroadcastToAll(ws.EventQueueCompleted, queueData)`

**Queries:**
- `GetCurrentQueue{ServiceType?}` → trả tất cả queue entries trong ngày còn WAITING/CALLED
- `GetQueueStats{}` → trả `{waiting_count, called_count, avg_wait_minutes}`

### 4. Infrastructure Layer

**`queue_repository_pg.go`:**
- Table: `queue_entries`
- Index trên `(created_at::date, service_type, status)`

### 5. Handlers (`internal/reception/handlers/queue_handler.go`)

Khởi tạo Custom WebSocket Handler:
```go
wsHandler := &ws.CustomWSHandler{
    Authenticate: func(c *fiber.Ctx) (string, error) {
        // Parse JWT token từ query ?token=
        // Return userID
    },
    OnConnect: func(userID string, sendJSON func(interface{})) {
        // Query GetCurrentQueue
        // queueData := ...
        // sendJSON(ws.WSEvent{Type: ws.EventQueueUpdated, Payload: queueData})
    },
}

app.Get("/api/v1/queue/ws", wsHandler.HandleUpgrade)
```

Các APIs HTTP bình thường:
```
GET  /api/v1/queue              → GetCurrentQueue
POST /api/v1/queue/checkin      → CheckInCommand
POST /api/v1/queue/call/:id     → CallQueueCommand
POST /api/v1/queue/skip/:id     → SkipQueueCommand
POST /api/v1/queue/complete/:id → CompleteQueueCommand
GET  /api/v1/queue/stats        → GetQueueStats
```

### 6. Module Registration
- Đăng ký module trong `cmd/api/main.go`. Khởi tạo DB repo và truyền vào handler.

## Database Migration
```sql
CREATE TABLE queue_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patients(id),
    visit_id        UUID REFERENCES visits(id),
    appointment_id  UUID REFERENCES appointments(id),
    service_type    VARCHAR(50) NOT NULL DEFAULT 'GENERAL',
    queue_number    VARCHAR(10) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'WAITING',
    called_at       TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_queue_today ON queue_entries((created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date, service_type, status);
```

## Kiểm tra hoàn thành
- [ ] `go build ./...` thành công, không có lỗi compile
- [ ] Custom WS Adapter hoạt động: Client kết nối WS thành công qua `?token=` và nhận được danh sách queue hiện tại ngay lập tức.
- [ ] HTTP `POST /queue/checkin` tạo queue_number thành công.
- [ ] Mọi thay đổi qua HTTP đều tự động gọi `ws.BroadcastToAll()`, dữ liệu được publish tới mọi client qua `go-common/websocket`.
- [ ] Heartbeat ping/pong (60s mặc định của thư viện) tự động chạy ổn định.
