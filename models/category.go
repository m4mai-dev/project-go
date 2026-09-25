package models

type Category struct {
	ID   uint   `gorm:"primaryKey;column:id_category" json:"id_category"`
	Name string `gorm:"column:category_name" json:"name"`
}

func (Category) TableName() string {
	return "category"
}