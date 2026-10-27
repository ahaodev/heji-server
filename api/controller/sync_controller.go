package controller

import (
	"net/http"
	"strconv"

	"shadmin/domain"
	"shadmin/internal/constants"

	"github.com/gin-gonic/gin"
)

// SyncController 增量同步接口。
type SyncController struct {
	SyncUseCase domain.SyncUseCase
}

// GetChanges 增量拉取变更
//
//	@Summary		Pull incremental changes
//	@Description	Return books and bills updated after `since` (milliseconds); deleted records carry deleted_at
//	@Tags			Sync
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			since	query		int												true	"Last sync cursor in milliseconds"
//	@Param			limit	query		int												false	"Max rows per entity (default 100, max 500)"
//	@Success		200		{object}	domain.Response{data=domain.SyncChangesResponse}	"Changes retrieved"
//	@Failure		400		{object}	domain.Response									"Missing or invalid since"
//	@Router			/sync/changes [get]
func (sc *SyncController) GetChanges(c *gin.Context) {
	since, err := strconv.ParseInt(c.Query("since"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.RespError("since is required and must be a millisecond timestamp"))
		return
	}

	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	result, err := sc.SyncUseCase.GetChanges(c.Request.Context(), c.GetString(constants.UserID), since, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(result))
}
