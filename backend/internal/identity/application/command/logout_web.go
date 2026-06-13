package command

import (
	"context"
	"fmt"

	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/pkg/auth"
)

type LogoutWebCommand struct {
	RefreshToken string
}

type LogoutWebHandler struct {
	rdb *commonRedis.Client
}

func NewLogoutWebHandler(rdb *commonRedis.Client) *LogoutWebHandler {
	return &LogoutWebHandler{rdb: rdb}
}

func (h *LogoutWebHandler) Handle(ctx context.Context, cmd LogoutWebCommand) error {
	if cmd.RefreshToken == "" {
		return nil // Nothing to do
	}
	rtHash := auth.HashToken(cmd.RefreshToken)
	rtKey := h.rdb.BuildKey(fmt.Sprintf("refresh:%s", rtHash))

	h.rdb.Delete(ctx, rtKey)
	return nil
}
