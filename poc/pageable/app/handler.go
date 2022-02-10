package app

import (
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"strings"
)

type Handler struct {
	userStore *UserStore
}

func NewHandler(userStore *UserStore) *Handler {
	return &Handler{
		userStore: userStore,
	}
}

func (h *Handler) HandleGetUsers(ctx *gin.Context) {
	type Opt struct {
		LastName string   `form:"lastName" binding:"required"`
		Sorts    []string `form:"sort"`
		Page     int      `form:"page,default=1"`
		Size     int      `form:"size,default=3"`
	}
	opt := Opt{}
	if err := ctx.ShouldBind(&opt); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	sorts, err := ParseSorts(opt.Sorts, func(property string) string {
		switch p := strings.ToLower(property); p {
		case "email", "name":
			return p
		default:
			return ""
		}
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	pageable := Pageable{
		PageNumber: opt.Page,
		PageSize:   opt.Size,
		Sorts:      sorts,
	}

	users, total, err := h.userStore.FindAllByLastName(ctx, opt.LastName, pageable)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	totalPages := calculateTotalPages(total, int64(opt.Size))
	ctx.JSON(http.StatusOK, gin.H{
		"data":          users,
		"totalPages":    totalPages,
		"totalElements": total,
	})
}

func calculateTotalPages(total, pageSize int64) int64 {
	if pageSize == 0 {
		return 1
	}
	return int64(math.Ceil(float64(total) / float64(pageSize)))
}
