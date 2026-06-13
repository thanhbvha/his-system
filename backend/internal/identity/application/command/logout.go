package command

import (
	"context"
	"fmt"

	commonRedis "github.com/thanhbvha/go-common/redis"
	"his-system/pkg/auth"
)

type LogoutCommand struct {
	RefreshToken string
}

type LogoutHandler struct {
	rdb *commonRedis.Client
}

func NewLogoutHandler(rdb *commonRedis.Client) *LogoutHandler {
	return &LogoutHandler{rdb: rdb}
}

func (h *LogoutHandler) Handle(ctx context.Context, cmd LogoutCommand) error {
	rtHash := auth.HashToken(cmd.RefreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))

	// Delete regardless of existence (idempotent)
	h.rdb.Delete(ctx, rtKey)
	return nil
}
