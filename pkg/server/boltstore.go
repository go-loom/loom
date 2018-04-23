package server

import (
	"github.com/boltdb/bolt"
	"github.com/go-loom/loom/log"

	"time"
)

var (
	boltBucketMessages = []byte("messages")
)

type BoltStore struct {
	Path string
	db   *bolt.DB
}

type BoltMessageBucket struct {
	name           []byte
	db             *bolt.DB
	ttl            time.Duration
	maxExpireItems int
	logger         log.Logger
}

func NewBoltStore(path string) *BoltStore {
	bs := &BoltStore{
		Path: path,
	}

	return bs
}

func (bs *BoltStore) Open() error {
	db, err := bolt.Open(bs.Path, 0600, nil)
	if err != nil {
		return err
	}

	bs.db = db
	err = bs.db.Update(func(tx *bolt.Tx) error {

		buckets := [][]byte{boltBucketMessages}
		for _, b := range buckets {
			_, err = tx.CreateBucketIfNotExists(b)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (bs *BoltStore) Close() error {
	err := bs.db.Close()
	return err
}

func (bs *BoltStore) MessageBucket(name string) MessageBucket {
	ttl, _ := time.ParseDuration("720h") // Default TTL TODO: MessageBucketWithTTL?
	b := &BoltMessageBucket{
		name:           []byte(name),
		db:             bs.db,
		ttl:            ttl,
		maxExpireItems: 30,
		logger:         log.New("MessageBucket"),
	}
	return b
}

func (b *BoltMessageBucket) bucket(tx *bolt.Tx) *bolt.Bucket {
	bucket := tx.Bucket(boltBucketMessages).Bucket(b.name)
	return bucket
}

func (b *BoltMessageBucket) createBucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	bucket := tx.Bucket(boltBucketMessages)
	bucket, err := bucket.CreateBucketIfNotExists(b.name)
	return bucket, err
}

func (b *BoltMessageBucket) Get(id MessageID) (*Message, error) {
	b.expireMessages()

	var m *Message
	err := b.db.View(func(tx *bolt.Tx) error {
		b := b.bucket(tx)
		if b == nil {
			return nil
		}

		v := b.Get(id.Bytes())

		var err error
		m, err = DecodeMessage(v)
		return err
	})

	return m, err

}

func (b *BoltMessageBucket) Put(msg *Message) error {
	b.expireMessages()

	err := b.db.Update(func(tx *bolt.Tx) error {
		b, err := b.createBucket(tx)
		if err != nil {
			return err
		}
		err = b.Put(msg.ID.Bytes(), msg.Encode())
		return err
	})

	return err
}

func (b *BoltMessageBucket) Del(id MessageID) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		b, err := b.createBucket(tx)
		if err != nil {
			return err
		}
		err = b.Delete(id.Bytes())
		return err
	})
	return err
}

func (b *BoltMessageBucket) Walk(walkFunc func(msg *Message) error) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := b.bucket(tx)
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			msg, err := DecodeMessage(v)
			if err != nil {
				return err
			}
			err = walkFunc(msg)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (b *BoltMessageBucket) DelBucket() error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(boltBucketMessages).DeleteBucket(b.name)
		return err
	})
	return err

}

func (b *BoltMessageBucket) expireMessages() {
	now := time.Now()

	err := b.db.Update(func(tx *bolt.Tx) error {
		i := 0
		bucket := b.bucket(tx)
		if bucket == nil {
			return nil
		}
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			i++
			if i >= b.maxExpireItems {
				return nil
			}
			if v == nil {
				continue
			}

			m, err := DecodeMessage(v)
			if err != nil {
				b.logger.Error("err: %v", err)
				return err
			}
			if m == nil {
				continue
			}

			if now.Sub(m.Created) >= b.ttl {
				b.logger.Info("Expire message: id:%s created:%v", string(m.ID[:]), m.Created)
				bucket.Delete(k)
			}
		}

		return nil
	})

	if err != nil {
		b.logger.Error("expire: err: %v", err)
	}
}
