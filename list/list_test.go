package list_test

import (
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/MasterOfBinary/redistypes/internal"
	"github.com/MasterOfBinary/redistypes/internal/test"
	"github.com/MasterOfBinary/redistypes/list"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

var conn redis.Conn

const (
	forwardSlice = false
	reverseSlice = true

	leftBlockingPop  = true
	rightBlockingPop = false
)

type scenarioStruct struct {
	name    string
	add     []interface{}
	wantErr bool
}

func ExampleNewRedisList() {
	netConn, _ := net.Dial("tcp", internal.GetHostAndPort())

	conn := redis.NewConn(netConn, time.Second, time.Second)
	defer conn.Close()

	l := list.NewRedisList(conn, test.RandomKey())

	values, _ := l.Range(0, -1)
	fmt.Println("Count:", len(values))
	// Output: Count: 0
}

func ExampleList_Range() {
	netConn, _ := net.Dial("tcp", internal.GetHostAndPort())

	conn := redis.NewConn(netConn, time.Second, time.Second)
	defer conn.Close()

	l := list.NewRedisList(conn, test.RandomKey())

	_, _ = l.RightPush("hello", "world", "how", "are", "you", "today")

	// Get the entire range
	values, _ := redis.Strings(l.Range(0, -1))
	fmt.Println(values)

	// Starting from the middle
	values, _ = redis.Strings(l.Range(2, -1))
	fmt.Println(values)

	// From beginning to middle
	values, _ = redis.Strings(l.Range(0, 1))
	fmt.Println(values)

	// Output:
	// [hello world how are you today]
	// [how are you today]
	// [hello world]
}

func TestRedisList_Name(t *testing.T) {
	name := test.RandomKey()
	l := list.NewRedisList(conn, name)
	assert.Equal(t, name, l.Name())
}

func TestRedisList_BlockingLeftPop(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	blockingPopTest(t, l, leftBlockingPop)
}

func TestRedisList_BlockingRightPop(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	blockingPopTest(t, l, rightBlockingPop)
}

func TestRedisList_LeftPop(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	t.Run("non-existing key", func(t *testing.T) {
		item, err := l.LeftPop()
		assert.Nil(t, err)
		assert.Nil(t, item)
	})

	t.Run("list with one item", func(t *testing.T) {
		_, _ = l.LeftPush("abc")
		item, err := redis.String(l.LeftPop())
		assert.Nil(t, err)
		assert.EqualValues(t, "abc", item)
	})

	t.Run("list with multiple items", func(t *testing.T) {
		_, _ = l.LeftPush("def", "ghi")
		item, err := redis.String(l.LeftPop())
		assert.Nil(t, err)
		assert.EqualValues(t, "ghi", item)
	})
}

func TestRedisList_LeftPush(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	scenarios := []scenarioStruct{
		{
			name: "non-existing key",
			add:  test.StringsToInterfaceSlice("abc"),
		},
		{
			name: "existing key",
			add:  test.StringsToInterfaceSlice("abc"),
		},
		{
			name: "add integers",
			add:  test.IntsToInterfaceSlice(1, 2),
		},
		{
			name: "add a string",
			add:  test.StringsToInterfaceSlice("abc"),
		},
		{
			name:    "no items",
			add:     nil,
			wantErr: true,
		},
	}

	wantCount := 0

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			count, err := l.LeftPush(scenario.add...)
			if scenario.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				wantCount += len(scenario.add)
				assert.EqualValues(t, wantCount, count)
			}
		})
	}

	t.Run("verify list data", func(t *testing.T) {
		verifySlice(t, l, wantCount, scenarios, reverseSlice)
	})
}

func TestRedisList_LeftPushX(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	t.Run("non-existing key", func(t *testing.T) {
		count, err := l.LeftPushX("abc")
		assert.Nil(t, err)
		assert.EqualValues(t, 0, count)
	})

	_, _ = l.RightPush("abc")

	scenarios := []scenarioStruct{
		{
			name: "existing key",
			add:  test.StringsToInterfaceSlice("abc"),
		},
		{
			name: "add integer",
			add:  test.IntsToInterfaceSlice(1),
		},
	}

	wantCount := 1

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			count, err := l.LeftPushX(scenario.add[0])
			if scenario.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				wantCount += len(scenario.add)
				assert.EqualValues(t, wantCount, count)
			}
		})
	}

	// Pop the one added by RightPush
	_, _ = l.RightPop()
	wantCount--

	t.Run("verify list data", func(t *testing.T) {
		verifySlice(t, l, wantCount, scenarios, reverseSlice)
	})
}

