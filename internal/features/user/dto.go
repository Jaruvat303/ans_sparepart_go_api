package user

type UserUpdateRequest struct {
	Email    string
	Password string
}

type UserResponse struct {
	ID       uint   `json:"id" example:"10"`
	Username string `json:"username" example:"john"`
	Email    string `json:"email" example:"john@mail.com"`
	Role     string `json:"role" example:"admin"`
	IsActive bool   `json:"is_active" example:"true"`
}

type ListQuery struct {
	Search string
	Limit  int
	Offset int
	Sort   string
}

type UpdateInput struct {
	Email    *string
	Password *string
}

type Item struct {
	ID       uint
	Username string
	Email    string
	IsActive bool
	Role     string
}

// type ListOutput struct {
// 	Items []*Item
// 	Total int64
// }
