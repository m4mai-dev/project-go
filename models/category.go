package models

type Category struct {
	ID   uint   `gorm:"primaryKey;column:id_category" json:"id_category"`
	Name string `gorm:"column:category_name" json:"category_name"`
}

type CategoryInput struct {
	Name         string `json:"name"`
	CategoryName string `json:"category_name"`
}

func (c CategoryInput) Value() string {
	if c.CategoryName != "" {
		return c.CategoryName
	}
	return c.Name
}

func (Category) TableName() string {
	return "category"
}