func TestRedisList_Length(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	t.Run("non-existing key", func(t *testing.T) {
		len, err := l.Length()
		assert.Nil(t, err)
		assert.Zero(t, len)
	})

	t.Run("list with a single value", func(t *testing.T) {
		l.LeftPush("abc")
		len, err := l.Length()
		assert.Nil(t, err)
		assert.EqualValues(t, 1, len)
	})

	t.Run("list with several values", func(t *testing.T) {
		l.LeftPush("def", "ghi")
		len, err := l.Length()
		assert.Nil(t, err)
		assert.EqualValues(t, 3, len)
	})

	t.Run("after removing all items", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			_, _ = l.LeftPop()
		}
		len, err := l.Length()
		assert.Nil(t, err)
		assert.EqualValues(t, 0, len)
	})
}

func TestRedisList_Range(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	t.Run("non-existing key", func(t *testing.T) {
		items, err := l.Range(0, -1)
		assert.Nil(t, err)
		assert.Empty(t, items)
	})

	added := test.StringsToInterfaceSlice("abc", "def", "ghi")
	added = append(added, 5)
	added = append(added, "jkl")
	added = append(added, 10)

	_, _ = l.RightPush(added...)

	t.Run("full range", func(t *testing.T) {
		items, err := l.Range(0, -1)
		assert.Nil(t, err)
		assert.Len(t, items, len(added))
		for i, item := range items {
			test.AssertEqual(t, added[i], item)
		}
	})

	t.Run("from start to middle", func(t *testing.T) {
		items, err := l.Range(0, 2)
		assert.Nil(t, err)
		assert.Len(t, items, 3)
		for i, item := range items {
			test.AssertEqual(t, added[i], item)
		}
	})

	t.Run("from middle to end", func(t *testing.T) {
		items, err := l.Range(3, -1)
		assert.Nil(t, err)
		assert.Len(t, items, 3)
		for i, item := range items {
			test.AssertEqual(t, added[i+3], item)
		}
	})
}

func TestRedisList_RightPop(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	t.Run("non-existing key", func(t *testing.T) {
		item, err := l.RightPop()
		assert.Nil(t, err)
		assert.Nil(t, item)
	})

	t.Run("list with one item", func(t *testing.T) {
		_, _ = l.RightPush("abc")
		item, err := redis.String(l.RightPop())
		assert.Nil(t, err)
		assert.EqualValues(t, "abc", item)
	})

	t.Run("list with multiple items", func(t *testing.T) {
		_, _ = l.RightPush("def", "ghi")
		item, err := redis.String(l.RightPop())
		assert.Nil(t, err)
		assert.EqualValues(t, "ghi", item)
	})
}

func TestRedisList_RightPush(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	scenarios := []scenarioStruct{
		{
			name: "non-existing key",
			add:  test.StringsToInterfaceSlice("abc"),
		},
		{
			name: "existing key",
			add:  test.StringsToInterfaceSlice("abc"),
		},
		{
			name: "add integers",
			add:  test.IntsToInterfaceSlice(1, 2),
		},
		{
			name: "add a string",
			add:  test.StringsToInterfaceSlice("abc"),
		},
		{
			name:    "no items",
			add:     nil,
			wantErr: true,
		},
	}

	wantCount := 0

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			count, err := l.RightPush(scenario.add...)
			if scenario.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				wantCount += len(scenario.add)
				assert.EqualValues(t, wantCount, count)
			}
		})
	}

	t.Run("verify list data", func(t *testing.T) {
		verifySlice(t, l, wantCount, scenarios, forwardSlice)
	})
}

func TestRedisList_RightPushX(t *testing.T) {
	l := list.NewRedisList(conn, test.RandomKey())

	t.Run("non-existing key", func(t *testing.T) {
		count, err := l.RightPushX("abc")
		assert.Nil(t, err)
		assert.EqualValues(t, 0, count)
	})

	_, _ = l.RightPush("abc")

	scenarios := []scenarioStruct{
		{
			name: "existing key",
			add:  test.StringsToInterfaceSlice("abc"),
		},
		{
			name: "add integer",
			add:  test.IntsToInterfaceSlice(1),
		},
	}

	wantCount := 1

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			count, err := l.RightPushX(scenario.add[0])
			if scenario.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				wantCount += len(scenario.add)
				assert.EqualValues(t, wantCount, count)
			}
		})
	}

	// Pop the one added by RightPush
	_, _ = l.LeftPop()
	wantCount--

	t.Run("verify list data", func(t *testing.T) {
		verifySlice(t, l, wantCount, scenarios, forwardSlice)
	})
}

