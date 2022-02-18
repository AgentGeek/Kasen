package services

import (
	"log"

	"kasen/config"
	"kasen/errs"

	"golang.org/x/crypto/bcrypt"
)

// LoginOptions represents the parameters for logging in.
type LoginOptions struct {
	Email       string
	RawPassword string
}

// Login logs in a user with the given options
// and returns a new refresh and session token if successful,
// or an error if user does not exist or the password is incorrect.
func Login(opts LoginOptions) (rt, st *Token, err error) {
	switch {
	case len(opts.Email) == 0:
		return nil, nil, errs.ErrEmailRequired
	case len(opts.Email) > 255:
		return nil, nil, errs.ErrEmailTooLong
	case len(opts.RawPassword) == 0:
		return nil, nil, errs.ErrPasswordRequired
	case len(opts.RawPassword) < 6:
		return nil, nil, errs.ErrPasswordTooShort
	case !isEmail(opts.Email):
		return nil, nil, errs.ErrEmailInvalid
	}

	u, err := GetUserByEmail(opts.Email)
	if err != nil {
		if err != errs.ErrUserNotFound {
			log.Println(err)
		}
		return nil, nil, errs.ErrInvalidCredentials
	}

	if err := u.CheckPassword(opts.RawPassword); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, nil, errs.ErrInvalidCredentials
		}
		log.Println(err)
		return nil, nil, errs.ErrUnknown
	}

	rt, st, err = CreateToken(u.ID)
	if err != nil {
		return nil, nil, err
	}
	return rt, st, nil
}

// Logout deletes the given session and refresh tokens.
func Logout(st, rt string) error {
	security := config.GetSecurity()
	DeleteToken(st, security.JWTSessionSecret)
	return DeleteToken(rt, security.JWTRefreshSecret)
}

// Register creates a new user with the given registration options
// and returns a new refresh and session token if successful,
// or an error if user already exists.
func Register(opts CreateUserOptions) (rt, st *Token, err error) {
	u, err := CreateUser(opts)
	if err != nil {
		return nil, nil, err
	}

	rt, st, err = CreateToken(u.ID)
	if err != nil {
		return nil, nil, err
	}
	return rt, st, nil
}

// RefreshToken refreshes session token using the given refresh token,
// returns uid of the user and a new session token if successful,
// or an error if the refresh token is invalid.
func RefreshToken(rt string) (uid int64, st *Token, err error) {
	uid, err = VerifyRefreshToken(rt)
	if err != nil {
		return 0, nil, err
	}

	st, err = createToken(uid, config.GetSecurity().JWTSessionSecret, SessionExpiration)
	if err != nil {
		log.Println(err)
		return 0, nil, errs.ErrUnknown
	}
	return uid, st, nil
}
