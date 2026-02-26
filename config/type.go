package config

import "time"

type Duration time.Duration

func (d *Duration) UnmarshalText(b []byte) error {
	du, err := time.ParseDuration(string(b))
	if err == nil {
		*d = Duration(du)
	}

	return err
}

func (d Duration) MarshalText() ([]byte, error) {
	s := time.Duration(d).String()
	return []byte(s), nil
}
