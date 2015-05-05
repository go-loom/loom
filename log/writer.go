package log

import (
	"fmt"
	"github.com/koding/logging"
	"io"
	"sync"
)

type SyncWriterHandler struct {
	*logging.BaseHandler
	w        io.Writer
	Colorize bool
	sync.Mutex
}

func NewSyncWriterHandler(w io.Writer) *SyncWriterHandler {
	return &SyncWriterHandler{
		BaseHandler: logging.NewBaseHandler(),
		w:           w,
	}
}

func (b *SyncWriterHandler) Handle(rec *logging.Record) {
	message := b.BaseHandler.FilterAndFormat(rec)
	if message == "" {
		return
	}
	b.Lock()
	defer b.Unlock()
	if b.Colorize {
		b.w.Write([]byte(fmt.Sprintf("\033[%dm", logging.LevelColors[rec.Level])))
	}
	fmt.Fprint(b.w, message)
	if b.Colorize {
		b.w.Write([]byte("\033[0m")) // reset color
	}
}

func (b *SyncWriterHandler) Close() {
}
