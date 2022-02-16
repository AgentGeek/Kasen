package services

import (
	"context"
	"errors"
	"strconv"
	"time"

	. "kasen/cache"

	"kasen/config"
	"kasen/errs"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/rs1703/logger"
)

type Token struct {
	ID       string
	String   string
	Expr     int64
	ExprDate *time.Time
}

const SessionExpiration = 15 * time.Minute    // 15 minutes
const RefreshExpiration = 24 * 30 * time.Hour // 30 days

func createToken(uid int64, secret []byte, exprDur time.Duration) (*Token, error) {
	expr := time.Now().Add(exprDur)
	t := &Token{ID: uuid.NewString(), Expr: expr.Unix(), ExprDate: &expr}
	tRaw := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  t.ID,
		"exp": t.Expr,
	})

	var err error
	if t.String, err = tRaw.SignedString(secret); err != nil {
		return nil, err
	} else if err = Redis.Set(context.Background(), t.ID, uid, expr.Sub(time.Now())).Err(); err != nil {
		return nil, err
	}
	return t, nil
}

func CreateToken(uid int64) (rt *Token, st *Token, err error) {
	security := config.GetSecurity()

	rt, err = createToken(uid, security.JWTRefreshSecret, RefreshExpiration)
	if err != nil {
		logger.Err.Println(err)
		return nil, nil, errs.ErrUnknown
	}

	st, err = createToken(uid, security.JWTSessionSecret, SessionExpiration)
	if err != nil {
		logger.Err.Println(err)
		return nil, nil, errs.ErrUnknown
	}
	return
}

func parseToken(t string, secret []byte) (*jwt.Token, error) {
	return jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Failed to parse token")
		}
		return secret, nil
	})
}

func DeleteToken(rt string, secret []byte) error {
	t, err := parseToken(rt, secret)
	if err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return errs.ErrInvalidToken
	}

	id, ok := claims["id"].(string)
	if !ok {
		return errs.ErrInvalidToken
	}

	if _, err = Redis.Del(context.Background(), id).Result(); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}
	return nil
}

func verifyToken(tStr string, secret []byte) (uid int64, err error) {
	t, err := parseToken(tStr, secret)
	if err != nil {
		return 0, err
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return 0, errs.ErrInvalidToken
	}

	id, ok := claims["id"].(string)
	if !ok {
		return 0, errs.ErrInvalidToken
	}

	uidStr, err := Redis.Get(context.Background(), id).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(uidStr, 10, 64)
}

func VerifySessionToken(st string) (uid int64, err error) {
	if uid, err = verifyToken(st, config.GetSecurity().JWTSessionSecret); err != nil {
		logger.Err.Println(err)
		return 0, errs.ErrUnknown
	}
	return
}

func VerifyRefreshToken(rt string) (uid int64, err error) {
	if uid, err = verifyToken(rt, config.GetSecurity().JWTRefreshSecret); err != nil {
		logger.Err.Println(err)
		return 0, errs.ErrUnknown
	}
	return
}
