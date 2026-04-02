package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type T1 struct {
	name string
}

func NewT1() *T1 {
	return &T1{name: "type1"}
}

func (t *T1) Reset() {
	println(t.name)
}

type T2 struct {
	name string
}

func NewT2() *T2 {
	return &T2{name: "type2"}
}

func (t *T2) Reset() {
	println(t.name)
}

func TestNew(t *testing.T) {
	resultData := New(NewT1)
	require.NotNil(t, resultData)
}

func TestGet(t *testing.T) {
	t.Run("Type1", func(t *testing.T) {
		pool := New(NewT1)
		resultData := pool.Get()
		require.NotNil(t, resultData)
		require.IsType(t, &T1{}, resultData)
		require.Equal(t, "type1", resultData.name)
	})

	t.Run("Type2", func(t *testing.T) {
		pool := New(NewT2)
		resultData := pool.Get()
		require.NotNil(t, resultData)
		require.IsType(t, &T2{}, resultData)
		require.Equal(t, "type2", resultData.name)
	})
}

func TestPut(t *testing.T) {
	pool := New(NewT1)
	item := pool.Get()
	require.Equal(t, "type1", item.name)
	item.name = "newType"
	pool.Put(item)
	require.Equal(t, "newType", pool.Get().name)
}
