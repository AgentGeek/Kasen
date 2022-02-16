package modext

import (
	"kasen/models"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Password    string   `json:"-"`
	Email       string   `json:"-"`
	Permissions []string `json:"permissions,omitempty"`

	Chapters []*Chapter `json:"-"`
}

func NewUser(user *models.User) *User {
	if user == nil {
		return nil
	}
	return &User{
		ID:          user.ID,
		Name:        user.Name,
		Password:    user.Password,
		Email:       user.Email,
		Permissions: user.Permissions,
	}
}

func (u *User) LoadChapters(user *models.User) *User {
	if user == nil || user.R == nil || len(user.R.Chapters) == 0 {
		return u
	}

	u.Chapters = make([]*Chapter, len(user.R.Chapters))
	for i, chapter := range user.R.Chapters {
		u.Chapters[i] = NewChapter(chapter)
	}

	return u
}

func (u *User) ToModel() *models.User {
	return &models.User{
		ID:          u.ID,
		Name:        u.Name,
		Password:    u.Password,
		Email:       u.Email,
		Permissions: u.Permissions,
	}
}

// CheckPassword checks if the given password matches the user's password.
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// HasPermissions checks if the user has one of the given permissions.
func (u *User) HasPermissions(perms ...string) bool {
	if len(u.Permissions) == 0 {
		return false
	}
	for _, perm := range perms {
		for _, userPerm := range u.Permissions {
			if userPerm == perm {
				return true
			}
		}
	}
	return false
}
