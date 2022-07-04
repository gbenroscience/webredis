package webredis

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisDB struct {
	*redis.Client
}

type RedisStore struct {
	Conn *redis.Client
}

func (rds *RedisStore) SetWithExpiry(key string, value interface{}, expiryDuration int64) (int, error) {
	p, err := json.Marshal(value)
	if err != nil {
		return RedisMarshalUpdateError, err
	}

	err = rds.Conn.Set(context.Background(), key, p, time.Duration(expiryDuration)*time.Second).Err()

	if err == nil {
		return RedisRecordUpdated, nil
	} else {
		return RedisRecordUpdateError, err
	}
}

func (rds *RedisStore) Set(key string, value interface{}) (int, error) {
	p, err := json.Marshal(value)
	if err != nil {
		return RedisMarshalUpdateError, err
	}

	err = rds.Conn.Set(context.Background(), key, p, 0*time.Second).Err()

	if err == nil {
		return RedisRecordUpdated, nil
	} else {
		return RedisRecordUpdateError, err
	}
}

func isPointer(i interface{}) bool {
	return reflect.ValueOf(i).Kind() == reflect.Ptr
}

// AddToSet fetches a set (or creates it if it does not already exist) identified
// by the `nameOfSet`. Then it adds the value to it
func (rds *RedisStore) AddToSet(nameOfSet string, value string) (int, error) {
	res := rds.Conn.SAdd(context.Background(), nameOfSet, value)

	err := res.Err()
	if err != nil {
		return RedisRecordUpdateError, err
	} else {
		return RedisRecordUpdated, nil
	}
}

// IsInSet Checks if a value exists in a set called `nameOfSet`. returns
// RedisRecordFound,nil if found and RedisRecordNotFound,nil If not found.
// Returns RedisRecordFetchError, err if an error occurred
func (rds *RedisStore) IsInSet(nameOfSet string, value string) (int, error) {
	res := rds.Conn.SIsMember(context.Background(), nameOfSet, value)

	err := res.Err()
	if err != nil {
		return RedisRecordFetchError, err
	}
	if res.Val() {
		return RedisRecordFound, nil
	} else {
		return RedisRecordNotFound, nil
	}
}

// DeleteFromSet Removes an item from the set. If the item does not exist in the set, it returns false and nil
// If it does, it deletes it and returns true and nil. If an error occurred while doing all this, it returns false and the error
func (rds *RedisStore) DeleteFromSet(nameOfSet, value string) (bool, error) {
	res := rds.Conn.SRem(context.Background(), nameOfSet, value)

	err := res.Err()
	val := res.Val()
	if err != nil {
		return false, err
	} else {
		if val == RedisRecordFound {
			return true, nil
		} else {
			return false, nil
		}
	}
}

// Get ...
// key is the name of the key whose value we wish to retrieve,
// dest .. is a pointer to the interface that we wish to decode the value into.
func (rds *RedisStore) Get(key string, dest interface{}) (int, error) {

	if !isPointer(dest) {
		return RedisInvalidArgsError, errors.New("the `dest` parameter can only be a pointer")
	}

	p, err := rds.Conn.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return RedisRecordNotFound, err
	} else if err != nil {
		return RedisRecordFetchError, err
	} else {
		err = json.Unmarshal([]byte(p), dest)

		if err == nil {
			return RedisRecordFound, nil
		} else {
			return RedisRecordUnmarshalError, err
		}

	}
}

func (rds *RedisStore) Delete(key string) (int64, error) {
	return rds.Conn.Pipeline().Del(context.Background(), key).Result()
}

func (rds *RedisStore) Close() error {
	return rds.Conn.Close()
}
