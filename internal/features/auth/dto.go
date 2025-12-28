package auth

// RegisterRequest ข้อมูลสำหรับการสมัครสมาขิก
type RegisterRequest struct {
	// example: json_doe
	Username string `json:"username" validate:"required,min=5,max=50"`
	// example: john@example.com
	Email string `json:"email" validate:"required,email,max=100"`
	// example: P@ssword123
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=5,max=50"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}


// LoginResponse ข้อมูล Token หลังจาก Login สำเร็จ
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGci01..."`
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
	Role     string
}

type LoginInput struct {
	Username string
	Password string
}
