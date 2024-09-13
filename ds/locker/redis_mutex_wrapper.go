package locker

import (
	"github.com/go-redsync/redsync/v4"
)

type RedSyncMutexWrapper redsync.Mutex

func (l *RedSyncMutexWrapper) Lock() {
	(*redsync.Mutex)(l).Lock()
}

func (l *RedSyncMutexWrapper) Unlock() {
	(*redsync.Mutex)(l).Unlock()
}
