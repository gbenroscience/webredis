# webredis
An http session manager backed by redis. Provides a sessionmanager for non-web apis also

This is a session store backed by redis; allowing you to save your web sessions in redis instead of in file systems etc.
We have used [go-redis](https://github.com/go-redis/redis) as the client implementation for **redis** access



This library contains 2 session stores.

They are:

1. ```RedisSessionStore```
2. ```RedisTokenStore```

### RedisSessionStore
Is defined as:
```Go
type RedisSessionStore struct {
	rcl *webredis.RedisStore
	//Encryption keys for session data
	keys string
	//applies to all sessions created in seconds, you may customize on the individual sessions
	// using session.Options.MaxAge = ...
	maxAgeDefault int
}
```
This session store is backed by redis and is used for managing web sessions. 
So your websites can have their sessions in one place; i.e redis, wthout needing your sessions to be sticky.

### RedisTokenStore
Is defined as:

```Go
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
```
This session store can be used by generic application backends to save data which needs be quickly accessed; e.g. application state etc.

## Usage

To create a session store for web, do:

```Go
webSessionStore := sessions.NewWebRedisStore(client, "32-byte-key-for-session-encoding", 7200)
```

To create a session store for a generic backend, do:

```Go
redisTokenStore := webredis.NewRedisTokenStore(client, "32-byte-key-for-session-encoding", 7200)
```

The ```client``` parameter is a ```*redis.Client```
The 2nd parameter is a 32 byte key used for encrypting the sessions
The 3rd parameter is the time in seconds that represents how long the session will live before it is expired by redis.

To create a session using the web session store, do:

```Go
sess, err := webSessionStore.Get(request, "user")
```
This returns a ```sessions.Session``` object

and using the generic store:

```Go
sess, err := redisTokenStore.Get(request, "user")
```
This returns a ```webredis.Session``` object


In both cases, if the session does not exist, a new one will be created. You may check if a new session was generated for you by using ```sess.IsNew```
Once created, you may begin to save user data on the session using any of the following:

```Go
sess.StoreInt("number-here", 223)
sess.StoreText("text-here", "User name")
sess.StoreBool("boolean-here", true)
sess.StoreFloat32("float32-key", 3.143)
sess.StoreFloat64("float64-key", 3.143)
sess.StoreByte("byte-key", []byte{1, 2, 3, 101})

type Boy struct{
Age int
Color string
Height float32
}
boy := new(Boy)
boy.Age = 12
boy.Color = "brown"
boy.Height = 1.51
sess.StoreAny("anykey-goes", boy)
```

Both the webredis.Session and the sessions.Session have these functions

When closing your server application, remember to call ```webSessionStore.Close()``` or ```redisTokenStore.Close()```
This will close the connections to ```redis```














