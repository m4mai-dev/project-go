package services

import (
	"belajar_go/config"
	"belajar_go/models"
	"errors"
)

func CreateProduct(product models.Product) (models.Product, error) {

	if product.ProductName == "" {
		return models.Product{}, errors.New("nama produk tidak boleh kosong")
	}

	query := `INSERT INTO products 
		(product_name, photo_url, deskripsi, spec, pricing_p_day, id_category) 
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := config.DB.Exec(
		query,
		product.ProductName,
		product.PhotoURL,
		product.Deskripsi,
		product.Spec,
		product.PricingPDay,
		product.IDCategory,
	)

	if err != nil {
		return models.Product{}, err
	}

	lastID, err := result.LastInsertId()

	if err != nil {
		return models.Product{}, err
	}

	product.ID = uint(lastID)

	return product, nil
}


func GetProducts() ([]models.Product, error) {

	rows, err := config.DB.Query(`
		SELECT
			id_product,
			product_name,
			photo_url,
			deskripsi,
			spec,
			pricing_p_day,
			id_category
		FROM products
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var products []models.Product

	for rows.Next() {

		var product models.Product

		err := rows.Scan(
			&product.ID,
			&product.ProductName,
			&product.PhotoURL,
			&product.Deskripsi,
			&product.Spec,
			&product.PricingPDay,
			&product.IDCategory,
		)

		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}