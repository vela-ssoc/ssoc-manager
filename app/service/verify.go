package service

import (
	"context"
	"fmt"
	"image"
	"math/rand"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/integration/dong"
	"github.com/vela-ssoc/vela-common-mb/integration/formwork"
	"github.com/vela-ssoc/vela-common-mb/logback"
	"github.com/vela-ssoc/vela-manager/app/internal/modview"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/wenlng/go-captcha/captcha"
)

type VerifyService interface {
	// Picture 生成图片验证码
	Picture(ctx context.Context, uname string) (*param.AuthPicture, error)

	// Verify 验证图片验证码是否正确，并返回是否需要多因子认证
	Verify(ctx context.Context, uname, captID string, points []*image.Point) (factor bool, err error)

	// DongCode 发送咚咚验证码
	DongCode(ctx context.Context, uname, captID string, view modview.LoginDong) error

	// Submit 最终登录验证
	Submit(ctx context.Context, uname, captID, dongCode string) error
}

func Verify(minute int, dcli dong.Client, tmpl formwork.Render, slog logback.Logger) VerifyService {
	capt := captcha.GetCaptcha()
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	if minute < 1 || minute > 10 {
		minute = 3
	}

	return &verifyService{
		slog:   slog,
		minute: minute,
		dcli:   dcli,
		tmpl:   tmpl,
		capt:   capt,
		random: random,
		expire: time.Duration(minute) * time.Minute,
		valids: make(map[string]*validInfo, 16),
	}
}

type verifyService struct {
	slog   logback.Logger
	minute int // 验证码有效分钟
	dcli   dong.Client
	tmpl   formwork.Render
	capt   *captcha.Captcha
	random *rand.Rand
	expire time.Duration // 验证码有效期
	mutex  sync.RWMutex
	valids map[string]*validInfo
}

func (vs *verifyService) Picture(ctx context.Context, uname string) (*param.AuthPicture, error) {
	dots, lb64, sb64, captID, err := vs.capt.Generate()
	if err != nil {
		return nil, err
	}
	size := len(dots)
	points := make([]captcha.CharDot, size)
	for i := 0; i < size; i++ {
		points[i] = dots[i]
	}

	tbl := query.User
	user, err := tbl.WithContext(ctx).
		Select(tbl.Dong).
		Where(tbl.Enable.Is(true)).
		Where(tbl.Username.Eq(uname)).
		First()
	factor := true
	if err == nil && user != nil {
		factor = user.Dong != "" && vs.dcli != nil
	}

	vs.storeValidInfo(points, uname, captID, factor)
	pic := &param.AuthPicture{
		ID:    captID,
		Board: lb64,
		Thumb: sb64,
	}

	return pic, nil
}

func (vs *verifyService) Verify(_ context.Context, uname, captID string, points []*image.Point) (dong bool, err error) {
	if vi := vs.loadValidInfo(uname); vi != nil {
		need, passed := vi.verify(captID, points)
		if !passed {
			return false, errcode.ErrPictureCode
		}
		return need, nil
	}
	return false, errcode.ErrPictureCode
}

func (vs *verifyService) DongCode(ctx context.Context, uname, captID string, view modview.LoginDong) error {
	vi := vs.loadValidInfo(uname)
	if vi == nil || !vi.verified(captID) {
		return errcode.ErrPictureCode
	}
	if !vi.dong {
		return errcode.ErrWithoutDongCode
	}

	num := vs.random.Intn(1_000_000) // 0-999999
	code := fmt.Sprintf("%06d", num) // 0 填充： 123 -> 000123
	vi.setDongCode(code)
	view.Code = code
	view.Minute = vs.minute

	// 查询用户信息
	tbl := query.User
	user, err := tbl.WithContext(ctx).
		Select(tbl.Dong).
		Where(tbl.Enable.Is(true)).
		Where(tbl.Username.Eq(uname)).
		First()
	if err != nil || user == nil || user.Dong == "" {
		return nil
	}

	title, body := vs.tmpl.LoginDong(ctx, "", view)
	if err = vs.dcli.Send(ctx, []string{user.Dong}, nil, title, body); err != nil {
		vs.slog.Warnf("发送咚咚验证码错误：%s", err)
	}

	return err
}

func (vs *verifyService) Submit(_ context.Context, uname, captID, dongCode string) error {
	vi := vs.loadValidInfo(uname)
	if vi == nil {
		return errcode.ErrPictureCode
	}
	if vi.submit(captID, dongCode) {
		return nil
	}
	return errcode.ErrDongCode
}

func (vs *verifyService) loadValidInfo(uname string) *validInfo {
	vs.mutex.RLock()
	vi := vs.valids[uname]
	vs.mutex.RUnlock()

	return vi
}

func (vs *verifyService) storeValidInfo(points []captcha.CharDot, uname, captID string, dong bool) {
	vi := &validInfo{
		points: points,
		uname:  uname,
		captID: captID,
		dong:   dong,
		expire: vs.expire,
		picAt:  time.Now(),
	}
	vs.mutex.Lock()
	vs.valids[uname] = vi
	vs.mutex.Unlock()
}

type validInfo struct {
	points   []captcha.CharDot
	uname    string
	captID   string
	dong     bool
	picUsed  bool
	picOK    bool
	expire   time.Duration
	picAt    time.Time
	dongAt   time.Time
	dongCode string
	dongUsed bool
	mutex    sync.Mutex
}

func (vi *validInfo) verify(captID string, points []*image.Point) (dong bool, passed bool) {
	size := len(points)
	now := time.Now()

	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	invalid := size == 0 ||
		vi.picUsed ||
		vi.captID != captID ||
		size != len(vi.points) ||
		now.After(vi.picAt.Add(vi.expire))
	vi.picUsed = true
	if invalid {
		return false, false
	}

	for i, point := range points {
		dot := vi.points[i]
		in := captcha.CheckPointDistWithPadding(int64(point.X), int64(point.Y),
			int64(dot.Dx), int64(dot.Dy), int64(dot.Width), int64(dot.Height), 6)
		if passed = in; !passed { // 错误一次就失效过期
			break
		}
	}
	if passed {
		vi.picOK = true
		if dong = vi.dong; dong {
			vi.picAt = time.Now()
		}
		return dong, true
	}

	return false, false
}

func (vi *validInfo) verified(captID string) bool {
	now := time.Now()
	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	vi.picUsed = true

	return vi.picOK &&
		vi.captID == captID &&
		now.Before(vi.picAt.Add(vi.expire))
}

func (vi *validInfo) submit(captID, dongCode string) bool {
	now := time.Now()

	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	pass := vi.picOK &&
		!vi.dongUsed &&
		vi.captID == captID &&
		dongCode == vi.dongCode &&
		now.Before(vi.dongAt.Add(vi.expire))
	vi.dongUsed = true

	return pass
}

func (vi *validInfo) setDongCode(code string) {
	vi.mutex.Lock()
	vi.dongCode = code
	vi.dongAt = time.Now()
	vi.mutex.Unlock()
}
