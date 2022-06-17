package webredis

import (
	"net/http"
)

type GenericSession interface {
	StoreInt(key string, val int)
	StoreText(key string, val string)
	StoreBool(key string, val bool)
	StoreFloat32(key string, val float32)
	StoreFloat64(key string, val float64)
	StoreByte(key string, val byte)
	StoreBytes(key string, val []byte)
	StoreAny(key string, val interface{})

	GetText(key string, defaultVal string) string
	GetBoolean(key string, defaultVal bool) bool
	GetInt(key string, defaultVal int) int
	GetByte(key string, defaultVal byte) byte
	GetBytes(key string, defaultVal []byte) []byte
	GetFloat32(key string, defaultVal float32) float32
	GetFloat64(key string, defaultVal float64) float64

	GetAny(key string) interface{}

	// DeleteAny You need to call RedisSessionStore.Save to persist this action to redis!
	DeleteAny(key string)
}

type GenericStore interface {
	Get(r *http.Request, name string) (*GenericSession, error)
	GetExisting(r *http.Request, name string) (*GenericSession, error)
	Save(s *GenericSession, r *http.Request, w http.ResponseWriter) error
	Delete(s *GenericSession) (int64, error)
	Close() error
}
