package goche

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"math"
	"testing"
	"time"
)

func TestSuiteLRU(t *testing.T) {
	suite.Run(t, new(SuiteLRU))
}

type SuiteLRU struct {
	suite.Suite

	cache Cache

	key string
	val any
}

func (s *SuiteLRU) SetupTest() {
	s.cache = NewLRU()

	s.key = "key"
	s.val = []byte("values")
}

func (s *SuiteLRU) TestSetGet() {
	s.cache.Set(context.Background(), s.key, s.val)

	val, ok := s.cache.Get(context.Background(), s.key)

	require.True(s.T(), ok)
	require.Equal(s.T(), s.val, val)
}

func (s *SuiteLRU) TestTableSetGet() {
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

func (s *SuiteLRU) TestExtrusion() {
	c := New(LRU, WithLimit(10))

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
		{
			key: "11",
		},
	}

	for _, tc := range tests {
		c.Set(context.Background(), tc.key, nil)
	}

	// giving time to delete least recently used value
	time.Sleep(time.Millisecond * 10)

	val, ok := c.Get(context.Background(), tests[0].key)
	require.False(s.T(), ok)
	require.Nil(s.T(), val)
}

func (s *SuiteLRU) TestWithLimitMod() {
	c := New(LRU, WithLimit(3))

	c.Set(context.Background(), "key", 1)
	c.Set(context.Background(), "key1", 2)
	c.Set(context.Background(), "key2", 3)
	c.Set(context.Background(), "key3", 4)

	// giving time to delete least recently used value
	time.Sleep(time.Millisecond * 10)

	val, ok := c.Get(context.Background(), "key")
	require.False(s.T(), ok)
	require.Nil(s.T(), val)
}

func (s *SuiteLRU) TestReplace() {
	c := New(LRU, WithLimit(2))

	c.Set(context.Background(), "key", 1)
	c.Set(context.Background(), "key1", 2)

	time.Sleep(time.Millisecond * 10)
	c.Get(context.Background(), "key")

	time.Sleep(time.Millisecond * 10)

	c.Set(context.Background(), "key2", 2)

	time.Sleep(time.Millisecond * 150)

	val, ok := c.Get(context.Background(), "key")
	require.True(s.T(), ok)
	require.NotNil(s.T(), val)

	val, ok = c.Get(context.Background(), "key1")
	require.False(s.T(), ok)
	require.Nil(s.T(), val)

	val, ok = c.Get(context.Background(), "key2")
	require.True(s.T(), ok)
	require.NotNil(s.T(), val)
}
