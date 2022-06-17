package sessions

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gbenroscience/webredis/utils"

	"github.com/gbenroscience/webredis"
	"github.com/go-redis/redis/v8"
)

type RedisSessionStore struct {
	rcl *webredis.RedisStore
	//Encryption keys for session data
	keys string
	//applies to all sessions created in seconds, you may customize on the individual sessions
	// using session.Options.MaxAge = ...
	maxAgeDefault int
}

type Options struct {
	Path   string `json:"path"`
	Domain string `json:"domain"`
	// MaxAge=0 means no Max-Age attribute specified and the cookie will be
	// deleted after the browser session ends.
	// MaxAge<0 means delete cookie immediately.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int  `json:"max_age"`
	Secure   bool `json:"secure"`
	HttpOnly bool `json:"http_only"`
	// Defaults to http.SameSiteDefaultMode
	SameSite http.SameSite
}

type Session struct {
	// ID is used as the key for saving the encoded Session in redis
	ID string `json:"id"`
	// name is used to identify the kind of session
	Name    string                 `json:"name"`
	Values  map[string]interface{} `json:"value"`
	IsNew   bool                   `json:"is_new"`
	Options *Options               `json:"options"`
}

// NewWebRedisStore Creates a pointer to a new RedisSessionStore
// redisClient: a client connection to redis
// secretKey: A 32 bytes long string to use for encrypting(using AES) and decryptng the session data
// defaultSessionAge: The age to apply to all sessions by default.It may be changed per session later
func NewWebRedisStore(redisClient *redis.Client, secretKey string, defaultSessionAge int) *RedisSessionStore {
	return &RedisSessionStore{rcl: &webredis.RedisStore{Conn: redisClient}, keys: secretKey, maxAgeDefault: defaultSessionAge}
}

// GetExisting returns a Session if one exists
func (rss *RedisSessionStore) GetExisting(sessionID string) (*Session, error) {
	session := new(Session)
	rs := rss.rcl

	var sessText string
	redisStat, err := rs.Get(sessionID, &sessText)

	if err != nil {
		return nil, err
	}

	session, err = rss.fromToken(sessText)

	if redisStat == webredis.RedisRecordFound {
		session.IsNew = false
		return session, nil
	} else {
		return nil, err
	}
}

// Get returns a Session if one exists, or creates a new one if not
func (rss *RedisSessionStore) Get(r *http.Request, name string) (*Session, error) {
	session := new(Session)
	rs := rss.rcl

	if c, err := r.Cookie(name); err == nil {
		sessionID := c.Value
		if len(sessionID) > 0 {
			var sessText string
			redisStat, err := rs.Get(sessionID, &sessText)

			if err != nil {
				//redis may be running on a configuration where it does not save to disk when power is lost.
				// So give the user a new session here.
				session = create(r, name, rss.maxAgeDefault)
				return session, nil
			}

			if redisStat == webredis.RedisRecordFound {
				// The cached session was retrieved
				session, err = rss.fromToken(sessText)
				if err != nil {
					//Data corruption occurred either with redis or the AES algorithm. Give a new session, please
					session = create(r, name, rss.maxAgeDefault)
					return session, nil
				}
				session.IsNew = false
				return session, nil
			} else if redisStat == webredis.RedisRecordNotFound {
				//Session possibly has expired in redis; most likely
				session = create(r, name, rss.maxAgeDefault)
				return session, nil
			} else {
				//Weird, weird, weird
				session = create(r, name, rss.maxAgeDefault)
				return session, nil
			}
		} else {
			//Session cookie set, but with no value... programming error most likely
			//Most likely from registration or login, since no session header exists
			session = create(r, name, rss.maxAgeDefault)
			return session, nil
		}

	} else {
		//Session cookie not set
		//Most likely from registration or login, since no session header exists
		session = create(r, name, rss.maxAgeDefault)
		return session, nil
	}

}

func create(r *http.Request, name string, maxAge int) *Session {
	sess := new(Session)
	var rnd = utils.NewRnd()
	id := rnd.GenULID()
	id = base64.RawURLEncoding.EncodeToString([]byte(id))
	sess.ID = id
	sess.Name = name
	sess.Values = make(map[string]interface{})
	sess.Options = new(Options)
	sess.Options.Domain = r.RemoteAddr
	sess.Options.Path = "/"
	sess.Options.HttpOnly = false
	sess.Options.MaxAge = maxAge
	sess.Options.SameSite = 1
	sess.IsNew = true

	return sess
}

// newCookieFromOptions returns an http.Cookie with the options set.
func newCookieFromOptions(name, value string, options *Options) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: options.SameSite,
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
func (s *Session) StoreByte(key string, val byte) {
	s.Values[key] = val
}
func (s *Session) StoreBytes(key string, val []byte) {
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
func (s *Session) GetBytes(key string, defaultVal []byte) []byte {
	if bits, ok := s.Values[key].([]byte); ok {
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

// DeleteAny You need to call RedisSessionStore.Save to persist this action to redis!
func (s *Session) DeleteAny(key string) {
	delete(s.Values, key)
}

// NewCookie returns an http.Cookie with the options set. It also sets
// the Expires field calculated based on the MaxAge value, for Internet
// Explorer compatibility.
func NewCookie(name, value string, options *Options) *http.Cookie {
	cookie := newCookieFromOptions(name, value, options)
	if options.MaxAge > 0 {
		cookie.MaxAge = options.MaxAge
		d := time.Duration(options.MaxAge) * time.Second
		cookie.Expires = time.Now().Add(d)
	} else if options.MaxAge < 0 {
		// Set it to the past to expire now.
		cookie.Expires = time.Unix(1, 0)
	}
	return cookie
}

// token generate the encrypted string sent to the browser and stored in Redis
func (rss *RedisSessionStore) token(s *Session) (string, error) {
	jsn := utils.Stringify(s)

	k, err := utils.NewKryptik(rss.keys, utils.ModeCBC)
	if err != nil {
		return "", err
	}
	return k.Encrypt(jsn)
}

// Token regenerate the oiginal Session from its token
func (rss *RedisSessionStore) fromToken(sessionToken string) (*Session, error) {
	k, err := utils.NewKryptik(rss.keys, utils.ModeCBC)
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
func (rss *RedisSessionStore) Save(s *Session, r *http.Request, w http.ResponseWriter) error {

	tkn, err := rss.token(s)

	if err != nil {
		return err
	}
	redisStat, err := rss.rcl.SetWithExpiry(s.ID, tkn, int64(s.Options.MaxAge)) // save session to redis
	if redisStat == webredis.RedisRecordUpdated {
		http.SetCookie(w, NewCookie(s.Name, s.ID, s.Options)) // send session id to browser as cookie
	}
	return err
}

// Delete Manually delete the session from redis
func (rss *RedisSessionStore) Delete(s *Session) (int64, error) {
	rs := rss.rcl
	return rs.Delete(s.ID)
}

// Close the redis connection once you are done
func (rts *RedisSessionStore) Close() error {
	return rts.rcl.Close()
}
