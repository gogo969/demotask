package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"task/contrib/helper"
	"time"
)

var (
	// The maximum duration to lock a key, Default: 10s
	LockTimeout time.Duration = 20 * time.Second
	// The maximum duration to wait to get the lock, Default: 0s, do not wait
	//WaitTimeout time.Duration
	// The maximum wait retry time to get the lock again, Default: 100ms
	WaitRetry time.Duration = 100 * time.Millisecond
)

const (
	defaultRedisKeyPrefix = "rlock:"
)

var (
	ctx = context.Background()
)

func Lock(r *redis.ClusterClient, id string) error {

	val := fmt.Sprintf("%s%s", defaultRedisKeyPrefix, id)
	ok, err := r.SetNX(ctx, val, "1", LockTimeout).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New(helper.RequestBusy)
	}

	return nil
}

func LockWait(r *redis.ClusterClient, id string, ttl time.Duration) error {

	val := fmt.Sprintf("%s%s", defaultRedisKeyPrefix, id)

	for {
		ok, err := r.SetNX(ctx, val, "1", ttl).Result()
		if err != nil {
			return err
		}

		if !ok {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return nil
	}
}

func LockTTL(r *redis.ClusterClient, id string, ttl time.Duration) error {

	val := fmt.Sprintf("%s%s", defaultRedisKeyPrefix, id)
	ok, err := r.SetNX(ctx, val, "1", ttl).Result()
	if err != nil || !ok {
		return err
	}

	return nil
}

func LockSetExpire(r *redis.ClusterClient, id string, expiration time.Duration) error {

	val := fmt.Sprintf("%s%s", defaultRedisKeyPrefix, id)
	ok, err := r.Expire(ctx, val, expiration).Result()
	if err != nil || !ok {
		return err
	}

	return nil
}

func Unlock(r *redis.ClusterClient, id string) {

	val := fmt.Sprintf("%s%s", defaultRedisKeyPrefix, id)
	res, err := r.Unlink(ctx, val).Result()
	if err != nil || res != 1 {
		fmt.Println("Unlock res = ", res)
		fmt.Println("Unlock err = ", err)
	}
}
