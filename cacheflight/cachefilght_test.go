package cacheflight

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo(t *testing.T) {
	count := uint32(0)
	g := New(0, 0)
	for i := 0; i < 10; i++ {
		val, err := g.Do("haha", func() (interface{}, error) {
			atomic.AddUint32(&count, 1)
			return "hoho", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "hoho", val)
	}
	assert.Equal(t, 1, count)

	count = 0
	g2 := New(0, 0).NoSingleFlight()
	for i := 0; i < 10; i++ {
		val, err := g2.Do("haha2", func() (interface{}, error) {
			atomic.AddUint32(&count, 1)
			return "hoho", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "hoho", val)
	}
	assert.Equal(t, 1, count)
}

func TestDoErr(t *testing.T) {
	count := uint32(0)
	g := New(0, 0)
	someErr := errors.New("Some error")
	for i := 0; i < 10; i++ {
		v, err := g.Do("key", func() (interface{}, error) {
			atomic.AddUint32(&count, 1)
			return nil, someErr
		})
		assert.Equal(t, someErr, err)
		assert.Nil(t, v)
	}
	assert.Equal(t, 10, count)

	count = 0
	g2 := New(0, 0).NoSingleFlight()
	for i := 0; i < 10; i++ {
		v, err := g2.Do("key", func() (interface{}, error) {
			atomic.AddUint32(&count, 1)
			return nil, someErr
		})
		assert.Equal(t, someErr, err)
		assert.Nil(t, v)
	}
	assert.Equal(t, 10, count)
}

func TestDoWithCondition(t *testing.T) {
	count := uint32(0)
	g := New(0, 0)
	someErr := errors.New("Some error")
	for i := 0; i < 10; i++ {
		v, err := g.DoWithCondition(
			"key",
			func() (interface{}, error) {
				atomic.AddUint32(&count, 1)
				return nil, someErr
			},
			func(val interface{}, err error) (bCache bool, expire, dirty time.Duration) {
				if err == someErr {
					return true, DefaultCacheExpiration, DefaultCacheDirty
				}
				return false, DefaultCacheExpiration, DefaultCacheDirty
			},
		)
		assert.Equal(t, someErr, err)
		assert.Nil(t, v)
	}
	assert.Equal(t, 1, count)

	count = 0
	g2 := New(0, 0).NoSingleFlight()
	for i := 0; i < 10; i++ {
		v, err := g2.DoWithCondition(
			"key",
			func() (interface{}, error) {
				atomic.AddUint32(&count, 1)
				return nil, someErr
			},
			func(val interface{}, err error) (bCache bool, expire, dirty time.Duration) {
				if err == someErr {
					return true, DefaultCacheExpiration, DefaultCacheDirty
				}
				return false, DefaultCacheExpiration, DefaultCacheDirty
			},
		)
		assert.Equal(t, someErr, err)
		assert.Nil(t, v)
	}
	assert.Equal(t, 1, count)
}

func TestDoDupSuppress(t *testing.T) {
	g := New(0, 0)
	var calls int32
	c := make(chan string)
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		return <-c, nil
	}

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			v, err := g.Do("key", fn)
			assert.NoError(t, err)
			assert.Equal(t, "bar", v.(string))
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond) // let goroutines above block
	c <- "bar"
	wg.Wait()
	assert.Equal(t, 1, atomic.LoadInt32(&calls))
}

func TestDoExpiration(t *testing.T) {
	g := New(time.Second, 0)
	var calls int32
	c := make(chan string)
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		return <-c, nil
	}

	go func() {
		c <- "bar1"
	}()
	v, err := g.Do("key", fn)
	assert.NoError(t, err)
	assert.Equal(t, "bar1", v)
	assert.Equal(t, 1, atomic.LoadInt32(&calls))

	for i := 0; i < 10; i++ {
		v, err := g.Do("key", fn)
		assert.NoError(t, err)
		assert.Equal(t, "bar1", v.(string))
	}
	assert.Equal(t, 1, atomic.LoadInt32(&calls))

	time.Sleep(time.Second + 500*time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			v, err := g.Do("key", fn)
			assert.NoError(t, err)
			assert.Equal(t, "bar1", v.(string))
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond) // let goroutines above block
	wg.Wait()
	c <- "bar2" // wg.Wait() can finishes earlier than fn()
	assert.Equal(t, 2, atomic.LoadInt32(&calls))

	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 10; i++ {
		v, err := g.Do("key", fn)
		assert.NoError(t, err)
		assert.Equal(t, "bar2", v.(string))
	}
	assert.Equal(t, 2, atomic.LoadInt32(&calls))
}

