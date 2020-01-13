//go:generate go run ./generate

package kbucket

import (
	"container/list"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

// Bucket holds a list of peers.
type Bucket struct {
	lk   sync.RWMutex
	list *list.List

	lastRefreshedAtLk sync.RWMutex
	lastRefreshedAt   time.Time // the last time we looked up a key in the bucket
}

func newBucket() *Bucket {
	b := new(Bucket)
	b.list = list.New()
	b.lastRefreshedAt = time.Now()
	return b
}

func (b *Bucket) RefreshedAt() time.Time {
	b.lastRefreshedAtLk.RLock()
	defer b.lastRefreshedAtLk.RUnlock()

	return b.lastRefreshedAt
}

func (b *Bucket) ResetRefreshedAt(newTime time.Time) {
	b.lastRefreshedAtLk.Lock()
	defer b.lastRefreshedAtLk.Unlock()

	b.lastRefreshedAt = newTime
}

func (b *Bucket) Peers() []peer.ID {
	b.lk.RLock()
	defer b.lk.RUnlock()
	ps := make([]peer.ID, 0, b.list.Len())
	for e := b.list.Front(); e != nil; e = e.Next() {
		id := e.Value.(peer.ID)
		ps = append(ps, id)
	}
	return ps
}

func (b *Bucket) Has(id peer.ID) bool {
	b.lk.RLock()
	defer b.lk.RUnlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		if e.Value.(peer.ID) == id {
			return true
		}
	}
	return false
}

func (b *Bucket) Remove(id peer.ID) bool {
	b.lk.Lock()
	defer b.lk.Unlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		if e.Value.(peer.ID) == id {
			b.list.Remove(e)
			return true
		}
	}
	return false
}

func (b *Bucket) MoveToFront(id peer.ID) {
	b.lk.Lock()
	defer b.lk.Unlock()
	for e := b.list.Front(); e != nil; e = e.Next() {
		if e.Value.(peer.ID) == id {
			b.list.MoveToFront(e)
		}
	}
}

func (b *Bucket) PushFront(p peer.ID) {
	b.lk.Lock()
	b.list.PushFront(p)
	b.lk.Unlock()
}

func (b *Bucket) PopBack() peer.ID {
	b.lk.Lock()
	defer b.lk.Unlock()
	last := b.list.Back()
	b.list.Remove(last)
	return last.Value.(peer.ID)
}

func (b *Bucket) Len() int {
	b.lk.RLock()
	defer b.lk.RUnlock()
	return b.list.Len()
}

// Split splits a buckets peers into two buckets, the methods receiver will have
// peers with CPL equal to cpl, the returned bucket will have peers with CPL
// greater than cpl (returned bucket has closer peers)
func (b *Bucket) Split(cpl int, target ID) *Bucket {
	b.lk.Lock()
	defer b.lk.Unlock()

	out := list.New()
	newbuck := newBucket()
	newbuck.list = out
	e := b.list.Front()
	for e != nil {
		peerID := ConvertPeerID(e.Value.(peer.ID))
		peerCPL := CommonPrefixLen(peerID, target)
		if peerCPL > cpl {
			cur := e
			out.PushBack(e.Value)
			e = e.Next()
			b.list.Remove(cur)
			continue
		}
		e = e.Next()
	}
	return newbuck
}
