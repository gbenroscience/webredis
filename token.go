package webredis

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gbenroscience/webredis/utils"
	"github.com/go-redis/redis/v8"
)

// RedisTokenStore Manages tokens usable with REST APIS and saves them to redis.
// Writes them to the specified header in the response automatically when the request has been processed
type RedisTokenStore struct {
	rcl *RedisStore
	//Encryption keys for token data
	keys string
	//applies to all sessions created in seconds, you may customize on the individual sessions
	// using session.Options.MaxAge = ...
	maxAgeDef int
}

// NewWebRedisStore Creates a pointer to a new RedisTokenStore
// redisClient: a client connection to redis
// secretKey: A 32 bytes long string to use for encrypting(using AES) and decryptng the session data
func NewRedisTokenStore(redisClient *redis.Client, secretKey string, defaultSessionAge int) *RedisTokenStore {
	return &RedisTokenStore{rcl: &RedisStore{Conn: redisClient}, keys: secretKey, maxAgeDef: defaultSessionAge}
}

type Session struct {
	// ID is used as the key for saving the encoded Session in redis
	ID string `json:"id"`
	// name is used to identify the kind of session
	Name   string                 `json:"name"`
	Values map[string]interface{} `json:"value"`
	IsNew  bool                   `json:"is_new"`
	MaxAge int                    `json:"max_age"`
}

func create(r *http.Request, name string, maxAge int) *Session {
	sess := new(Session)
	var rnd = utils.NewRnd()
	id := rnd.GenULID()
	id = base64.RawURLEncoding.EncodeToString([]byte(id))
	sess.ID = id
	sess.Name = name
	sess.Values = make(map[string]interface{})
	sess.MaxAge = maxAge
	sess.IsNew = true

	return sess
}

// GetExisting returns a Session if one exists
// r The http request
// name  name assigned to the session token, analogous, perhaps to the cookie name in web sessions
// This name is always included as an header name in your response and its value is set to be the value of the session id.
func (rts *RedisTokenStore) GetExisting(r *http.Request, name string) (*Session, error) {
	rs := rts.rcl

	if sessionID := r.Header.Get(name); len(sessionID) > 0 {

		var sessText string
		redisStat, err := rs.Get(sessionID, &sessText)

		if err != nil {
			return nil, err
		}
		if redisStat == RedisRecordFound {
			// The cached session was retrieved
			session, err := rts.fromToken(sessText)
			if err != nil {
				return nil, err
			}
			session.IsNew = false
			return session, nil
		} else {
			return nil, err
		}

	} else {
		return nil, errors.New("session does not exist")
	}
}

// Get returns a Session if one exists, or creates a new one if not
// r The http request
// name  name assigned to the session token, analogous, perhaps to the cookie name in web sessions
// This name is always included as an header name in your response and its value is set to be the value of the session id.
func (rts *RedisTokenStore) Get(r *http.Request, name string) (*Session, error) {
	rs := rts.rcl

	if sessionID := r.Header.Get(name); len(sessionID) > 0 {

		var sessText string
		redisStat, err := rs.Get(sessionID, &sessText)

		if err != nil {
			//redis may be running on a configuration where it does not save to disk when power is lost.
			// So give the user a new session here.
			session := create(r, name, rts.maxAgeDef)
			return session, nil
		}

		if redisStat == RedisRecordFound {
			// The cached session was retrieved
			session, err := rts.fromToken(sessText)
			if err != nil {
				//Data corruption occurred either with redis or the AES algorithm. Give a new session, please
				session = create(r, name, rts.maxAgeDef)
				return session, nil
			}
			session.IsNew = false
			return session, nil
		} else if redisStat == RedisRecordNotFound {
			//Session possibly has expired in redis; most likely
			session := create(r, name, rts.maxAgeDef)
			return session, nil
		} else {
			//Weird, weird, weird
			session := create(r, name, rts.maxAgeDef)
			return session, nil
		}

	} else {
		//Session header set, but with no value... programming error most likely
		//Most likely from registration or login, since no session header exists
		session := create(r, name, rts.maxAgeDef)
		return session, nil
	}

}

func (s *Session) StoreInt(key string, val int) {
	s.Values[key] = val
}
func (s *Session) StoreText(key string, val string) {
	s.Values[key] = val
}
func (s *Session) StoreBool(key string, val bool) {
	s.Values[key] = val
}
func (s *Session) StoreFloat32(key string, val float32) {
	s.Values[key] = val
}
func (s *Session) StoreFloat64(key string, val float64) {
	s.Values[key] = val
}
func (s *Session) StoreByte(key string, val []byte) {
	s.Values[key] = val
}
func (s *Session) StoreAny(key string, val interface{}) {
	s.Values[key] = val
}

func (s *Session) GetText(key string, defaultVal string) string {
	if txt, ok := s.Values[key].(string); ok {
		return txt
	}
	return defaultVal
}
func (s *Session) GetBoolean(key string, defaultVal bool) bool {
	if boole, ok := s.Values[key].(bool); ok {
		return boole
	}
	return defaultVal
}
func (s *Session) GetInt(key string, defaultVal int) int {
	if bits, ok := s.Values[key].(int); ok {
		return bits
	}
	return defaultVal
}
func (s *Session) GetByte(key string, defaultVal byte) byte {
	if bits, ok := s.Values[key].(byte); ok {
		return bits
	}
	return defaultVal
}
func (s *Session) GetFloat32(key string, defaultVal float32) float32 {
	if bits, ok := s.Values[key].(float32); ok {
		return bits
	}
	return defaultVal
}
func (s *Session) GetFloat64(key string, defaultVal float64) float64 {
	if bits, ok := s.Values[key].(float64); ok {
		return bits
	}
	return defaultVal
}

func (s *Session) GetAny(key string) interface{} {
	return s.Values[key]
}

// DeleteAny You need to call RedisTokenStore.Save to persist this action to redis!
func (s *Session) DeleteAny(key string) {
	delete(s.Values, key)
}

// token generate the encrypted string sent to the browser and stored in Redis
func (rts *RedisTokenStore) token(s *Session) (string, error) {
	jsn := utils.Stringify(s)

	k, err := utils.NewKryptik(rts.keys, utils.ModeCBC)
	if err != nil {
		return "", err
	}
	return k.Encrypt(jsn)
}

// Token regenerate the oiginal Session from its token
func (rts *RedisTokenStore) fromToken(sessionToken string) (*Session, error) {
	k, err := utils.NewKryptik(rts.keys, utils.ModeCBC)
	if err != nil {
		return nil, err
	}
	jsn, err := k.Decrypt(sessionToken)
	if err != nil {
		return nil, err
	}
	var s Session
	err = json.NewDecoder(bytes.NewBufferString(jsn)).Decode(&s)
	return &s, err
}

// Save saves a session in redis
func (rts *RedisTokenStore) Save(s *Session, r *http.Request, w http.ResponseWriter) error {

	tkn, err := rts.token(s)

	if err != nil {
		return err
	}
	redisStat, err := rts.rcl.SetWithExpiry(s.ID, tkn, int64(s.MaxAge)) // save session to redis
	if redisStat == RedisRecordUpdated {
		w.Header().Set(s.Name, s.ID)
	}
	return err
}

// Delete Manually delete the session from redis
func (rts *RedisTokenStore) Delete(s *Session) (int64, error) {
	rs := rts.rcl
	return rs.Delete(s.ID)
}

// Close the redis connection once you are done
func (rts *RedisTokenStore) Close() error {
	return rts.rcl.Close()
}
