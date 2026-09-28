package services

import (
	"belajar_go/config"
	"belajar_go/models"
)

func CreateRental(input models.CreateRentalInput) (models.Rental, error) {

	query := `INSERT INTO rentals (id_user, id_unit, delivery_address, start_date, return_date, rental_fee) 
	          VALUES (?, ?, ?, ?, ?, ?)`

	result, err := config.DB.Exec(query, input.IDUser, input.IDUnit, input.DeliveryAddress, input.StartDate, input.ReturnDate, input.RentalFee)
	if err != nil {
		return models.Rental{}, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return models.Rental{}, err
	}

	updateStatusQuery := `UPDATE item_instances SET status = 'rented' WHERE id = ?`
	_, err = config.DB.Exec(updateStatusQuery, input.IDUnit)
	if err != nil {
		return models.Rental{}, err
	}

	newRental := models.Rental{
		ID:              uint(lastID),
		IDUser:          input.IDUser,
		IDUnit:          input.IDUnit,
		DeliveryAddress: input.DeliveryAddress,
		StartDate:       input.StartDate,
		ReturnDate:      input.ReturnDate, 
		RentalFee:       input.RentalFee,
	}

	return newRental, nil
}
