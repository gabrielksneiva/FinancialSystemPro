package repositories

import "gorm.io/gorm"

type NewDatabase struct {
	DB *gorm.DB
}

func (d *NewDatabase) Insert(value any) error {
	return d.DB.Create(value).Error
}

func (db *NewDatabase) FindUserByField(field string, value any) (*User, error) {
	var user User
	err := db.DB.Model(&User{}).Where(field+" = ?", value).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
