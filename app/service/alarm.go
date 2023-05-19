package service

// AlarmService 告警模块
type AlarmService interface{}

func Alarm() AlarmService {
	return &alarmService{}
}

type alarmService struct{}

func (alt *alarmService) name() {
}
