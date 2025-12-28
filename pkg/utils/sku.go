package utils

import (
	"ans-spareparts-api/pkg/apperror"
	"regexp"
	"strings"
)

// SKUValidationRegex อนุญาตเฉพาะตัวอักษร A-Z, 0-9, และเครื่องหมายยัติภังค์ (-) เท่านั้น
var SKUValidationRegex = regexp.MustCompile(`^[A-Z0-9-]+$`)

// ValidateAndNormalizeSKU ทำการตรวจสอบและปรับแต่ง SKU
func ValidateAndNormalizeSKU(rawSKU string) (string, error) {

	// Trimspace and Normalize
	sku := strings.TrimSpace(rawSKU)
	sku = strings.ToUpper(sku)

	if sku == "" {
		return "", apperror.ErrInvalidSKU
	}

	const (
		minLen = 3
		maxLen = 50
	)

	// length check
	if len(sku) < minLen || len(sku) > maxLen {
		return "", apperror.ErrInvalidSKU
	}

	// Regex validation
	if !SKUValidationRegex.MatchString(sku) {
		return "", apperror.ErrInvalidSKU
	}

	return sku, nil

}
