package webredis

const (
	// A record was found in redis
	RedisRecordFound = 1
	// When a value was not found in redis
	RedisRecordNotFound = 2
	// When an error occurs while trying to Get a value from redis
	RedisRecordFetchError = 3

	// When a record was successfully updated in redis
	RedisRecordUpdated = 4

	// An error occurred while updating a value in redis
	RedisMarshalUpdateError = 5
	// An error occurred while unmarshalling a struct to save in redis
	RedisRecordUnmarshalError = 6
	//An error occurred while updating a record in redis.
	RedisRecordUpdateError = 7

	// The user supplied an invalid interface to decode the redis record into
	RedisInvalidArgsError = 8
)
