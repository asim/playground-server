package docker

import (
	"errors"
)

type uidPool struct {
	pool chan int
}

func newUidPool(lower int, upper int) *uidPool {
	size := upper - lower + 1

	// create uids
	pool := make(chan int, size)

	for i := 0; i < size; i++ {
		pool <- lower + i
	}

	return &uidPool{
		pool: pool,
	}
}

func (p *uidPool) Get() (int, error) {
	select {
	case uid := <-p.pool:
		return uid, nil
	default:
		return 0, errors.New("Uid pool is empty")
	}
}

func (p *uidPool) Put(uid int) error {
	select {
	case p.pool <- uid:
		return nil
	default:
		return errors.New("Uid pool is full")
	}
}
