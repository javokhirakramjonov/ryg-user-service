package model

type User struct {
	ID       int32  `json:"id" gorm:"primaryKey;autoIncrement"`
	FullName string `json:"full_name" gorm:"column:full_name"`
	Password string `json:"password"`
	Email    string `json:"email" gorm:"unique"`
	Role     string `json:"role" gorm:"type:varchar(10)"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
}

func (User) TableName() string {
	return "users"
}
