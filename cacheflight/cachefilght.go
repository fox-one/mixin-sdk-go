package cacheflight

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	DefaultCacheDirty      = time.Minute * 5
	DefaultCacheExpiration = time.Minute * 1
)

var ErrUnmodified = errors.New("cacheflight: unmodified")

type Fn func() (val interface{}, err error)

type ShouldCache func(val interface{}, err error) (bCache bool, expire, dirty time.Duration)

/*
  - sometimes you may need the oldVal to do logic.
    eg. consider a case like http 304(Not modified),
    server is told the oldVal's hash to decide if it's modified
*/
type FnHijack func(oldVal interface{}, oldErr error) (val interface{}, err error)

type cacheResult struct {
	val    interface{}
	err    error
	ctime  time.Time
	expire time.Duration
	dirty  time.Duration
}

type Group struct {
	cacheExpiration time.Duration
	cacheDirty      time.Duration

	defaultShouldCache ShouldCache
	mu                 sync.RWMutex // protects cache
	cache              map[string]cacheResult
	sf                 *singleflight.Group
}

func New(cacheExpiration, cacheDirty time.Duration) *Group {

	if cacheExpiration == 0 {
		cacheExpiration = DefaultCacheExpiration
	}
	if cacheDirty == 0 {
		cacheDirty = DefaultCacheDirty
	}
	if cacheExpiration > cacheDirty {
		panic(fmt.Sprintf("cacheExpiration(%v) should <= cacheDirty(%v)", cacheExpiration, cacheDirty))
	}
	var defaultShouldCache = func(val interface{}, err error) (bCache bool, expire, dirty time.Duration) {
		return err == nil, cacheExpiration, cacheDirty
	}
	g := &Group{
		cacheExpiration:    cacheExpiration,
		cacheDirty:         cacheDirty,
		defaultShouldCache: defaultShouldCache,
		cache:              make(map[string]cacheResult),
		sf:                 &singleflight.Group{},
	}
	go g.loopClearCache()
	return g
}

func (g *Group) NoSingleFlight() *Group {
	g.sf = nil
	return g
}

func (g *Group) loopClearCache() {
	duration := g.cacheExpiration
	if duration < 5*time.Second {
		duration = 5 * time.Second
	}
	ticker := time.Tick(duration)
	for {
		select {
		case now := <-ticker:
			g.mu.Lock()
			for k, v := range g.cache {
				if now.After(v.ctime.Add(v.dirty)) {
					delete(g.cache, k)
				}
			}
			g.mu.Unlock()
		}
	}
}

func (g *Group) doFlight(key string, fn FnHijack, shouldCache ShouldCache) (interface{}, error) {

	if g.sf != nil {
		v, err, _ := g.sf.Do(key, func() (interface{}, error) {
			return g.do(key, fn, shouldCache)
		})
		return v, err
	} else {
		return g.do(key, fn, shouldCache)
	}
}

func (g *Group) do(key string, fn FnHijack, shouldCache ShouldCache) (interface{}, error) {
	g.mu.RLock()
	cr, ok := g.cache[key]
	g.mu.RUnlock()

	now := time.Now()
	if ok && now.Before(cr.ctime.Add(cr.expire)) {
		return cr.val, cr.err
	}
	val, err := fn(cr.val, cr.err)
	if err == ErrUnmodified {
		g.mu.Lock()
		cr.ctime = time.Now()
		g.cache[key] = cr
		g.mu.Unlock()
		return cr.val, cr.err
	}

	if shouldCache == nil {
		shouldCache = g.defaultShouldCache
	}
	bCache, expire, dirty := shouldCache(val, err)
	if expire == 0 || dirty == 0 || expire > dirty {
		expire = g.cacheExpiration
		dirty = g.cacheDirty
	}
	if bCache {
		g.mu.Lock()
		result := cacheResult{val: val, err: err, ctime: time.Now(), expire: expire, dirty: dirty}
		g.cache[key] = result
		g.mu.Unlock()
	}
	return val, err
}

func (g *Group) Do(key string, fn Fn) (interface{}, error) {
	return g.DoWithCondition(key, fn, nil)
}

func (g *Group) DoWithCondition(key string, fn Fn, shouldCache ShouldCache) (interface{}, error) {
	fnHijack := func(oldVal interface{}, oldErr error) (val interface{}, err error) {
		return fn()
	}
	return g.HijackDoWithCondition(key, fnHijack, shouldCache)

}

func (g *Group) HijackDo(key string, fn FnHijack) (interface{}, error) {
	return g.HijackDoWithCondition(key, fn, nil)
}

func (g *Group) HijackDoWithCondition(key string, fn FnHijack, shouldCache ShouldCache) (interface{}, error) {
	g.mu.RLock()
	cr, ok := g.cache[key]
	g.mu.RUnlock()

	if !ok {
		return g.doFlight(key, fn, shouldCache)
	}
	now := time.Now()
	if now.After(cr.ctime.Add(cr.dirty)) {
		return g.doFlight(key, fn, shouldCache)
	}
	if now.After(cr.ctime.Add(cr.expire)) {
		go g.doFlight(key, fn, shouldCache)
	}
	return cr.val, cr.err
}
