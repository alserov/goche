package goche

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestSuiteLFU(t *testing.T) {
	suite.Run(t, new(SuiteLFU))
}

type SuiteLFU struct {
	suite.Suite

	cache Cache

	key string
	val any
}

func (s *SuiteLFU) SetupTest() {
	s.cache = NewLFU()

	s.key = "key"
	s.val = []byte("values")
}

func (s *SuiteLFU) TestSetGet() {
	s.cache.Set(context.Background(), s.key, s.val)

	val, ok := s.cache.Get(context.Background(), s.key)

	require.True(s.T(), ok)
	require.Equal(s.T(), s.val, val)
}

func (s *SuiteLFU) TestTableSetGet() {
	tests := []struct {
		key string
		val any
	}{
		{
			key: "1",
			val: []any{},
		},
		{
			key: "2",
			val: nil,
		},
		{
			key: "3",
			val: math.MaxInt64,
		},
		{
			key: "4",
			val: math.MaxFloat64,
		},
		{
			key: "5",
			val: "123",
		},
	}

	for _, tc := range tests {
		s.T().Run(fmt.Sprintf("Set: %s", tc.key), func(t *testing.T) {
			s.cache.Set(context.Background(), tc.key, tc.val)
		})
	}

	for _, tc := range tests {
		s.T().Run(fmt.Sprintf("Get: %s", tc.key), func(t *testing.T) {
			_, ok := s.cache.Get(context.Background(), tc.key)
			require.True(s.T(), ok)
		})
	}
}

func (s *SuiteLFU) TestExtrusion() {
	if os.Getenv("env") != "CI" {
		c := New(LFU, WithLimit(10))

		tests := []struct {
			key string
		}{
			{
				key: "1",
			},
			{
				key: "2",
			},
			{
				key: "3",
			},
			{
				key: "4",
			},
			{
				key: "5",
			},
			{
				key: "6",
			},
			{
				key: "7",
			},
			{
				key: "8",
			},
			{
				key: "9",
			},
			{
				key: "10",
			},
		}

		for _, tc := range tests {
			c.Set(context.Background(), tc.key, 52)
		}

		time.Sleep(time.Millisecond * 10)

		for _, tc := range tests[:len(tests)-1] {
			c.Get(context.Background(), tc.key)
		}

		time.Sleep(time.Millisecond * 10)

		c.Set(context.Background(), "11", 52)

		val, ok := c.Get(context.Background(), fmt.Sprintf("%d", rand.Intn(len(tests))))
		require.True(s.T(), ok)
		require.NotNil(s.T(), val)

		val, ok = c.Get(context.Background(), tests[len(tests)-1].key)
		require.False(s.T(), ok)
		require.Nil(s.T(), val)
	}
}

func (s *SuiteLFU) TestWithLimitMod() {
	if os.Getenv("env") != "CI" {
		c := New(LFU, WithLimit(2))

		tests := []struct {
			key string
		}{
			{
				key: "1",
			},
			{
				key: "2",
			},
			{
				key: "3",
			},
		}

		for _, tc := range tests {
			c.Set(context.Background(), tc.key, 52)
		}

		time.Sleep(time.Millisecond * 1)

		val, ok := c.Get(context.Background(), tests[0].key)
		require.False(s.T(), ok)
		require.Nil(s.T(), val)

		val, ok = c.Get(context.Background(), "3")
		require.True(s.T(), ok)
		require.NotNil(s.T(), val)
	}
}

func (s *SuiteLFU) TestWithClearMod() {
	c := New(LFU, WithClear(time.Millisecond*50))

	key := "key"

	c.Set(context.Background(), key, "val")

	val, ok := c.Get(context.Background(), key)
	require.True(s.T(), ok)
	require.NotNil(s.T(), val)

	time.Sleep(time.Millisecond * 100)

	val, ok = c.Get(context.Background(), key)
	require.False(s.T(), ok)
	require.Nil(s.T(), val)
}

func TestSwap(t *testing.T) {
	n1 := newLFUNode("1", 1)
	n2 := newLFUNode("2", 2)
	n1.next = n2
	n2.prev = n1

	l := lfu{}

	require.True(t, l.swap(n1, n2))

	require.Equal(t, n1.prev, n2)
	require.Equal(t, n2.next, n1)
	require.Nil(t, n1.next)
	require.Nil(t, n2.prev)
	require.Equal(t, l.head, n2)
	require.Equal(t, l.tail, n1)

	require.True(t, l.swap(n2, n1))

	require.Equal(t, n2.prev, n1)
	require.Equal(t, n1.next, n2)
	require.Nil(t, n2.next)
	require.Nil(t, n1.prev)
	require.Equal(t, l.head, n1)
	require.Equal(t, l.tail, n2)

	require.False(t, l.swap(nil, n1))
}
