package worker

import "testing"

func TestPoolSimple(t *testing.T) {
	pool, err := NewPool(5)
	if err != nil {
		t.Error(err)
	}

	err = pool.Start()
	if err != nil {
		t.Error(err)
	}
}
