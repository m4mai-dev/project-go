package models

type Product struct {
	ID          uint    `gorm:"primaryKey;column:id_product" json:"id"`
	ProductName string  `gorm:"column:product_name" json:"product_name"`
	PhotoURL    string  `gorm:"column:photo_url" json:"photo_url"`
	Deskripsi   string  `gorm:"column:deskripsi" json:"deskripsi"`
	Spec        string  `gorm:"column:spec" json:"spec"`
	PricingPDay float64 `gorm:"column:pricing_p_day" json:"pricing_p_day"`
	IDCategory  uint    `gorm:"column:id_category" json:"id_category"`
}

func (Product) TableName() string {
	return "products"
}