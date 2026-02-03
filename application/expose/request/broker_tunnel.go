package request

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/time/rate"
)

type BrokerTunnelLimit struct {
	Int64ID
	Unlimit bool      `json:"unlimit" query:"unlimit"`
	Limit   ByteCount `json:"limit"   query:"limit"   validate:"gte=0"`
}

type BrokerTunnelSpeedtest struct {
	Int64ID
	Size ByteCount `json:"size" query:"size" validate:"gte=0"`
}

type ByteCount float64

func (bc *ByteCount) UnmarshalText(text []byte) error {
	return bc.UnmarshalBind(string(text))
}

func (bc *ByteCount) UnmarshalBind(str string) error {
	str = strings.ToLower(strings.TrimSpace(str))
	if str == "inf" {
		*bc = ByteCount(rate.Inf)
		return nil
	}

	input := []byte(str)
	var idx int
	for _, r := range input {
		if (r >= '0' && r <= '9') || r == '.' || r == '+' || r == '-' {
			idx++
			continue
		}
		break
	}

	numeric := strings.TrimSpace(string(input[:idx]))
	unit := strings.TrimSpace(string(input[idx:]))
	unit = strings.ToLower(unit)
	if numeric == "" {
		return errors.New("missing numeric value")
	}

	val, err := strconv.ParseFloat(numeric, 64)
	if err != nil {
		return err
	}

	switch unit {
	case "", "b":
	case "k", "kb", "kib":
		val *= 1024
	case "m", "mb", "mib":
		val *= 1024 * 1024
	case "g", "gb", "gib":
		val *= 1024 * 1024 * 1024
	case "t", "tb", "tib":
		val *= 1024 * 1024 * 1024 * 1024
	default:
		return errors.New("invalid unit: " + unit)
	}
	*bc = ByteCount(val)

	return nil
}
