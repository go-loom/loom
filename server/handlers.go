package server

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func PushHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	queueName := ps.ByName("queue")
	queueValue, err := ioutil.ReadAll(r.Body)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	msg, err := broker.PushMessage(queueName, queueValue)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	send(w, http.StatusCreated, Json{
		"id":      string(msg.ID[:]),
		"value":   string(msg.Value.([]byte)),
		"created": msg.Created,
	})

	return
}

func PopHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	queueName := ps.ByName("queue")
	logger.Debug("Pop", "qname", queueName)

	clientGone := w.(http.CloseNotifier).CloseNotify()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/plain")
	//w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fmt.Fprintf(w, "# ~1KB of junk to force browsers to start rendering immediately: \n")
	io.WriteString(w, strings.Repeat("# xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", 13))

	for {

		select {
		case <-clientGone:
			logger.Info("Consumer " + r.RemoteAddr + " disconnected ")
			return
		default:
			msg := broker.PopMessage(queueName)
			value := msg.Value.([]byte)
			logger.Debug("Dequeue", "id", string(msg.ID[:]), "msg", string(value))

			w.Write(value)
			w.(http.Flusher).Flush()

		}
	}

}

func DeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	queueName := ps.ByName("queue")
	id := ps.ByName("id")

	var msgId MessageID
	copy(msgId[:], id)

	err := broker.FinishMessage(queueName, msgId)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type Json map[string]interface{}

func send(w http.ResponseWriter, code int, data Json) error {
	bytes, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(bytes)

	return nil
}
