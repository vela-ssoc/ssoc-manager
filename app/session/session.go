package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/xgfone/ship/v5"
)

type Ident struct {
	ID       int64     `json:"id"`
	Username string    `json:"username"`
	Nickname string    `json:"nickname"`
	Token    string    `json:"token"`
	IssuedAt time.Time `json:"issued_at"`
}

type Session interface {
	ship.Session
	Destroy(id int64) error
}

// Issued 签发 session 信息
func Issued(u *model.User) *Ident {
	now := time.Now()
	id := strconv.FormatInt(u.ID, 32)
	nano := strconv.FormatInt(now.UnixNano(), 36)
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	nonce := hex.EncodeToString(buf)
	words := []string{"bearer", id, nano, nonce}
	token := strings.Join(words, ".")

	return &Ident{
		ID:       u.ID,
		Username: u.Username,
		Nickname: u.Nickname,
		Token:    token,
		IssuedAt: now,
	}
}

// Cast 将 v 断言为内部使用的 session struct
func Cast(v any) *Ident {
	ident, _ := v.(*Ident)
	return ident
}

// DBSess 数据库存放 session 管理器
func DBSess(interval time.Duration) Session {
	if interval <= 0 {
		interval = time.Hour
	} else if interval < time.Minute {
		interval = time.Minute
	}

	return &sessDB{
		interval: interval,
		timeout:  10 * time.Second,
	}
}

// sessDB 数据库存放 session 管理器
type sessDB struct {
	interval time.Duration // session 有效活动间隔
	timeout  time.Duration
}

// GetSession 根据 token 获取 session 信息
func (ssd *sessDB) GetSession(token string) (any, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ship.ErrSessionNotExist
	}

	ctx, cancel := context.WithTimeout(context.Background(), ssd.timeout)
	defer cancel()

	tbl := query.User
	dao := tbl.WithContext(ctx)
	user, err := dao.Where(tbl.Enable.Is(true)).
		Where(dao.Or(tbl.Token.Eq(token)).Or(tbl.AccessKey.Eq(token))).
		First()
	if err != nil {
		return nil, ship.ErrSessionNotExist
	}
	ident := &Ident{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Token:    token,
		IssuedAt: user.IssueAt.Time,
	}
	if user.AccessKey == token {
		return ident, nil
	}

	// 检查 session 是否有效
	now := time.Now()
	sessAt := user.SessionAt.Time
	expiredAt := sessAt.Add(ssd.interval)
	if expiredAt.Before(now) {
		return nil, ship.ErrSessionNotExist
	}

	// 如果用户每次获取 session 每次就续约，虽然 session 的精准度会更高，
	// 但是也会增加数据库的操作次数，且用户对 session 续期时间不是很敏感，
	// 所以此处会隔一段时间再对 session 实行续约。
	var renew bool
	diff := now.Sub(sessAt)
	if ssd.interval <= 10*time.Minute {
		renew = diff >= 30*time.Second
	} else {
		renew = diff >= 5*time.Minute
	}
	if renew {
		_, _ = tbl.WithContext(ctx).
			Where(tbl.Token.Eq(token)).
			UpdateColumn(tbl.SessionAt, now)
	}

	return ident, nil
}

// SetSession 存放 session
func (ssd *sessDB) SetSession(token string, val any) error {
	info := Cast(val)
	if info == nil {
		return ship.ErrInvalidSession
	}
	if info.Token != token {
		return ship.ErrInvalidSession
	}
	ctx, cancel := context.WithTimeout(context.Background(), ssd.timeout)
	defer cancel()

	now := sql.NullTime{Valid: true, Time: time.Now()}
	tbl := query.User
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(info.ID)).
		Where(tbl.Enable.Is(true)).
		UpdateColumnSimple(tbl.Token.Value(token), tbl.SessionAt.Value(now))

	return err
}

// DelSession 根据 token 删除 session
func (ssd *sessDB) DelSession(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ssd.timeout)
	defer cancel()

	tbl := query.User
	_, err := tbl.WithContext(ctx).
		Where(tbl.Token.Eq(token)).
		UpdateColumnSimple(tbl.Token.Value(""))

	return err
}

func (ssd *sessDB) Destroy(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), ssd.timeout)
	defer cancel()

	tbl := query.User
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		UpdateColumn(tbl.Token, "")

	return err
}
