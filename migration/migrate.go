package migration

import "context"

type Executor interface {
	Exec(ctx context.Context) error
}

func Exec(ctx context.Context, exes ...Executor) error {
	for _, ex := range exes {
		if err := ex.Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}