func TestMain(m *testing.M) {
	netConn, err := net.Dial("tcp", internal.GetHostAndPort())
	if err != nil {
		fmt.Printf("Error opening net connection, err: %v", err)
		os.Exit(1)
	}

	conn = redis.NewConn(netConn, time.Second, time.Second)
	defer conn.Close()

	os.Exit(m.Run())
}

func clearList(l list.List) {
	values, _ := l.Range(0, -1)
	for range values {
		_, _ = l.RightPop()
	}
}

func verifySlice(t *testing.T, l list.List, wantCount int, scenarios []scenarioStruct, reverse bool) {
	got, err := l.Range(0, -1)
	assert.Nil(t, err)
	assert.EqualValues(t, wantCount, len(got))
	assert.Len(t, got, wantCount)
	i := 0
	for _, scenario := range scenarios {
		if !scenario.wantErr {
			for _, item := range scenario.add {
				index := i
				if reverse == reverseSlice {
					index = len(got) - i - 1
				}
				test.AssertEqual(t, item, got[index])
				i++
			}
		}
	}
}

func blockingPopTest(t *testing.T, l list.List, blockingPop bool) {
	defer clearList(l)

	t.Run("list with one item", func(t *testing.T) {
		_, _ = l.LeftPush("abc")
		var item interface{}
		var err error
		if blockingPop == leftBlockingPop {
			item, err = redis.String(l.BlockingLeftPop(0))
		} else {
			item, err = redis.String(l.BlockingRightPop(0))
		}
		assert.Nil(t, err)
		assert.EqualValues(t, "abc", item)
	})

	t.Run("timeout tests", func(t *testing.T) {
		scenarios := []struct {
			name     string
			duration time.Duration
			wantErr  bool
		}{
			{
				name:     "1s",
				duration: time.Second,
				wantErr:  false,
			},
			{
				name:     "10s",
				duration: 10 * time.Second,
				wantErr:  false,
			},
			{
				name:     "1s5ms",
				duration: 1*time.Second + 5*time.Millisecond,
				wantErr:  true,
			},
			{
				name:     "5ns",
				duration: 5 * time.Nanosecond,
				wantErr:  true,
			},
			{
				name:     "1s950ms",
				duration: 1*time.Second + 950*time.Millisecond,
				wantErr:  true,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				_, _ = l.RightPush(1)

				var err error
				if blockingPop == leftBlockingPop {
					_, err = redis.String(l.BlockingLeftPop(scenario.duration))
				} else {
					_, err = redis.String(l.BlockingRightPop(scenario.duration))
				}
				if scenario.wantErr {
					assert.NotNil(t, err)
					_, _ = l.RightPop()
				} else {
					assert.Nil(t, err)
				}
			})
		}
	})

	t.Run("blocking test", func(t *testing.T) {
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			netConn, _ := net.Dial("tcp", internal.GetHostAndPort())

			conn2 := redis.NewConn(netConn, 5*time.Second, 5*time.Second)
			defer conn2.Close()

			l2 := list.NewRedisList(conn2, l.Name())

			var item interface{}
			var err error
			if blockingPop == leftBlockingPop {
				item, err = l2.BlockingLeftPop(2 * time.Second)
			} else {
				item, err = l2.BlockingRightPop(2 * time.Second)
			}
			assert.Nil(t, err)
			test.AssertEqual(t, 1, item)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
			if blockingPop == leftBlockingPop {
				_, _ = l.LeftPush(1)
			} else {
				_, _ = l.RightPush(1)
			}
		}()

		wg.Wait()
	})

	t.Run("left or right test", func(t *testing.T) {
		_, _ = l.RightPush(1, 2, 3, 4, 5)
		if blockingPop == leftBlockingPop {
			value, err := l.BlockingLeftPop(0)
			assert.Nil(t, err)
			test.AssertEqual(t, 1, value)
		} else {
			value, err := l.BlockingRightPop(0)
			assert.Nil(t, err)
			test.AssertEqual(t, 5, value)
		}
	})
}
