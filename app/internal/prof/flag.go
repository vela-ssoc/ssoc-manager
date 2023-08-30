package prof

import (
	"flag"
	"strings"
)

// pprofFlag Mostly copied from https://github.com/google/pprof/blob/master/internal/driver/flags.go
type pprofFlag struct {
	Args  []string
	s     flag.FlagSet
	usage []string
}

// Bool implements the plugin.FlagSet interface.
func (p *pprofFlag) Bool(o string, d bool, c string) *bool {
	return p.s.Bool(o, d, c)
}

// Int implements the plugin.FlagSet interface.
func (p *pprofFlag) Int(o string, d int, c string) *int {
	return p.s.Int(o, d, c)
}

// Float64 implements the plugin.FlagSet interface.
func (p *pprofFlag) Float64(o string, d float64, c string) *float64 {
	return p.s.Float64(o, d, c)
}

// String implements the plugin.FlagSet interface.
func (p *pprofFlag) String(o, d, c string) *string {
	return p.s.String(o, d, c)
}

// BoolVar implements the plugin.FlagSet interface.
func (p *pprofFlag) BoolVar(b *bool, o string, d bool, c string) {
	p.s.BoolVar(b, o, d, c)
}

// IntVar implements the plugin.FlagSet interface.
func (p *pprofFlag) IntVar(i *int, o string, d int, c string) {
	p.s.IntVar(i, o, d, c)
}

// Float64Var implements the plugin.FlagSet interface.
// the value of the flag.
func (p *pprofFlag) Float64Var(f *float64, o string, d float64, c string) {
	p.s.Float64Var(f, o, d, c)
}

// StringVar implements the plugin.FlagSet interface.
func (p *pprofFlag) StringVar(s *string, o, d, c string) {
	p.s.StringVar(s, o, d, c)
}

// StringList implements the plugin.FlagSet interface.
func (p *pprofFlag) StringList(o, d, c string) *[]*string {
	return &[]*string{p.s.String(o, d, c)}
}

// AddExtraUsage implements the plugin.FlagSet interface.
func (p *pprofFlag) AddExtraUsage(eu string) {
	p.usage = append(p.usage, eu)
}

// ExtraUsage implements the plugin.FlagSet interface.
func (p *pprofFlag) ExtraUsage() string {
	return strings.Join(p.usage, "\n")
}

// Parse implements the plugin.FlagSet interface.
func (p *pprofFlag) Parse(usage func()) []string {
	p.s.Usage = usage
	_ = p.s.Parse(p.Args)
	args := p.s.Args()
	if len(args) == 0 {
		usage()
	}
	return args
}
