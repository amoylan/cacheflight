package cacheflight

import "time"
import "sync"

type result struct {
	lastExecuteTime time.Time
	wg              sync.WaitGroup
	done            bool
	ret             interface{}
	err             error
}

// Group is core struct
type Group struct {
	sync.RWMutex
	resultMap  map[string]*result
	expiration time.Duration
}

// NewGroup return a new ttl cache group
func NewGroup(cacheExpiration time.Duration) (group *Group) {
	return &Group{
		resultMap:  make(map[string]*result),
		expiration: cacheExpiration,
	}
}

// Do cache
func (g *Group) Do(key string, fn func() (interface{}, error)) (ret interface{}, err error) {
	g.Lock()
	now := time.Now()
	if r, ok := g.resultMap[key]; ok {
		// 缓存时间内
		if now.Sub(r.lastExecuteTime) < g.expiration {
			g.Unlock()
			return r.ret, r.err
		}
		if r.done {
			r.done = false
		} else {
			g.Unlock()
			r.wg.Wait()
			return r.ret, r.err
		}
	}
	r := new(result)
	r.wg.Add(1)
	g.resultMap[key] = r
	g.Unlock()

	r.ret, r.err = fn()
	r.wg.Done()

	g.Lock()
	r.done = true
	r.lastExecuteTime = now
	g.Unlock()
	return r.ret, r.err
}
