package database

import (
	"log"
	"sync"
	"time"

	"github.com/lib/pq"
)

var callbacks = make([]func(), 0)

// this is just paranoia. this is only modified in init and only used after init, but might as well be safe and mutex it :)
var callbacksLock sync.Mutex

func CallbackOnUsersTableUpdate(callback func()) {
	callbacksLock.Lock()
	defer callbacksLock.Unlock()
	callbacks = append(callbacks, callback)
}

func fireCallbacks() {
	callbacksLock.Lock()
	defer callbacksLock.Unlock()
	log.Println("Postgres trigger 'users_updated' got pinged!")
	for _, callback := range callbacks {
		callback()
	}
}

func setupListener(url string) {
	minReconn := 10 * time.Second
	maxReconn := time.Minute
	listener := pq.NewListener(url, minReconn, maxReconn, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Println("WARNING: Postgres listener hit some kind of error!")
			log.Println(err)
		}
	})
	err := listener.Listen("users_updated")
	if err != nil {
		panic(err)
	}
	log.Println("Postgres listener created")
	go func() {
		for range listener.Notify {
			fireCallbacks()
		}
	}()
}
