package admin

import (
	"errors"
	"net/http"
	"strconv"

	"materialcore/internal/dto"
	"materialcore/internal/pkg/authcontext"
	"materialcore/internal/pkg/response"
	"materialcore/internal/service"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	svc *service.MaterialService
}

func NewHandlers(svc *service.MaterialService) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrBadRequest):
		response.Fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrConflict):
		response.Fail(c, http.StatusConflict, err.Error())
	default:
		response.Fail(c, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handlers) DashboardStats(c *gin.Context) {
	stats, err := h.svc.DashboardStats(authcontext.TenantID(c))
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, stats)
}

func (h *Handlers) ListCategories(c *gin.Context) {
	tree, err := h.svc.ListCategories(authcontext.TenantID(c))
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, tree)
}

func (h *Handlers) CreateCategory(c *gin.Context) {
	var req dto.CategoryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	node, err := h.svc.CreateCategory(authcontext.TenantID(c), req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Created(c, node)
}

func (h *Handlers) UpdateCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.CategoryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	node, err := h.svc.UpdateCategory(authcontext.TenantID(c), id, req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, node)
}

func (h *Handlers) DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteCategory(authcontext.TenantID(c), id); err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *Handlers) ReorderCategories(c *gin.Context) {
	var req dto.CategoryReorderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.ReorderCategories(authcontext.TenantID(c), req.Items); err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *Handlers) ListMaterials(c *gin.Context) {
	var q dto.MaterialListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.svc.ListMaterials(authcontext.TenantID(c), q)
	if err != nil {
		h.mapError(c, err)
		return
	}
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 24
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *Handlers) GetMaterial(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	m, err := h.svc.GetMaterial(authcontext.TenantID(c), id)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, m)
}

func (h *Handlers) CreateMaterial(c *gin.Context) {
	var req dto.MaterialCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	m, err := h.svc.CreateMaterial(authcontext.TenantID(c), authcontext.UserID(c), req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Created(c, m)
}

func (h *Handlers) UpdateMaterial(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.MaterialUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	m, err := h.svc.UpdateMaterial(authcontext.TenantID(c), id, req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, m)
}

func (h *Handlers) DeleteMaterial(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteMaterials(authcontext.TenantID(c), []uint64{id}); err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *Handlers) BatchDeleteMaterials(c *gin.Context) {
	var req dto.BatchIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.DeleteMaterials(authcontext.TenantID(c), req.IDs); err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}
