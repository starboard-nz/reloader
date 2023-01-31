package reloader

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Loader[T any] struct {
	lock   sync.RWMutex
	reload bool // data should be reloaded
	valid  bool // data has been loaded
	data   *T
	deltaT time.Duration
	loader func() (*T, error)
}

func NewLoader[T any](loader func() (*T, error), deltaT time.Duration) *Loader[T] {
	l := &Loader[T]{
		deltaT: deltaT,
		loader: loader,
	}

	if deltaT != 0 {
		l.autoReset()
	}

	return l
}

func (l *Loader[T]) autoReset() {
	go func() {
		c := time.Tick(l.deltaT)
		for range c {
			l.Reset()
		}
	}()	
}

// Load loads the data initially, and then reloads it if Reset() was called
func (l *Loader[T]) Load() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.valid && !l.reload {
		// another goroutine has already loaded the data
		return nil
	}

	d, err := l.loader()
	if err == nil {
		l.data = d
		l.reload = false
		l.valid = true
	}

	return err
}

// Reset causes the data to be reloaded next time ensureLoaded() is called
func (l *Loader[T]) Reset() {
	l.lock.Lock()
	l.reload = true
	l.lock.Unlock()
}

// Do runs a function that can read the protected data.
// The function must not modify the data.
func (l *Loader[T]) Do(f func(*T) error) error {
	err := l.ensureLoaded()
	l.lock.RLock()
	if err != nil {
		log.Error().Msgf("reloader: reload failed: %v", err)

		if !l.valid {
			l.lock.RUnlock()
			return err
		}

		log.Warn().Msgf("reloader: using previously loaded data")
	}

	err = f(l.data)
	l.lock.RUnlock()

	return err
}

// ensureLoaded is a function called by functions that need access to the protected / cached data.
// It's like sync.Once but reloads the data if Reset() was called.
func (l *Loader[T]) ensureLoaded() error {
	l.lock.RLock()
	if !l.valid || l.reload {
		l.lock.RUnlock()
		return l.Load()
	}

	l.lock.RUnlock()

	return nil
}

