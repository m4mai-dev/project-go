package services

import (
	"belajar_go/config"
	"belajar_go/models"
)

func CreateItemInstance(input models.CreateItemInstanceInput) (models.ItemInstance, error) {
	if input.Status == "" {
		input.Status = "available"
	}

	productID := input.GetProductID()
	query := `INSERT INTO item_instances (id_products, asset_code, status) VALUES (?, ?, ?)`

	result, err := config.DB.Exec(query, productID, input.AssetCode, input.Status)
	if err != nil {
		return models.ItemInstance{}, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return models.ItemInstance{}, err
	}

	newItemInstance := models.ItemInstance{
		ID:         uint(lastID),
		IDUnit:     uint(lastID),
		IDProducts: productID,
		AssetCode:  input.AssetCode,
		Status:     input.Status,
	}

	return newItemInstance, nil
}
