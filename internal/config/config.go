package config

import (
	"flag"
	"fmt"
)

type Config struct {
	Address       string
	Database      string
	NetworkPeriod int64
	GuardBefore   int64
	GuardAfter    int64
}

func Default() Config {
	return Config{Address: ":18080", Database: "tsnsched.db", NetworkPeriod: 1_000_000, GuardBefore: 100, GuardAfter: 100}
}

func Flags(args []string) (Config, bool, error) {
	c := Default()
	fs := flag.NewFlagSet("tsnsched", flag.ContinueOnError)
	smoke := fs.Bool("smoke-test", false, "run the deterministic smoke test")
	fs.StringVar(&c.Address, "addr", c.Address, "HTTP listen address")
	fs.StringVar(&c.Database, "db", c.Database, "SQLite database path")
	fs.Int64Var(&c.NetworkPeriod, "period-ns", c.NetworkPeriod, "network cycle period")
	fs.Int64Var(&c.GuardBefore, "guard-before-ns", c.GuardBefore, "guard band before a slot")
	fs.Int64Var(&c.GuardAfter, "guard-after-ns", c.GuardAfter, "guard band after a slot")
	if err := fs.Parse(args); err != nil {
		return c, false, err
	}
	if c.NetworkPeriod <= 0 || c.GuardBefore < 0 || c.GuardAfter < 0 {
		return c, false, fmt.Errorf("invalid scheduler configuration")
	}
	return c, *smoke, nil
}
