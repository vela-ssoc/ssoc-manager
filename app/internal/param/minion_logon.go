package param

import "time"

type MinionLogonAttack struct {
	Addr  string `json:"addr"`
	Msg   string `json:"msg"`
	Count int    `json:"count"`
}

type MinionLogonHistory struct {
	Page
	Name     string `json:"name"             query:"name"`
	MinionID int64  `json:"minion_id,string" query:"minion_id"`
}

type minionRecent struct {
	Date    string `json:"date"    gorm:"column:date"`
	Success int    `json:"success" gorm:"column:success"`
	Failed  int    `json:"failed"  gorm:"column:failed"`
	Logout  int    `json:"logout"  gorm:"column:logout"`
}

type MinionRecent []*minionRecent

type minionRecentTemp struct {
	Date  string `gorm:"column:date"`
	Msg   string `gorm:"column:msg"`
	Count int    `gorm:"column:count"`
}

type MinionRecentTemps []*minionRecentTemp

func (rts MinionRecentTemps) Format(n int) MinionRecent {
	ret := make(MinionRecent, 0, n)
	idx := make(map[string]*minionRecent, n)
	ondDay, now := 24*time.Hour, time.Now()
	for i := 0; i < n; i++ {
		du := time.Duration(i)
		format := now.Add(-(du * ondDay)).Format("01-02")
		rre := &minionRecent{Date: format}
		ret = append(ret, rre)
		idx[format] = rre
	}

	const success, failed, logout = "登录成功", "登录失败", "用户注销"
	for _, rt := range rts {
		date := rt.Date
		resp := idx[date]
		if resp == nil {
			continue
		}

		msg, count := rt.Msg, rt.Count
		switch msg {
		case success:
			resp.Success = count
		case failed:
			resp.Failed = count
		case logout:
			resp.Logout = count
		}
	}

	return ret
}

type MinionLogonCount struct {
	Success int `json:"success" gorm:"column:success"`
	Failed  int `json:"failed"  gorm:"column:failed"`
	Logout  int `json:"logout"  gorm:"column:logout"`
}
