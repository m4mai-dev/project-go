package models

type ItemInstance struct {
	ID          uint   `json:"id"`
	IDUnit      uint   `json:"id_unit,omitempty"` // Kompatibilitas dengan ERD
	IDProducts  uint   `json:"id_products"`
	ProductName string `json:"product_name,omitempty"`
	AssetCode   string `json:"asset_code"`
	Status      string `json:"status"` // 'available', 'rented', 'maintenance'
}

type CreateItemInstanceInput struct {
	IDProducts uint   `json:"id_products"`
	IDProduk   uint   `json:"id_produk,omitempty"` // Dukungan untuk nama field id_produk
	AssetCode  string `json:"asset_code"`
	Status     string `json:"status"`
}

func (i CreateItemInstanceInput) GetProductID() uint {
	if i.IDProducts != 0 {
		return i.IDProducts
	}
	return i.IDProduk
}
