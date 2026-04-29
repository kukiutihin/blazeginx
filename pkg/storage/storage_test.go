package storage

import (
	"strconv"
	"testing"
	"time"
)

func checkGet(act any, ok bool, req any, t *testing.T) {
	if !ok {
		t.Errorf("expected: %s, but got: %s", req, "")
	}
	if act != req {
		t.Errorf("expected: %s, but got: %s", req, act)
	}
}

func assertEquals(act any, req any, t *testing.T) {
	if act != req {
		t.Errorf("expected: %s, but got: %s", req, act)
	}
}

func fill(stor *Storage, n uint) {
	for i := range n {
		stor.Add(strconv.FormatInt(int64(i), 10), i)
	}
}

func TestValidAddGet(t *testing.T) {
	t.Parallel()
	stor := New(time.Hour, -1)
	assertEquals(0, stor.Size(), t)

	stor.Add("keks", "skibidi")
	res, ok := stor.Get("keks")
	checkGet(res, ok, "skibidi", t)
	assertEquals(1, stor.Size(), t)

	stor.Add("keks1", 88)
	res, ok = stor.Get("keks1")
	checkGet(res, ok, 88, t)
	assertEquals(2, stor.Size(), t)
}

func TestGetNotExists(t *testing.T) {
	t.Parallel()
	stor := New(time.Hour, -1)

	stor.Add("keks", 42)
	assertEquals(1, stor.Size(), t)

	_, ok := stor.Get("kekas")
	if ok {
		t.Errorf("expected: false, but got: true")
	}

	// check consistent
	assertEquals(stor.Size(), 1, t)
}

func TestAddDuplicate(t *testing.T) {
	t.Parallel()
	stor := New(time.Hour, -1)

	stor.Add("keks", "skibidi")
	assertEquals(stor.Size(), 1, t)

	stor.Add("keks", 42)
	res, ok := stor.Get("keks")
	checkGet(res, ok, 42, t)
	assertEquals(stor.Size(), 1, t)
}

func TestDeleteValid(t *testing.T) {
	t.Parallel()
	stor := New(time.Hour, -1)
	fill(stor, 11)
	assertEquals(stor.Size(), 11, t)

	stor.Delete("1")
	stor.Delete("2")

	assertEquals(stor.Size(), 9, t)
}

func TestDeleteNonExists(t *testing.T) {
	t.Parallel()
	stor := New(time.Hour, -1)

	// must do nothing
	stor.Delete("keks")
	assertEquals(stor.Size(), 0, t)

	stor.Add("keksoid", 1337)
	stor.Delete("keks")
	assertEquals(stor.Size(), 1, t)
}

func TestGC(t *testing.T) {
	t.Parallel()
	stor := New(time.Second*6, time.Second*3)

	stor.AddWithExpiration("skibop", 10, time.Second*1)
	stor.AddWithExpiration("skibobidi", 12, time.Second*1)
	stor.Add("long live skib", 13) // with default expiration
	assertEquals(stor.Size(), 3, t)

	time.Sleep(time.Second * 3)
	assertEquals(stor.Size(), 1, t)
	res, ok := stor.Get("long live skib")
	checkGet(res, ok, 13, t)

	time.Sleep(time.Second * 3)
	assertEquals(stor.Size(), 0, t)
}
