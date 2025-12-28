package inventory

type ListQuery struct {
	Limit  int
	Offset int
	Sort   string
}

type UpdateQuantityInput struct {
	ProductID uint
	Quantity  int
}

type Item struct {
	ID        uint
	ProductID uint
	Quantity  int
}

type ListOutput struct {
	Items []*Item
	Total int64
}

type UpdateQuantityRequest struct {
	ProductID uint
	Quantity  int
}

type InventoryResponse struct {
	ID        uint
	ProductID uint
	Quantity  int
}

type InventoryListResponse struct {
	Inventories []*InventoryResponse
	Total       int64
}
