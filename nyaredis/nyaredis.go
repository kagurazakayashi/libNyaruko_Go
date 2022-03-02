package nyaredis

import (
	"strconv"

	redis "github.com/go-redis/redis/v8"
	cmap "github.com/orcaman/concurrent-map"
)

type NyaRedis NyaRedisT
type NyaRedisT struct {
	db *redis.Client
}

func Init(confCMap cmap.ConcurrentMap) (NyaRedis, error) {
	redisaddress, _ := confCMap.Get("redis_addr")
	redisport, _ := confCMap.Get("redis_port")
	redispassword, _ := confCMap.Get("redis_pwd")
	redisdbidstr, _ := confCMap.Get("redis_db")
	redisdbid, _ := strconv.Atoi(redisdbidstr.(string))
	nRedisDB := redis.NewClient(&redis.Options{
		Addr:     redisaddress.(string) + ":" + redisport.(string),
		Password: redispassword.(string),
		DB:       redisdbid,
	})
	_, err := nRedisDB.Ping(nRedisDB.Context()).Result()
	if err != nil {
		return NyaRedis{}, err
	}
	return NyaRedis{db: nRedisDB}, nil
}

func (p NyaRedis) Close() {
	p.db.Close()
	p.db = nil
}
