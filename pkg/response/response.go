package response

import "github.com/gofiber/fiber/v2"

// ErrorBody รูปแบบโครงสร้่าง Error มาตรฐาน
type ErrorBody struct {
	// example: BAD_REQUEST
	Code string `json:"code"`
	// example: invalid input data
	Message string `json:"message"`
}

// ใช้ตอบ error จาก middleware/handler
func Error(c *fiber.Ctx, status int, code, msg string) error {
	return c.Status(status).JSON(
		ErrorBody{Code: code, Message: msg},
	)
}

// JSON รูปแบบโครงสร้่างสำหรับการ คืนค่าแบบ Json
func JSON(c *fiber.Ctx, status int, v any) error {
	return c.Status(status).JSON(v)
}

// ใช้ตอบกลับสำเร็จ (ไม่มี Body)
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// ใช้ตอบกลับสำเร็จ (มี Data)
func OK(c *fiber.Ctx, data any) error {
	return JSON(c, fiber.StatusOK, data)
}

// ใช้ตอบกลับสร้างสำเร็จ
func Created(c *fiber.Ctx, data any) error {
	return JSON(c, fiber.StatusCreated, data)
}

// ใช้ตอบกลับแบบมีรายการ + หน้า
func Page(c *fiber.Ctx, data any, total int64, limit, offset int) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": fiber.Map{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}
