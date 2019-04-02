package dao

import (
	"context"
	"testing"
)

func TestDaopingRedis(t *testing.T) {
	err := d.pingRedis(context.Background())
}

func TestDaoAddServerOnline(t *testing.T) {
	var (
		c   = context.Background()
		rid = 12345
	)

	r, err := d.GetServerVid(c, rid)
}
