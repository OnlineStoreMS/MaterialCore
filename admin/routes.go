package admin

import "github.com/gin-gonic/gin"

func RegisterRoutes(g *gin.RouterGroup, h *Handlers, exportH *ExportHandler) {
	g.GET("/dashboard/stats", h.DashboardStats)

	g.GET("/categories", h.ListCategories)
	g.POST("/categories", h.CreateCategory)
	g.PUT("/categories/reorder", h.ReorderCategories)
	g.PUT("/categories/:id", h.UpdateCategory)
	g.DELETE("/categories/:id", h.DeleteCategory)

	g.GET("/materials", h.ListMaterials)
	g.POST("/materials", h.CreateMaterial)
	g.POST("/materials/batch-delete", h.BatchDeleteMaterials)
	g.POST("/materials/watermark", h.WatermarkMaterials)
	g.POST("/materials/export-zip", exportH.ExportZip)
	g.GET("/materials/:id", h.GetMaterial)
	g.PUT("/materials/:id", h.UpdateMaterial)
	g.POST("/materials/:id/rotate", h.RotateMaterial)
	g.DELETE("/materials/:id", h.DeleteMaterial)
}
