package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/text3cn/goodle/providers/goodlog"
	"gtiv/app/imbiz/im/model"
	"strconv"

	"github.com/zhenjl/cityhash"
)

const (
	_prefixMidServer    = "imuid:%d" // mid -> key:server
	_prefixKeyServer    = "imkey:%s" // key -> server
	_prefixServerOnline = "ol_%s"    // server -> online
)

func keyMidServer(mid int64) string {
	return fmt.Sprintf(_prefixMidServer, mid)
}

func keyKeyServer(key string) string {
	return fmt.Sprintf(_prefixKeyServer, key)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(_prefixServerOnline, key)
}

// pingRedis check redis connection.
func (d *Dao) pingRedis(c context.Context) (err error) {
	conn := d.redis
	conn.Do(c, "SET", "PING", "PONG")
	return
}

// AddMapping add a mapping.
// Mapping:
//
//	mid -> key_server
//	key -> server
func (d *Dao) AddMapping(ctx context.Context, mid int64, key, server string) (err error) {
	conn := d.redis
	if mid > 0 {
		result := conn.Do(ctx, "HSET", keyMidServer(mid), key, server)
		if result.Err() != nil {
			goodlog.Errorf("conn.Send(HSET %d,%s,%s) error(%v)", mid, server, key, err)
		}
		if result = conn.Do(ctx, "EXPIRE", keyMidServer(mid), d.redisExpire); err != nil {
			goodlog.Errorf("conn.Send(EXPIRE %d,%s,%s) error(%v)", mid, key, server, err)
		}
	}
	if result := conn.Do(ctx, "SET", keyKeyServer(key), server); result.Err() != nil {
		goodlog.Errorf("conn.Send(HSET %d,%s,%s) error(%v)", mid, server, key, err)
		return
	}
	if result := conn.Do(ctx, "EXPIRE", keyKeyServer(key), d.redisExpire); result.Err() != nil {
		goodlog.Errorf("conn.Send(EXPIRE %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	return
}

// ExpireMapping expire a mapping.
func (d *Dao) ExpireMapping(ctx context.Context, mid int64, key string) (has bool, err error) {
	conn := d.redis
	var n = 1
	if mid > 0 {
		if res := conn.Do(ctx, "EXPIRE", keyMidServer(mid), d.redisExpire); res.Err() != nil {
			goodlog.Errorf("conn.Send(EXPIRE %d,%s) error(%v)", mid, key, err)
			return
		}
		n++
	}
	if res := conn.Do(ctx, "EXPIRE", keyKeyServer(key), d.redisExpire); res.Err() != nil {
		goodlog.Errorf("conn.Send(EXPIRE %d,%s) error(%v)", mid, key, err)
		return
	}
	return
}

// DelMapping del a mapping.
func (d *Dao) DelMapping(ctx context.Context, mid int64, key, server string) (has bool, err error) {
	conn := d.redis
	if mid > 0 {
		if result := conn.Do(ctx, "HDEL", keyMidServer(mid), key); result.Err() != nil {
			goodlog.Errorf("conn.Send(HDEL %d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	if result := conn.Do(ctx, "DEL", keyKeyServer(key)); result.Err() != nil {
		goodlog.Errorf("conn.Send(HDEL %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	return
}

// ServersByKeys get a server by key.
func (d *Dao) ServersByKeys(ctx context.Context, keys []string) (res []string, err error) {
	conn := d.redis
	var args []interface{}
	for _, key := range keys {
		args = append(args, keyKeyServer(key))
	}
	if result := conn.Do(ctx, "MGET", args); result.Err() != nil {
		goodlog.Errorf("conn.Do(MGET %v) error(%v)", args, err)
	}
	return
}

// KeysByMids get a key server by mid.
func (d *Dao) KeysByMids(ctx context.Context, mids []int64) (ret map[string]string, olMids []int64, err error) {
	conn := d.redis
	ret = make(map[string]string)
	for _, mid := range mids {
		if result := conn.HGetAll(ctx, keyMidServer(mid)); result.Err() != nil {
			goodlog.Errorf("conn.Do(HGETALL %d) error(%v)", mid, err)
			return
		} else {
			mp := result.Val()
			if mp != nil {
				for k, v := range mp {
					ret[k] = v
					olMids = append(olMids, mid)
				}
			}
		}
	}
	//if err = conn.Flush(); err != nil {
	//	log.Errorf("conn.Flush() error(%v)", err)
	//	return
	//}
	//for idx := 0; idx < len(mids); idx++ {
	//	var (
	//		res map[string]string
	//	)
	//	if res, err = redis.StringMap(conn.Receive()); err != nil {
	//		log.Errorf("conn.Receive() error(%v)", err)
	//		return
	//	}
	//	if len(res) > 0 {
	//		olMids = append(olMids, mids[idx])
	//	}
	//	for k, v := range res {
	//		ress[k] = v
	//	}
	//}
	return
}

// AddServerOnline add a server online.
func (d *Dao) AddServerOnline(c context.Context, server string, online *model.Online) (err error) {
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64] = rMap
		}
		rMap[room] = count
	}
	key := keyServerOnline(server)
	for hashKey, value := range roomsMap {
		err = d.addServerOnline(c, key, strconv.FormatInt(int64(hashKey), 10), &model.Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return
		}
	}
	return
}

func (d *Dao) addServerOnline(ctx context.Context, key string, hashKey string, online *model.Online) (err error) {
	conn := d.redis
	b, _ := json.Marshal(online)
	if result := conn.Do(ctx, "HSET", key, hashKey, b); result.Err() != nil {
		goodlog.Errorf("conn.Send(SET %s,%s) error(%v)", key, hashKey, err)
		return
	}
	if result := conn.Do(ctx, "EXPIRE", key, d.redisExpire); result.Err() != nil {
		goodlog.Errorf("conn.Send(EXPIRE %s) error(%v)", key, err)
		return
	}
	return
}

// ServerOnline get a server online.
func (d *Dao) ServerOnline(c context.Context, server string) (online *model.Online, err error) {
	online = &model.Online{RoomCount: map[string]int32{}}
	key := keyServerOnline(server)
	for i := 0; i < 64; i++ {
		ol, err := d.serverOnline(c, key, strconv.FormatInt(int64(i), 10))
		if err == nil && ol != nil {
			online.Server = ol.Server
			if ol.Updated > online.Updated {
				online.Updated = ol.Updated
			}
			for room, count := range ol.RoomCount {
				online.RoomCount[room] = count
			}
		}
	}
	return
}

func (d *Dao) serverOnline(ctx context.Context, key string, hashKey string) (online *model.Online, err error) {
	//conn := d.redis
	//b, err :=  conn.Do(ctx, "HGET", key, hashKey)
	//if err != nil {
	//	if err != redis.ErrNil {
	//		log.Errorf("conn.Do(HGET %s %s) error(%v)", key, hashKey, err)
	//	}
	//	return
	//}
	//online = new(model.Online)
	//if err = json.Unmarshal(b, online); err != nil {
	//	log.Errorf("serverOnline json.Unmarshal(%s) error(%v)", b, err)
	//	return
	//}
	return
}

// DelServerOnline del a server online.
func (d *Dao) DelServerOnline(ctx context.Context, server string) (err error) {
	conn := d.redis
	key := keyServerOnline(server)
	if result := conn.Do(ctx, "DEL", key); result.Err() != nil {
		goodlog.Errorf("conn.Do(DEL %s) error(%v)", key, err)
	}
	return
}
