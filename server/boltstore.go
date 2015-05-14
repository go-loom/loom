package server

import (
	"bytes"
	"encoding/gob"
	"github.com/boltdb/bolt"
	"time"
)

var (
	bucket_messages    = []byte("messages")
	bucket_pending_ids = []byte("pending_ids")
	bucket_tasks       = []byte("tasks")
)

type BoltStore struct {
	Path string
	db   *bolt.DB
}

func NewBlotStore(path string) *BoltStore {
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

		buckets := [][]byte{bucket_messages, bucket_pending_ids, bucket_tasks}
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

func (bs *BoltStore) PutMessage(msg *Message) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket_messages)
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		enc.Encode(msg)

		return b.Put(msg.ID[:], buf.Bytes())
	})
	return err
}

func (bs *BoltStore) GetMessage(id MessageID) (*Message, error) {
	var msg *Message

	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket_messages)

		v := b.Get(id[:])

		var err error
		msg, err = bs.decodeMsg(v)
		if err != nil {
			return err
		}
		return nil
	})

	return msg, err
}

func (bs *BoltStore) RemoveMessage(id MessageID) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket_messages)
		if err != nil {
			return err
		}

		err = b.Delete(id[:])
		return err
	})

	return err
}

func (bs *BoltStore) WalkMessage(walkFunc func(*Message) error) error {
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket_messages)

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			msg, err := bs.decodeMsg(v)
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

func (bs *BoltStore) GetPendingMsgIDList(st *time.Time, ed *time.Time) ([]MessageID, error) {
	var ids []MessageID
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket_pending_ids)
		c := b.Cursor()

		b_st := []byte(st.Format(time.RFC3339))
		b_ed := []byte(ed.Format(time.RFC3339))

		for k, v := c.Seek(b_st); k != nil && bytes.Compare(k, b_ed) <= 0; k, v = c.Next() {

			var id MessageID
			copy(id[:], v)
			ids = append(ids, id)
		}

		return nil
	})

	return ids, err
}

func (bs *BoltStore) WalkPendingMsgId(st *time.Time, ed *time.Time,
	walkFunc func(ts *time.Time, id MessageID) error) error {
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket_pending_ids)
		c := b.Cursor()

		b_st := []byte(st.Format(time.RFC3339))
		b_ed := []byte(ed.Format(time.RFC3339))

		for k, v := c.Seek(b_st); k != nil && bytes.Compare(k, b_ed) <= 0; k, v = c.Next() {

			ts, err := time.Parse(time.RFC3339, string(k))

			var id MessageID
			copy(id[:], v)
			err = walkFunc(&ts, id)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (bs *BoltStore) PutPendingMsgID(ts *time.Time, id MessageID) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket_pending_ids)

		b_id := []byte(ts.Format(time.RFC3339))
		err := b.Put(b_id, id[:])
		return err
	})

	return err
}

func (bs *BoltStore) RemovePendingMsgID(ts *time.Time) error {
	b_id := []byte(ts.Format(time.RFC3339))

	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket_pending_ids)

		err := b.Delete(b_id)
		return err

	})

	return err
}

//	SaveTasks(id MessageID, workerId string, tasks []map[string]interface{}) error
//	LoadTasks(id MessageID) (map[string][]map[string]interface{}, error)

func (bs *BoltStore) SaveTasks(id MessageID, workerId string, tasks []map[string]interface{}) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket_tasks)
		if err != nil {
			return err
		}

		data := map[string][]map[string]interface{}{
			workerId: tasks,
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		enc.Encode(data)

		return b.Put(id[:], buf.Bytes())
	})

	return err
}

func (bs *BoltStore) LoadTasks(id MessageID) (tasks map[string][]map[string]interface{}, err error) {

	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket_tasks)
		v := b.Get(id[:])

		var result *map[string][]map[string]interface{}
		buf := bytes.NewBuffer(v)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&result)
		if err != nil {
			return err
		}

		if result != nil {
			tasks = *result
		}

		return nil
	})

	return
}

func (bs *BoltStore) decodeMsg(b []byte) (*Message, error) {
	var msg Message
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&msg)
	return &msg, err

}
