package middleware

import (
	"sync"
	"time"
)

type blockedToken struct {
	expiredAt time.Time
}

var (
	jwtBlocklist = make(map[string]blockedToken)
	mutex        sync.RWMutex
)

func BlockJWT(token string, expiredAt time.Time) {
	mutex.Lock()
	defer mutex.Unlock()

	jwtBlocklist[token] = blockedToken{
		expiredAt: expiredAt,
	}
}

func IsJWTBlocked(token string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	data, exists := jwtBlocklist[token]
	if !exists {
		return false
	}

	// auto clean kalau sudah expired
	if time.Now().After(data.expiredAt) {
		delete(jwtBlocklist, token)
		return false
	}

	return true
}
