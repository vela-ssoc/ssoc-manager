package param

import "time"

type RiskAttack struct {
	Subject  string `json:"subject"   gorm:"column:subject"`
	RemoteIP string `json:"remote_ip" gorm:"column:remote_ip"`
	Count    int    `json:"count"     gorm:"column:count"`
}

type riskRecentTemp struct {
	Date     string `gorm:"column:date"`
	RiskType string `gorm:"column:risk_type"`
	Count    int    `gorm:"column:count"`
}

type RiskRecentTemps []*riskRecentTemp

type riskRecentChart struct {
	Type  string `json:"type"`
	Count []int  `json:"count"`
}

type RiskRecentChart []*riskRecentChart

type RecentCharts struct {
	Date []string           `json:"date"`
	Risk []*riskRecentChart `json:"risk"`
}

func (rrs RiskRecentTemps) Charts(days int) *RecentCharts {
	index := make(map[string]*riskRecentChart, days)
	serial := make(map[string]int, days)
	dates := make([]string, days) // 创建并初始化
	day := 24 * time.Hour
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.Add(-time.Duration(i) * day).Format("01-02")
		dates[i] = date
		serial[date] = i
	}

	charts := make([]*riskRecentChart, 0, days)
	for _, rr := range rrs {
		tp, date := rr.RiskType, rr.Date
		chart := index[tp]
		if chart == nil {
			chart = &riskRecentChart{Type: tp, Count: make([]int, days)}
			index[tp] = chart
			charts = append(charts, chart)
		}
		if n, exist := serial[date]; exist {
			chart.Count[n] = rr.Count
		}
	}

	return &RecentCharts{
		Date: dates,
		Risk: charts,
	}
}
