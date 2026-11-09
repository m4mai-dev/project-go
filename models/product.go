package models

type Product struct {
	ID             uint    `gorm:"primaryKey;column:id_product" json:"id"`
	ProductID      uint    `gorm:"-" json:"id_product,omitempty"`
	ProductName    string  `gorm:"column:product_name" json:"product_name"`
	PhotoURL       string  `gorm:"column:photo_url" json:"photo_url"`
	Deskripsi      string  `gorm:"column:deskripsi" json:"deskripsi"`
	Spec           string  `gorm:"column:spec" json:"spec"`
	PricingPDay    float64 `gorm:"column:pricing_p_day" json:"pricing_p_day"`
	IDCategory     uint    `gorm:"column:id_category" json:"id_category"`
	CategoryName   string  `gorm:"-" json:"category_name,omitempty"`
	AvailableUnits int     `gorm:"-" json:"available_units"`
	TotalUnits     int     `gorm:"-" json:"total_units"`
}

type ProductDetailResponse struct {
	Product
	Units []ItemInstance `json:"units,omitempty"`
}

type ProductInput struct {
	ProductName string  `json:"product_name"`
	PhotoURL    string  `json:"photo_url"`
	Deskripsi   string  `json:"deskripsi"`
	Spec        string  `json:"spec"`
	PricingPDay float64 `json:"pricing_p_day"`
	IDCategory  uint    `json:"id_category"`
}

func (Product) TableName() string {
	return "products"
}
