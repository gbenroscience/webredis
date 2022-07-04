package datastore

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gbenroscience/webredis"
	"github.com/gbenroscience/webredis/utils"
	"github.com/go-redis/redis/v8"
)

// RedisItemStore Manages general data storage that is not tied to any request/response or session using redis.
// It uses a key - value system to store data in redis
type RedisItemStore struct {
	rcl *webredis.RedisStore
	//Encryption keys for token data
	keys string
	//applies to all sessions created in seconds, you may customize on the individual sessions
	// using session.Options.MaxAge = ...
	maxAgeDef int
}

// NewRedisItemStore Creates a pointer to a new RedisItemStore
// redisClient: a client connection to redis
// secretKey: A 32 byte long string to use for encrypting(using AES) and decrypting the session data
func NewRedisItemStore(redisClient *redis.Client, secretKey string, defaultSessionAge int) *RedisItemStore {
	return &RedisItemStore{rcl: &webredis.RedisStore{Conn: redisClient}, keys: secretKey, maxAgeDef: defaultSessionAge}
}

type RedisItem struct {
	// ID is used as the key for saving the encoded Session in redis
	ID     string                 `json:"id"`
	Values map[string]interface{} `json:"value"`
	maxAge int
}

func create(id string, maxAge int) *RedisItem {
	sess := new(RedisItem)
	sess.ID = id
	sess.Values = make(map[string]interface{})
	sess.maxAge = maxAge
	return sess
}

// GetExisting returns a Session if one exists
// r The http request
// id  The id assigned to the RedisItem, analogous
// This name is always included as an header name in your response and its value is set to be the value of the session id.
func (rts *RedisItemStore) GetExisting(id string) (*RedisItem, error) {
	rs := rts.rcl

	var sessText string
	redisStat, err := rs.Get(id, &sessText)
	if err != nil {
		return nil, err
	}
	if redisStat == webredis.RedisRecordFound {
		// The cached session was retrieved
		item, err := fromToken(sessText, rts.keys)
		if err != nil {
			return nil, err
		}
		return item, nil
	} else {
		return nil, err
	}

}

// Get returns a Session if one exists, or creates a new one if not
// r The http request
// id Used to retrieve the RedisItem
func (rts *RedisItemStore) Get(id string, maxAgeHack ...int) (*RedisItem, error) {
	rs := rts.rcl
	maxAge := rts.maxAgeDef
	ln := len(maxAgeHack)
	if ln == 1 {
		maxAge = maxAgeHack[0]
	} else if ln > 1 {
		return nil, errors.New("`maxAgeHack` can only be empty or have one item which is the maximum age`")
	}
	var sessText string
	redisStat, err := rs.Get(id, &sessText)

	if err != nil {
		//redis may be running on a configuration where it does not save to disk when power is lost.
		// So give the user a new item here.
		item := create(id, maxAge)
		return item, nil
	}

	if redisStat == webredis.RedisRecordFound {
		// The cached session was retrieved
		item, err := fromToken(sessText, rts.keys)
		if err != nil {
			//Data corruption occurred either with redis or the AES algorithm. Give a new session, please
			item = create(id, rts.maxAgeDef)
			return item, nil
		}
		return item, nil
	} else if redisStat == webredis.RedisRecordNotFound {
		//Session possibly has expired in redis; most likely
		item := create(id, rts.maxAgeDef)
		return item, nil
	} else {
		//Weird, weird, weird
		item := create(id, rts.maxAgeDef)
		return item, nil
	}

}

func (s *RedisItem) StoreInt(key string, val int) {
	s.Values[key] = val
}
func (s *RedisItem) StoreText(key string, val string) {
	s.Values[key] = val
}
func (s *RedisItem) StoreBool(key string, val bool) {
	s.Values[key] = val
}
func (s *RedisItem) StoreFloat32(key string, val float32) {
	s.Values[key] = val
}
func (s *RedisItem) StoreFloat64(key string, val float64) {
	s.Values[key] = val
}
func (s *RedisItem) StoreByte(key string, val byte) {
	s.Values[key] = val
}
func (s *RedisItem) StoreBytes(key string, val []byte) {
	s.Values[key] = val
}
func (s *RedisItem) StoreAny(key string, val interface{}) {
	s.Values[key] = val
}

func (s *RedisItem) GetText(key string, defaultVal string) string {
	if txt, ok := s.Values[key].(string); ok {
		return txt
	}
	return defaultVal
}
func (s *RedisItem) GetBoolean(key string, defaultVal bool) bool {
	if boole, ok := s.Values[key].(bool); ok {
		return boole
	}
	return defaultVal
}
func (s *RedisItem) GetInt(key string, defaultVal int) int {
	if bits, ok := s.Values[key].(int); ok {
		return bits
	}
	return defaultVal
}
func (s *RedisItem) GetByte(key string, defaultVal byte) byte {
	if bits, ok := s.Values[key].(byte); ok {
		return bits
	}
	return defaultVal
}
func (s *RedisItem) GetBytes(key string, defaultVal []byte) []byte {
	if bits, ok := s.Values[key].([]byte); ok {
		return bits
	}
	return defaultVal
}
func (s *RedisItem) GetFloat32(key string, defaultVal float32) float32 {
	if bits, ok := s.Values[key].(float32); ok {
		return bits
	}
	return defaultVal
}
func (s *RedisItem) GetFloat64(key string, defaultVal float64) float64 {
	if bits, ok := s.Values[key].(float64); ok {
		return bits
	}
	return defaultVal
}

func (s *RedisItem) GetAny(key string) interface{} {
	return s.Values[key]
}

// DeleteAny You need to call RedisItemStore.Save to persist this action to redis!
func (s *RedisItem) DeleteAny(key string) {
	delete(s.Values, key)
}

// token generate the encrypted string sent to the browser and stored in Redis
func token(s *RedisItem, encryptionKeys string) (string, error) {
	jsn := utils.Stringify(s)

	k, err := utils.NewKryptik(encryptionKeys, utils.ModeCBC)
	if err != nil {
		return "", err
	}
	return k.Encrypt(jsn)
}

// Token regenerate the original Session from its token
func fromToken(sessionToken string, encryptionKeys string) (*RedisItem, error) {
	k, err := utils.NewKryptik(encryptionKeys, utils.ModeCBC)
	if err != nil {
		return nil, err
	}
	jsn, err := k.Decrypt(sessionToken)
	if err != nil {
		return nil, err
	}
	var item RedisItem
	err = json.NewDecoder(bytes.NewBufferString(jsn)).Decode(&item)
	return &item, err
}

// Save saves a session in redis
func (rts *RedisItemStore) Save(s *RedisItem) error {

	tkn, err := token(s, rts.keys)

	if err != nil {
		return err
	}
	_, err = rts.rcl.SetWithExpiry(s.ID, tkn, int64(s.maxAge)) // save session to redis

	return err
}


// Delete Manually delete the session from redis
func (rts *RedisItemStore) Delete(s *RedisItem) (int64, error) {
	rs := rts.rcl
	return rs.Delete(s.ID)
}

// Close the redis connection once you are done
func (rts *RedisItemStore) Close() error {
	return rts.rcl.Close()
}
