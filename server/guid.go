package server

// This code is copied from `nsq.io's guid.go`

import (
	"encoding/hex"
	"errors"
	"time"
)

const (
	workerIDBits   = uint64(10)
	sequenceBits   = uint64(12)
	workerIDShift  = sequenceBits
	timestampShift = sequenceBits + workerIDBits
	sequenceMask   = int64(-1) ^ (int64(-1) << sequenceBits)

	// Tue, 21 Mar 2006 20:50:14.000 GMT
	twepoch = int64(1288834974657)
)

var ErrTimeBackwards = errors.New("time has gone backwards")
var ErrSequenceExpired = errors.New("sequence expired")

type guid int64

type guidFactory struct {
	sequence      int64
	lastTimestamp int64
}

func (f *guidFactory) NewGUID(workerID int64) (guid, error) {
	ts := time.Now().UnixNano() / 1e6

	if ts < f.lastTimestamp {
		return 0, ErrTimeBackwards
	}

	if f.lastTimestamp == ts {
		f.sequence = (f.sequence + 1) & sequenceMask
		if f.sequence == 0 {
			return 0, ErrSequenceExpired
		}
	} else {
		f.sequence = 0
	}

	f.lastTimestamp = ts

	id := ((ts - twepoch) << timestampShift) |
		(workerID << workerIDShift) |
		f.sequence

	return guid(id), nil
}

func (g guid) Hex() MessageID {
	var h MessageID
	var b [8]byte

	b[0] = byte(g >> 56)
	b[1] = byte(g >> 48)
	b[2] = byte(g >> 40)
	b[3] = byte(g >> 32)
	b[4] = byte(g >> 24)
	b[5] = byte(g >> 16)
	b[6] = byte(g >> 8)
	b[7] = byte(g)

	hex.Encode(h[:], b[:])
	return h
}
