package models

type Item struct {
	UserID string `gorm:"primaryKey"`
	Date   string `gorm:"primaryKey"`

	Title string
	Body  string
	Tags  StringList `gorm:"type:json"`
	// AssetIDs StringList `gorm:"type:json"`
}

// func (u Item) FromDB() goserver.Item {
// 	return goserver.Item{
// 		Email:     u.Login,
// 		StartDate: u.StartDate,
// 	}
// }
