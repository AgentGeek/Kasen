package services

import (
	"database/sql"
	"strings"

	. "kasen/database"

	"kasen/constants"
	"kasen/errs"
	"kasen/models"
	"kasen/modext"

	"github.com/rs1703/logger"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

var UserCols = models.UserColumns
var UserRels = models.UserRels

// hashPassword hashes the given rawPassword using bcrypt.
func hashPassword(rawPassword string) (string, error) {
	buf, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	return string(buf), err
}

// CreateUserOptions represents the parameters for creating a user.
type CreateUserOptions struct {
	Name        string `validate:"required,min=3,max=32"`
	Email       string `validate:"required,email,max=255"`
	RawPassword string `validate:"required,min=8"`
}

func (opts *CreateUserOptions) validate() error {
	opts.Name = strings.TrimSpace(opts.Name)
	opts.Email = strings.TrimSpace(opts.Email)

	switch {
	case len(opts.Name) == 0:
		return errs.ErrUserNameRequired
	case len(opts.Name) < 3:
		return errs.ErrUserNameTooShort
	case len(opts.Name) > 32:
		return errs.ErrUserNameTooLong
	case len(opts.Email) == 0:
		return errs.ErrEmailRequired
	case len(opts.Email) > 255:
		return errs.ErrEmailTooLong
	case !isEmail(opts.Email):
		return errs.ErrEmailInvalid
	case len(opts.RawPassword) == 0:
		return errs.ErrPasswordRequired
	case len(opts.RawPassword) < 6:
		return errs.ErrPasswordTooShort
	}
	return nil
}

// This function simply calls CreateUserEx with the global Write connection.
func CreateUser(opts CreateUserOptions) (*modext.User, error) {
	return CreateUserEx(WriteDB, opts)
}

// CreateUserEx creates a new user with the given options.
// Returns the created user if successful, or an error if user already exists or if
// the given options are invalid.
func CreateUserEx(e boil.Executor, opts CreateUserOptions) (*modext.User, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(opts.RawPassword)
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	user := &models.User{
		Name:     opts.Name,
		Email:    opts.Email,
		Password: hashedPassword,
		Permissions: []string{
			constants.PermCreateProject,
			constants.PermUploadCover,
			constants.PermSetCover,

			constants.PermEditUser,
			constants.PermDeleteUser,
			constants.PermCreateChapter,
			constants.PermEditChapter,
			constants.PermLockChapter,
			constants.PermPublishChapter,
			constants.PermUnlockChapter,
			constants.PermUnpublishChapter,
		},
	}

	if err := user.Insert(e, boil.Infer()); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewUser(user), nil
}

// This function simply calls GetUserEx with the global Read connection.
func GetUser(id int64) (*modext.User, error) {
	return GetUserEx(ReadDB, id)
}

// GetUserEx gets a user by ID.
// Returns the user if found, or an error if the user does not exist.
func GetUserEx(e boil.Executor, id int64) (*modext.User, error) {
	user, err := models.FindUser(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		}

		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewUser(user), nil
}

// This function simply calls GetUserByEmailEx with the global Read connection.
func GetUserByEmail(email string) (*modext.User, error) {
	return GetUserByEmailEx(ReadDB, email)
}

// GetUserByEmailEx gets a user by email.
// Returns the user if found, or an error if the user does not exist.
func GetUserByEmailEx(e boil.Executor, email string) (*modext.User, error) {
	user, err := models.Users(Where("email ILIKE ?", email)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		}

		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewUser(user), nil
}

// This function simply calls GetUsersEx with the global Read connection.
func GetUsers() ([]*modext.User, error) {
	return GetUsersEx(ReadDB)
}

// GetUsersEx gets all users.
func GetUsersEx(e boil.Executor) ([]*modext.User, error) {
	users, err := models.Users().All(e)
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	results := make([]*modext.User, len(users))
	for i, user := range users {
		results[i] = modext.NewUser(user)
	}

	return results, nil
}

// This function simply calls UpdateUserNameEx with the global Write connection.
func UpdateUserName(user *modext.User, name string) error {
	return UpdateUserNameEx(WriteDB, user, name)
}

// UpdateUserNameEx updates the name of the given user.
func UpdateUserNameEx(e boil.Executor, user *modext.User, name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return errs.ErrUserNameRequired
	} else if len(name) < 3 {
		return errs.ErrUserNameTooShort
	} else if len(name) > 32 {
		return errs.ErrUserNameTooLong
	}

	u := user.ToModel()
	u.Name = name
	user.Name = name

	if err := u.Update(WriteDB, boil.Whitelist(UserCols.Name, UserCols.UpdatedAt)); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls UpdateUserEmailEx with the global Write connection.
func UpdateUserEmail(user *modext.User, email string) error {
	return UpdateUserEmailEx(WriteDB, user, email)
}

// UpdateUserEmailEx updates the email of the given user.
// Returns an error if the email is already in use.
func UpdateUserEmailEx(e boil.Executor, user *modext.User, email string) error {
	email = strings.TrimSpace(email)

	if len(email) == 0 {
		return errs.ErrEmailRequired
	} else if len(email) > 255 {
		return errs.ErrEmailTooLong
	} else if !isEmail(email) {
		return errs.ErrEmailInvalid
	}

	if CheckUserExistsByEmailEx(e, email) {
		return errs.ErrEmailTaken
	}

	u := user.ToModel()
	u.Email = email

	if err := u.Update(e, boil.Whitelist(UserCols.Email, UserCols.UpdatedAt)); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// UpdateUserPasswordOptions represents the parameters for updating a user's password.
type UpdateUserPasswordOptions struct {
	CurrentRawPassword string `json:"currentPassword"`
	NewRawPassword     string `json:"newPassword"`
}

// This function simply calls UpdateUserPasswordEx with the global Write connection.
func UpdateUserPassword(user *modext.User, opts UpdateUserPasswordOptions) error {
	return UpdateUserPasswordEx(WriteDB, user, opts)
}

// UpdateUserPasswordEx updates the password of the given user with the given options.
// Returns an error if the current password is incorrect.
func UpdateUserPasswordEx(e boil.Executor, user *modext.User, opts UpdateUserPasswordOptions) error {
	if len(opts.CurrentRawPassword) == 0 {
		return errs.ErrCurrentPasswordRequired
	} else if len(opts.CurrentRawPassword) < 6 {
		return errs.ErrCurrentPasswordTooShort
	} else if len(opts.NewRawPassword) == 0 {
		return errs.ErrNewPasswordRequired
	} else if len(opts.NewRawPassword) < 6 {
		return errs.ErrNewPasswordTooShort
	}

	if err := user.CheckPassword(opts.CurrentRawPassword); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return errs.ErrInvalidCredentials
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	hashedPassword, err := hashPassword(opts.NewRawPassword)
	if err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	u := user.ToModel()
	u.Password = hashedPassword

	if err := u.Update(e, boil.Whitelist(UserCols.Password, UserCols.UpdatedAt)); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls AddUserPermissionEx with the global Write connection.
func AddUserPermission(user *modext.User, permission string) ([]string, error) {
	return AddUserPermissionEx(WriteDB, user, permission)
}

// AddUserPermissionEx adds the given permission to the given user.
// Returns the updated permissions of the user.
func AddUserPermissionEx(e boil.Executor, user *modext.User, permission string) ([]string, error) {
	if len(permission) == 0 {
		return nil, errs.ErrPermissionRequired
	}

	if !user.HasPermissions(permission) {
		u := user.ToModel()
		u.Permissions = append(u.Permissions, permission)
		user.Permissions = u.Permissions

		if err := u.Update(e, boil.Whitelist(UserCols.Permissions, UserCols.UpdatedAt)); err != nil {
			logger.Err.Println(err)
			return nil, errs.ErrUnknown
		}
	}

	return user.Permissions, nil
}

// This function simply calls UpdateUserPermissionsEx with the global Write connection.
func UpdateUserPermissions(user *modext.User, permissions []string) ([]string, error) {
	return UpdateUserPermissionsEx(WriteDB, user, permissions)
}

// UpdateUserPermissionsEx updates the permissions of the given user.
// Returns the updated permissions of the user.
func UpdateUserPermissionsEx(e boil.Executor, user *modext.User, permissions []string) ([]string, error) {
	u := user.ToModel()
	u.Permissions = permissions
	user.Permissions = u.Permissions

	if err := u.Update(e, boil.Whitelist(UserCols.Permissions, UserCols.UpdatedAt)); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return permissions, nil
}

// This function simply calls DeleteUserEx with the global Write connection.
func DeleteUser(user *modext.User) error {
	return DeleteUserEx(WriteDB, user)
}

// DeleteUserEx deletes the given user.
func DeleteUserEx(e boil.Executor, user *modext.User) error {
	if err := user.ToModel().Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls CheckUserExistsByEmailEx with the global Read connection.
func CheckUserExistsByEmail(email string) bool {
	return CheckUserExistsByEmailEx(ReadDB, email)
}

// CheckUserExistsByEmailEx checks if a user with the given email exists.
func CheckUserExistsByEmailEx(e boil.Executor, email string) bool {
	exists, _ := models.Users(Where("email ILIKE ?", email)).Exists(e)
	return exists
}
