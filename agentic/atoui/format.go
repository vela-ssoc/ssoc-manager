package atoui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func joinValueUnit(value, unit string) string {
	value = strings.TrimSpace(value)
	unit = strings.TrimSpace(unit)

	switch {
	case value == "":
		return ""
	case unit == "":
		return value
	case unit == "%":
		return value + unit
	default:
		return value + " " + unit
	}
}

func trendText(trend, delta string) string {
	delta = strings.TrimSpace(delta)
	if delta == "" {
		return ""
	}

	switch strings.ToLower(strings.TrimSpace(trend)) {
	case "up":
		return "↑ " + delta
	case "down":
		return "↓ " + delta
	case "flat":
		return "→ " + delta
	default:
		return delta
	}
}

func formatCell(value any, format string) (string, error) {
	if value == nil {
		return "", nil
	}

	switch format {
	case "", "text":
		return fmt.Sprint(value), nil
	case "number":
		n, err := toFloat(value)
		if err != nil {
			return "", err
		}
		return formatNumber(n), nil
	case "percent":
		n, err := toFloat(value)
		if err != nil {
			return "", err
		}
		return formatNumber(n) + "%", nil
	case "bytes":
		n, err := toFloat(value)
		if err != nil {
			return "", err
		}
		return formatBytes(n), nil
	case "duration":
		n, err := toFloat(value)
		if err != nil {
			return "", err
		}
		return formatDuration(n), nil
	case "datetime":
		return formatDateTime(value), nil
	default:
		return "", fmt.Errorf("format=%q 不支持", format)
	}
}

func toFloat(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	default:
		return 0, fmt.Errorf("%v 不是数字", value)
	}
}

func formatNumber(value float64) string {
	text := strconv.FormatFloat(value, 'f', -1, 64)
	sign := ""
	if strings.HasPrefix(text, "-") {
		sign = "-"
		text = text[1:]
	}

	integer, fraction, found := strings.Cut(text, ".")
	if len(integer) > 3 {
		var builder strings.Builder
		offset := len(integer) % 3
		if offset > 0 {
			builder.WriteString(integer[:offset])
		}
		for i := offset; i < len(integer); i += 3 {
			if builder.Len() > 0 {
				builder.WriteByte(',')
			}
			builder.WriteString(integer[i : i+3])
		}
		integer = builder.String()
	}

	if found {
		return sign + integer + "." + fraction
	}

	return sign + integer
}

func formatBytes(value float64) string {
	const unit = 1024

	if value < unit {
		return strconv.FormatFloat(value, 'f', -1, 64) + " B"
	}

	units := [...]string{"KB", "MB", "GB", "TB", "PB", "EB"}
	abs := value
	index := -1
	for abs >= unit && index < len(units)-1 {
		abs /= unit
		index++
	}

	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(abs, 'f', 2, 64), "0"), ".") + " " + units[index]
}

func formatDuration(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}

	total := int64(seconds)
	days := total / 86400
	hours := (total % 86400) / 3600
	minutes := (total % 3600) / 60
	secs := total % 60

	parts := make([]string, 0, 4)
	if days > 0 {
		parts = append(parts, strconv.FormatInt(days, 10)+"d")
	}
	if hours > 0 {
		parts = append(parts, strconv.FormatInt(hours, 10)+"h")
	}
	if minutes > 0 {
		parts = append(parts, strconv.FormatInt(minutes, 10)+"m")
	}
	if secs > 0 || len(parts) == 0 {
		parts = append(parts, strconv.FormatInt(secs, 10)+"s")
	}

	return strings.Join(parts, " ")
}

func formatDateTime(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case float64:
		return time.Unix(int64(v), 0).Format("2006-01-02 15:04:05")
	case int64:
		return time.Unix(v, 0).Format("2006-01-02 15:04:05")
	case int:
		return time.Unix(int64(v), 0).Format("2006-01-02 15:04:05")
	case json.Number:
		seconds, err := v.Int64()
		if err != nil {
			return v.String()
		}
		return time.Unix(seconds, 0).Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprint(value)
	}
}
