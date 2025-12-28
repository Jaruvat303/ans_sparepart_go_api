package fixtures

import "ans-spareparts-api/internal/domain"

// --- Fixtures ---
func ValidUser() *domain.User {
	return &domain.User{
		ID: 1, Username: "testUser", Email: "testuser@mail.com",
		Password: "P@ssword1234", Role: "cashier", IsActive: true,
	}
}

func ValidCategory() *domain.Category {
	return &domain.Category{
		ID: 1, Name: "Wheel",
	}
}

func ValidListCategory() []*domain.Category {
	return []*domain.Category{
		{ID: 1, Name: "Wheel"},
		{ID: 2, Name: "Oil"},
		{ID: 3, Name: "Oring"},
	}
}

func ValidInventory() *domain.Inventory {
	return &domain.Inventory{
		ID:        1,
		ProductID: 1,
		Quantity:  1,
	}
}

func ValidListInventory() []*domain.Inventory {
	return []*domain.Inventory{
		{ID: 1, ProductID: 1, Quantity: 11},
		{ID: 2, ProductID: 2, Quantity: 20},
		{ID: 3, ProductID: 3, Quantity: 31},
	}
}

func ValidProductLite() *domain.Product {
	return &domain.Product{
		ID:          1,
		Name:        "Test",
		Description: "Description",
		Price:       1,
		CategoryID:  1,
		IsActive:    true,
	}
}

func ValidProduct() *domain.Product {
	return &domain.Product{
		ID:          1,
		Name:        "Test",
		Description: "Desciption",
		SKU:         "TestSku",
		Price:       1,
		CategoryID:  1,
		Category: domain.Category{
			ID:   1,
			Name: "Category Test",
		},
		Inventory: domain.Inventory{
			ID:        1,
			ProductID: 1,
			Quantity:  1,
		},
		IsActive: true,
	}
}

func ValidListProduct() []*domain.Product {
	return []*domain.Product{
		{ID: 1, Name: "product1", Description: "desc1", Price: 1, CategoryID: 1, IsActive: true},
		{ID: 2, Name: "product2", Description: "desc2", Price: 2, CategoryID: 2, IsActive: true},
	}
}
