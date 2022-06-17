package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"
)

// RandomLife ...
type RandomLife struct {
	SeededRand *rand.Rand
}

// Letters of the alphabet in upper and lower case
const (
	ALPHABET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	DIGITS   = "0123456789"
)

// NewRnd ...
func NewRnd() RandomLife {
	return RandomLife{
		SeededRand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

//NextInt - Generates a number between 0 and max, max. excluded
func (rnd *RandomLife) NextInt(max int) int {

	var mu sync.Mutex

	// lock/unlock when accessing the rand from a goroutine
	mu.Lock()
	i := 0 + rnd.SeededRand.Intn(max)
	mu.Unlock()

	return i
}

//NextFloat - Generates a number between 0 and 1
func (rnd *RandomLife) NextFloat() float64 {

	var mu sync.Mutex

	// lock/unlock when accessing the rand from a goroutine
	mu.Lock()
	i := 0 + rnd.SeededRand.Float64()
	mu.Unlock()

	return i

}

//GenULID - Generates a ULID
func (rnd *RandomLife) GenULID() string {
	return rnd.genUlid()
}

func (rnd *RandomLife) genUlid() string {
	t := time.Now().UTC()

	var mu sync.Mutex

	// lock/unlock when accessing the rand from a goroutine
	mu.Lock()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	mu.Unlock()

	id := ulid.MustNew(ulid.Timestamp(t), entropy)

	return id.String()
}

// GenerateRndFloat ...Supply min and max
func (rnd *RandomLife) GenerateRndFloat(min float32, max float32) float32 {
	return min + rnd.SeededRand.Float32()*(max-min)
}

// CurrentTimeStamp  The time now
func CurrentTimeStamp() int {
	return int(time.Now().UnixMilli())
}

func DumpStruct(b interface{}) {
	s, _ := json.MarshalIndent(b, "", "\t")
	fmt.Println(string(s))
}

func GetObjectBytes(b interface{}) []byte {
	s, _ := json.Marshal(b)
	return s
}

// Stringify Gets the json format of a struct and returns it without indentation.
func Stringify(b interface{}) string {
	s, _ := json.Marshal(b)
	return string(s)
}

// StringifyObject Gets the json format of a struct and returns it with indentation.
func StringifyObject(b interface{}) string {
	s, _ := json.MarshalIndent(b, "", "\t")
	return string(s)
}

// DecodeItem Decodes a json string into a pointer to a generic Golang struct. Pass a pointer to this function
func DecodeItem(jsn string, destPtr interface{}) error {
	return json.NewDecoder(bytes.NewBufferString(jsn)).Decode(destPtr)
}

// DecodeBytes Decodes json bytes into a pointer to a generic Golang struct. Pass a pointer to this function
func DecodeBytes(jsn []byte, destPtr interface{}) error {
	return json.NewDecoder(bytes.NewBuffer(jsn)).Decode(destPtr)
}
