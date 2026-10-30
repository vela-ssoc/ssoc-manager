package migration

import (
	"context"
	"log/slog"
	"time"

	model2 "github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewUser(pdb polydb.Sourcer, log *slog.Logger) *User {
	return &User{
		pdb: pdb,
		log: log,
	}
}

func (u *User) Exec(ctx context.Context) error {
	mdb := u.pdb.MDB()
	coll := mdb.User()
	cnt, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil || cnt != 0 {
		return err
	}

	rdb := u.pdb.RDB()
	tbl := rdb.User
	dao := tbl.WithContext(ctx)

	users, err := dao.Find()
	if err != nil || len(users) == 0 {
		return err
	}

	now := time.Now()
	news := make([]*model.User, 0, len(users))
	for _, usr := range users {
		company := "group"
		if len(usr.Username) == 5 {
			company = "security"
		}

		mod := &model.User{
			Username: usr.Username,
			Nickname: usr.Nickname,
			Company:  company,
			Enabled:  usr.Enable,
			OAuth:    usr.Domain == model2.UdOA,
			Password: model.UserPassword{
				Secret:    usr.Password,
				ChangedAt: now,
			},
			Session: model.UserSession{
				Token:    usr.Token,
				IssuedAt: usr.IssueAt.Time,
				LastedAt: usr.SessionAt.Time,
			},
			TOTP: model.UserTOTP{
				Secret:     usr.TotpSecret,
				Verified:   usr.TotpBind,
				VerifiedAt: now,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
		news = append(news, mod)
	}
	_, err = coll.InsertMany(ctx, news)

	return err
}
