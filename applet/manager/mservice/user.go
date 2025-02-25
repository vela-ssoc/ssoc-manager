package mservice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"gorm.io/gen/field"
)

func NewUser(qry *query.Query, log *slog.Logger) *User {
	return &User{
		qry: qry,
		log: log,
	}
}

type User struct {
	qry *query.Query
	log *slog.Logger
}

func (usr *User) ResetAccessKey(ctx context.Context, uid int64) error {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	secret := hex.EncodeToString(buf)

	nano := time.Now().UnixNano()
	sid := strconv.FormatInt(uid, 32)
	sat := strconv.FormatInt(nano, 32)
	words := []string{"access", sid, sat, secret}
	ak := strings.Join(words, ".")

	tbl := usr.qry.User
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(uid)).
		UpdateSimple(tbl.AccessKey.Value(ak))

	return err
}

func (usr *User) CleanOTP(ctx context.Context, uid int64) error {
	tbl := usr.qry.User
	updates := []field.AssignExpr{
		tbl.TotpBind.Value(false),
		tbl.TotpSecret.Value(""),
	}
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(uid)).
		UpdateSimple(updates...)

	return err
}
