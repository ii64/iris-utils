package longpoll
import (
	"log"
	"errors"
	"sync/atomic"
	"time"
)
var DEBUG = false
type Manager struct {
	opt Options
	events map[string]*lpEvent
}
func time2EMS(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
func (self *Manager) Status(eventName string) (r *lpEvent, isExist bool) {
	r, isExist = self.events[eventName]
	return
}
func (self *Manager) Subscribe(eventName string) (data interface{}, isInvalid bool) {
	event, ok := self.events[eventName]
	if !ok {
		event = &lpEvent{[]int64{time2EMS(time.Now())},uint32(1), make(chan interface{})}
		self.events[eventName] = event
	}else{
		(*event).timestamp = append((*event).timestamp, time2EMS(time.Now()))
		atomic.AddUint32(&event.size, 1)
	}
	if atomic.LoadUint32(&event.size) >= self.opt.MaxConnection && self.opt.MaxConnection != UNLIMITED_CONN {
		_ = atomic.SwapUint32(&event.size, atomic.LoadUint32(&event.size) - 1)
		isInvalid = true
		return
	}
	if self.opt.TimeoutSeconds == UNSET {
		data = <-event.data
		return
	}else{
		select{
		case <-time.After(time.Duration(self.opt.TimeoutSeconds)*time.Second):
			_ = atomic.SwapUint32(&event.size, atomic.LoadUint32(&event.size) - 1)
			isInvalid = true
			return
		case data = <-event.data:
			return
		}
	}
}
func (self *Manager) Publish(eventName string, data interface{}) (isError bool) {
	event, ok := self.events[eventName]
	if !ok {
		isError = true
		return 
	}
	for i := uint32(0); i < atomic.LoadUint32(&event.size); i++ {
		go func() {
			select{
			case <-time.After(1*time.Second/2):
				if DEBUG {
					log.Println("cannot emit to poller", i)
				}
			case event.data <- data:
				// nothing
			}
		}()
	}
	delete(self.events, eventName)
	return
}
func NewManager(opt Options) (mgr Manager, err error) {
	if opt.MaxConnection == 0 {
		err = errors.New("invalid MaxConnection values")
		return
	}
	mgr = Manager{
		opt:opt, 
		events: map[string]*lpEvent{},
	}
	if DEBUG {
		log.Printf("Options: %#v\n", opt)
	}
	return
}