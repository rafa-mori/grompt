package types

import (
	gl "github.com/rafa-mori/grompt/internal/module/logger"

	"sync"
	"time"
)

// IMutexes is an interface that defines the methods for a mutex context map.
type IMutexes interface {
	MuLock()
	MuUnlock()
	MuRLock()
	MuRUnlock()
	MuTryLock() bool
	MuTryRLock() bool

	MuWaitCond()
	MuSignalCond()
	MuBroadcastCond()

	GetMuSharedCtx() any
	SetMuSharedCtx(ctx any)
	GetMuSharedCtxValidate() func(any) (bool, error)
	SetMuSharedCtxValidate(validate func(any) (bool, error))
	MuWaitCondWithTimeout(timeout time.Duration) bool

	MuAdd(delta int)
	MuDone()
	MuWait()
}

// muCtx is the mutex context map
type muCtx struct {
	// MuCtxM is a mutex for the ctx map.
	MuCtxM *sync.RWMutex
	// MuCtxL is a mutex for sync.Cond in the ctx map.
	MuCtxL *sync.RWMutex
	// MuCtxCond is a condition variable for the ctx map.
	MuCtxCond *sync.Cond
	// MuCtxWg is a wait group for the ctx map.
	MuCtxWg *sync.WaitGroup
}

// Mutexes is a struct that holds the mutex context map
type Mutexes struct {
	// muCtx is the mutex context map
	*muCtx

	// MuCtxM is a mutex for the ctx map.
	MuCtxM *sync.RWMutex
	// MuCtxL is a mutex for sync.Cond in the ctx map.
	MuCtxL *sync.RWMutex
	// MuCtxCond is a condition variable for the ctx map.
	MuCtxCond *sync.Cond
	// MuCtxWg is a wait group for the ctx map.
	MuCtxWg *sync.WaitGroup

	// muSharedM is a mutex for the shared context.
	muSharedM *sync.RWMutex
	// muSharedCtx is the shared context for Cond. This is used to synchronize states across multiple goroutines.
	muSharedCtx any
	// muSharedCtxValidate is the shared context validation function. This is used to validate the shared context defining if it needs to wait or not.
	muSharedCtxValidate func(any) (bool, error)
}

// newMuCtx creates a new mutex context map
func newMuCtx(mSharedCtxM *sync.RWMutex) *muCtx {
	mu := &muCtx{
		MuCtxM:    &sync.RWMutex{},
		MuCtxCond: sync.NewCond(mSharedCtxM),
		MuCtxWg:   &sync.WaitGroup{},
	}
	return mu
}

// NewMutexesType creates a new mutex context map struct pointer.
func NewMutexesType() *Mutexes {
	mu := &Mutexes{
		MuCtxM:              &sync.RWMutex{},
		MuCtxL:              &sync.RWMutex{},
		MuCtxWg:             &sync.WaitGroup{},
		muSharedM:           &sync.RWMutex{},
		muSharedCtx:         nil,
		muSharedCtxValidate: nil,
	}
	mu.muCtx = newMuCtx(mu.muSharedM)
	mu.MuCtxCond = sync.NewCond(mu.muSharedM)
	return mu
}

// NewMutexes creates a new mutex context map interface.
func NewMutexes() IMutexes { return NewMutexesType() }

// MuLock locks the mutex
func (m *Mutexes) MuLock() { m.MuCtxM.Lock() }

// MuUnlock unlocks the mutex
func (m *Mutexes) MuUnlock() { m.MuCtxM.Unlock() }

// MuRLock locks the mutex for reading
func (m *Mutexes) MuRLock() { m.MuCtxL.RLock() }

// MuRUnlock unlocks the mutex for reading
func (m *Mutexes) MuRUnlock() { m.MuCtxL.RUnlock() }

// GetMuSharedCtx returns the shared context
func (m *Mutexes) GetMuSharedCtx() any {
	m.muSharedM.RLock()
	defer m.muSharedM.RUnlock()

	return m.muSharedCtx
}

// SetMuSharedCtx sets the shared context
func (m *Mutexes) SetMuSharedCtx(ctx any) {
	m.muSharedM.Lock()
	defer m.muSharedM.Unlock()

	m.muSharedCtx = ctx
}

// GetMuSharedCtxValidate returns the shared context validation function
func (m *Mutexes) GetMuSharedCtxValidate() func(any) (bool, error) {
	m.muSharedM.RLock()
	defer m.muSharedM.RUnlock()

	return m.muSharedCtxValidate
}

// SetMuSharedCtxValidate sets the shared context validation function
func (m *Mutexes) SetMuSharedCtxValidate(validate func(any) (bool, error)) {
	m.muSharedM.Lock()
	defer m.muSharedM.Unlock()

	m.muSharedCtxValidate = validate
}

// MuWaitCondWithTimeout waits for the condition variable to be signaled with a timeout
func (m *Mutexes) MuWaitCondWithTimeout(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ch := make(chan struct{})
	go func() {
		m.MuCtxCond.Wait()
		close(ch)
	}()

	select {
	case <-ch:
		return true
	case <-timer.C:
		return false
	}
}

// MuWaitCond waits for the condition variable to be signaled
func (m *Mutexes) MuWaitCond() {

	m.MuCtxCond.Wait()
}

// MuSignalCond signals the condition variable
func (m *Mutexes) MuSignalCond() {
	m.muSharedM.Lock()
	defer m.muSharedM.Unlock()

	if m.muSharedCtxValidate != nil {
		isValid, err := m.muSharedCtxValidate(m.muSharedCtx)
		if err != nil || !isValid {
			gl.LogObjLogger(m, "warn", "Condition signal aborted due to validation failure")
			return
		}
	}

	gl.LogObjLogger(m, "info", "Signaling condition variable")
	m.MuCtxCond.Signal()
}

// MuBroadcastCond broadcasts the condition variable
func (m *Mutexes) MuBroadcastCond() {
	m.MuCtxCond.Broadcast()
}

// MuAdd adds a delta to the wait group counter
func (m *Mutexes) MuAdd(delta int) { m.MuCtxWg.Add(delta) }

// MuDone signals that the wait group is done
func (m *Mutexes) MuDone() { m.MuCtxWg.Done() }

// MuWait waits for the wait group counter to reach zero
func (m *Mutexes) MuWait() { m.MuCtxWg.Wait() }

// MuTryLock tries to lock the mutex without blocking
func (m *Mutexes) MuTryLock() bool {
	return m.MuCtxM.TryLock()
}

// MuTryRLock tries to lock the mutex for reading without blocking
func (m *Mutexes) MuTryRLock() bool {
	return m.MuCtxL.TryRLock()
}
