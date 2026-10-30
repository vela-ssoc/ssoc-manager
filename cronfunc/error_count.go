package cronfunc

type errorContinuous struct {
	continuous int // 连续错误次数
}

// check 检查是否出错，并返回累计连续错误次数
func (ec *errorContinuous) check(err error) int {
	if err == nil {
		ec.continuous = 0
	} else {
		ec.continuous++
	}

	return ec.continuous
}

// threshold 检查是否出错，并返回累计连续错误次数是否大于等于某个阈值。
func (ec *errorContinuous) threshold(err error, gte int) bool {
	return ec.check(err) >= gte
}

func (ec *errorContinuous) threshold10(err error) bool {
	return ec.threshold(err, 10)
}
