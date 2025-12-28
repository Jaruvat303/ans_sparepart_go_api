package category

type ListQuery struct {
	Search string
	Limit  int
	Offset int
	Sort   string
}

type Item struct {
	ID   uint
	Name string
}
type ListOutput struct {
	Items []*Item
	Total int64
}

type CategoryRequest struct {
	Name string `json:"name"`
}

type CategoryResponse struct {
	ID   uint
	Name string
}

type CategoryListResponse struct {
	Categories []*CategoryResponse
	Total      int64
}
