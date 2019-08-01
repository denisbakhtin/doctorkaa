package models

import (
	"golang.org/x/crypto/bcrypt"
)

//Login view model
type Login struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required"`
}

//Forgot view model
type Forgot struct {
	Email string `form:"email" binding:"required"`
}

//Reset password view model
type Reset struct {
	Hash            string `form:"hash" binding:"required"`
	Password        string `form:"password" binding:"required"`
	PasswordConfirm string `form:"password_confirm" binding:"required"`
}

//Register view model
type Register struct {
	Email           string `form:"email" binding:"required"`
	Password        string `form:"password" binding:"required"`
	PasswordConfirm string `form:"password_confirm" binding:"required"`
}

//Manage user view model
type Manage struct {
	Email           string `form:"email" binding:"required"`
	Password        string `form:"password" binding:"required"`
	PasswordConfirm string `form:"password_confirm" binding:"required"`
	NewPassword     string `form:"new_password" binding:"required"`
}

//User type contains user info
type User struct {
	ID uint64 `gorm:"primary_key" form:"id"`

	Email      string `form:"email" binding:"required"`
	Password   string `form:"password" binding:"required"`
	ForgotHash string `binding:"-"`
}

//BeforeCreate gorm hook
func (u *User) BeforeCreate() (err error) {
	return u.HashPassword()
}

//HashPassword replaces _raw_ password with its hash
func (u *User) HashPassword() (err error) {
	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	u.Password = string(hash)
	return
}

//GenerateForgotHash generates secure hash for password restore
func (u *User) GenerateForgotHash() error {
	str, err := GenerateRandomStringURLSafe(50)
	if err != nil {
		return err
	}
	u.ForgotHash = str
	return nil
}
