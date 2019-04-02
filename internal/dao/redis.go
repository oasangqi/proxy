package dao

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
)

const (
	_prefixKeyRoom = "htid:%d" // mid -> key:server
)

func keyRoom(rid int64) string {
	return fmt.Sprintf(_prefixKeyRoom, rid)
}

// pingRedis check redis connection.
func (d *Dao) pingRedis(c context.Context) (err error) {
	conn := d.redis.Get()
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}

func (d *Dao) GetServerVid(c context.Context, rid int64) (vid int64, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	vid, err = redis.Int64(conn.Do("HGET", keyRoom(rid), "vid"))
	if err != nil {
		if err != redis.ErrNil {
			glog.Errorf("conn.Do(HGET %s vid) error(%v)", keyRoom(rid), err)
		}
		return
	}
	return
}
