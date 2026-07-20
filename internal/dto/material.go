package dto

type CategoryNode struct {
	ID       uint64         `json:"id"`
	ParentID uint64         `json:"parentId"`
	Name     string         `json:"name"`
	Code     string         `json:"code"`
	Sort     int            `json:"sort"`
	Enabled  int            `json:"enabled"`
	Children []CategoryNode `json:"children,omitempty"`
}

type CategoryCreateReq struct {
	ParentID uint64 `json:"parentId"`
	Name     string `json:"name" binding:"required"`
	Sort     int    `json:"sort"`
	Enabled  *int   `json:"enabled"`
}

type CategoryUpdateReq struct {
	ParentID *uint64 `json:"parentId"`
	Name     *string `json:"name"`
	Sort     *int    `json:"sort"`
	Enabled  *int    `json:"enabled"`
}

type CategoryReorderReq struct {
	Items []CategoryReorderItem `json:"items" binding:"required"`
}

type CategoryReorderItem struct {
	ID       uint64 `json:"id" binding:"required"`
	ParentID uint64 `json:"parentId"`
	Sort     int    `json:"sort"`
}

type MaterialDTO struct {
	ID         uint64   `json:"id"`
	CategoryID uint64   `json:"categoryId"`
	Title      string   `json:"title"`
	MediaType  string   `json:"mediaType"`
	URL        string   `json:"url"`
	CoverURL   string   `json:"coverUrl"`
	FileName   string   `json:"fileName"`
	Mime       string   `json:"mime"`
	SizeBytes  int64    `json:"sizeBytes"`
	DurationMs int64    `json:"durationMs"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Tags       []string `json:"tags"`
	Remark     string   `json:"remark"`
	ProductID  *uint64  `json:"productId"`
	ProductSn  string   `json:"productSn"`
	Sort       int      `json:"sort"`
	Status     int      `json:"status"`
	CreatedBy  uint64   `json:"createdBy"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
}

type MaterialCreateReq struct {
	CategoryID uint64   `json:"categoryId"`
	Title      string   `json:"title" binding:"required"`
	MediaType  string   `json:"mediaType" binding:"required"`
	URL        string   `json:"url" binding:"required"`
	CoverURL   string   `json:"coverUrl"`
	FileName   string   `json:"fileName"`
	Mime       string   `json:"mime"`
	SizeBytes  int64    `json:"sizeBytes"`
	DurationMs int64    `json:"durationMs"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Tags       []string `json:"tags"`
	Remark     string   `json:"remark"`
	ProductID  *uint64  `json:"productId"`
	ProductSn  string   `json:"productSn"`
	Sort       int      `json:"sort"`
}

type MaterialUpdateReq struct {
	CategoryID *uint64  `json:"categoryId"`
	Title      *string  `json:"title"`
	CoverURL   *string  `json:"coverUrl"`
	Tags       []string `json:"tags"`
	Remark     *string  `json:"remark"`
	ProductID  *uint64  `json:"productId"`
	ClearProductID bool `json:"clearProductId"`
	ProductSn  *string  `json:"productSn"`
	Sort       *int     `json:"sort"`
	Status     *int     `json:"status"`
}

type MaterialListQuery struct {
	CategoryID  uint64 `form:"categoryId"`
	IncludeChildren bool `form:"includeChildren"`
	MediaType   string `form:"mediaType"`
	Keyword     string `form:"keyword"`
	ProductSn   string `form:"productSn"`
	HasProduct  string `form:"hasProduct"` // yes | no | ""
	Tag         string `form:"tag"`
	Page        int    `form:"page"`
	PageSize    int    `form:"pageSize"`
}

type BatchIDsReq struct {
	IDs []uint64 `json:"ids" binding:"required"`
}

type DashboardStats struct {
	MaterialCount    int64            `json:"materialCount"`
	ImageCount       int64            `json:"imageCount"`
	VideoCount       int64            `json:"videoCount"`
	WeekNewCount     int64            `json:"weekNewCount"`
	CategoryCount    int64            `json:"categoryCount"`
	ByCategory       []CategoryCount  `json:"byCategory"`
}

type CategoryCount struct {
	CategoryID   uint64 `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Count        int64  `json:"count"`
}
