package router

import (
	"ans-spareparts-api/internal/features/auth"
	"ans-spareparts-api/internal/features/category"
	"ans-spareparts-api/internal/features/inventory"
	"ans-spareparts-api/internal/features/product"
	"ans-spareparts-api/internal/features/user"
	"ans-spareparts-api/internal/infra/jwtx"
	"ans-spareparts-api/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

type Deps struct {
	AuthUC      auth.Service
	UserUC      user.Service
	ProductUC   product.Service
	CategoryUC  category.Service
	InventoryUC inventory.Service

	TokenManager jwtx.TokenManager
}

func RegisterRoutes(app *fiber.App, d Deps) {
	// Create handler
	authHandler := auth.NewHandler(d.AuthUC)
	userHandler := user.NewHandler(d.UserUC, d.AuthUC)
	productHandler := product.NewHandler(d.ProductUC)
	categoryHandler := category.NewHandler(d.CategoryUC)
	inventoryHandler := inventory.NewHandler(d.InventoryUC)

	// --- กำหนด Group /v1 ---
	api := app.Group("/v1")
	// --- RequireAuth path
	requireAuth := api.Group("/", middleware.RequireAuth(d.TokenManager))
	// --- RequireRole path
	requireRole := requireAuth.Group("/", middleware.RequireRole("manager"))

	// ---  Auth ไม่ต้องใช้ JWT ---
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	requireAuth.Post("/auth/logout", authHandler.Logout)

	// ---  User (ต้อง Login) ---
	users := requireAuth.Group("/users")
	users.Get("/:id", userHandler.GetProfile)
	users.Delete("/:id", userHandler.DeleteProfile)

	// --- Products (ต้อง Login) ---
	products := requireAuth.Group("/products")
	products.Get("/", productHandler.List)
	products.Get("/:id", productHandler.GetProductDetail)
	// --- Products (ต้อง Login และ เป็น Manager) ---
	productManager := requireRole.Group("/products")
	productManager.Post("/:id", productHandler.CreateProduct)
	productManager.Patch("/:id", productHandler.UpdateProduct)
	productManager.Delete("/:id", productHandler.DeleteProduct)
	// เรียก Inventory ด้วย ProductID
	products.Get("/:id/inventory", inventoryHandler.GetInventoryByProductID)

	// --- Category (ต้่อง Login )
	categories := requireAuth.Group("/categories")
	categories.Get("/", categoryHandler.List)
	categories.Get("/:id", categoryHandler.GetCategory)
	// --- Category (ต้่อง Login และ Role == "manager")
	categoriesManager := requireRole.Group("/categories")
	categoriesManager.Patch("/:id", categoryHandler.UpdateCategory)
	categoriesManager.Delete("/:id", categoryHandler.DeleteCategory)

	// --- Inventory (ต้อง Login) ---
	inventories := requireAuth.Group("/inventories")
	inventories.Get("/", inventoryHandler.List)
	inventories.Get("/:id", inventoryHandler.GetInventoryByID)
	inventories.Get("/:id", inventoryHandler.UpdateQuantity)

}
