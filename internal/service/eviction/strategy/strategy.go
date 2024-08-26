package strategy

import "time"

// Strategy Design Pattern
type EvictionStrategy interface {
	Get(string) (Value, *time.Time, bool)
	Put(string, Value)
	CleanUp(ttl time.Duration)
	Len() int
}

type Value interface {
	Len() int
}

// ttl support for delete stale cached data
type Entry struct {
	Key      string
	Value    Value
	UpdateAt *time.Time
}

// ttl expired function
func (ele *Entry) Expired(duration time.Duration) (ok bool) {
	if ele.UpdateAt == nil {
		ok = false
	} else {
		ok = ele.UpdateAt.Add(duration).Before(time.Now())
	}
	return
}

// ttl touch function
func (ele *Entry) Touch() {
	//ele.UpdateAt=time.Now()
	nowTime := time.Now()
	ele.UpdateAt = &nowTime
}