func TestDoDirty(t *testing.T) {
	g := New(time.Second/2, time.Second)

	var calls int32
	c := make(chan string)
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		return <-c, nil
	}

	go func() {
		c <- "bar1"
	}()
	v, err := g.Do("key", fn)
	assert.NoError(t, err)
	assert.Equal(t, "bar1", v)
	assert.Equal(t, 1, atomic.LoadInt32(&calls))

	time.Sleep(time.Second + 500*time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			v, err := g.Do("key", fn)
			assert.NoError(t, err)
			assert.Equal(t, "bar2", v.(string))
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond) // let goroutines above block
	c <- "bar2"
	wg.Wait()
	assert.Equal(t, 2, atomic.LoadInt32(&calls))
}

func TestHijackDoWithCondition(t *testing.T) {
	g := New(300*time.Millisecond, 500*time.Millisecond)

	var calls, modifies int32
	fn := func(oldVal interface{}, oldErr error) (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		if oldVal != nil && oldVal.(string) == "no change" {
			return "", ErrUnmodified
		}
		atomic.AddInt32(&modifies, 1)
		return "no change", nil
	}

	v, err := g.HijackDo("key", fn)
	assert.NoError(t, err)
	assert.Equal(t, "no change", v.(string))
	assert.Equal(t, 1, atomic.LoadInt32(&calls))
	assert.Equal(t, 1, atomic.LoadInt32(&modifies))

	time.Sleep(600 * time.Millisecond)

	v, err = g.HijackDo("key", fn)
	assert.NoError(t, err)
	assert.Equal(t, "no change", v.(string))
	assert.Equal(t, 2, atomic.LoadInt32(&calls))
	assert.Equal(t, 1, atomic.LoadInt32(&modifies))

	for i := 0; i < 10; i++ {
		v, err = g.HijackDo("key", fn)
		assert.NoError(t, err)
		assert.Equal(t, "no change", v.(string))
	}
	assert.Equal(t, 2, atomic.LoadInt32(&calls))
	assert.Equal(t, 1, atomic.LoadInt32(&modifies))
}

func TestHijackDoWithConditionExpire(t *testing.T) {
	g := New(300*time.Millisecond, 500*time.Millisecond)

	var calls, modifies int32
	fn := func(oldVal interface{}, oldErr error) (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		if oldVal != nil && oldVal.(string) == "no change" {
			return "", ErrUnmodified
		}
		atomic.AddInt32(&modifies, 1)
		return "no change", nil
	}
	shouleCahce := func(val interface{}, err error) (bCache bool, expire, dirty time.Duration) {
		return err == nil, 600 * time.Millisecond, 800 * time.Millisecond
	}
	v, err := g.HijackDoWithCondition("key", fn, shouleCahce)
	assert.NoError(t, err)
	assert.Equal(t, "no change", v.(string))
	assert.Equal(t, 1, atomic.LoadInt32(&calls))
	assert.Equal(t, 1, atomic.LoadInt32(&modifies))

	time.Sleep(850 * time.Millisecond)

	v, err = g.HijackDo("key", fn)
	assert.NoError(t, err)
	assert.Equal(t, "no change", v.(string))
	assert.Equal(t, 2, atomic.LoadInt32(&calls))
	assert.Equal(t, 1, atomic.LoadInt32(&modifies))

	for i := 0; i < 10; i++ {
		v, err = g.HijackDoWithCondition("key", fn, shouleCahce)
		assert.NoError(t, err)
		assert.Equal(t, "no change", v.(string))
	}
	assert.Equal(t, 2, atomic.LoadInt32(&calls))
	assert.Equal(t, 1, atomic.LoadInt32(&modifies))
}
