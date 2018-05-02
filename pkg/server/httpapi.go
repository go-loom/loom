package server

import (
	"encoding/json"
	"github.com/go-loom/loom/pkg/config"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type Json map[string]interface{}

type httpApiHandler struct {
	broker *Broker
}

func (h *httpApiHandler) PushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.ListHandler(w, r)
		return
	}
	if r.Method != "POST" {
		send(w, http.StatusMethodNotAllowed, Json{"error": "Not supported method"})
		return
	}

	queueName := mux.Vars(r)["queue"]

	queueValue, err := ioutil.ReadAll(r.Body)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	var job config.Job
	if err := json.Unmarshal(queueValue, &job); err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	msg, err := h.broker.PushMessage(queueName, &job)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	send(w, http.StatusCreated, msg.JSON())

	return
}

func (h *httpApiHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		send(w, http.StatusMethodNotAllowed, Json{"error": "Not supported method"})
		return
	}

	queueName := mux.Vars(r)["queue"]
	id := mux.Vars(r)["id"]

	var msgId MessageID
	copy(msgId[:], id)

	msg, err := h.broker.GetMessage(queueName, msgId)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}
	if msg == nil {
		send(w, http.StatusNotFound, Json{"error": "NotFound"})
		return
	}

	msgJson := msg.JSON()
	send(w, http.StatusCreated, msgJson)

	return
}

func (h *httpApiHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		send(w, http.StatusMethodNotAllowed, Json{"error": "Not supported method"})
		return
	}

	queueName := mux.Vars(r)["queue"]

	topic, ok := h.broker.Topics[queueName]
	if !ok {
		send(w, http.StatusNotFound, Json{"error": "the queue doesn't exist"})
		return
	}

	queue, ok := topic.Queue.(*LQueue)
	if !ok {
		send(w, http.StatusInternalServerError, Json{"error": "topic.Queue is wrong"})
		return
	}
	items := queue.List()

	messages := make([]Json, 0, len(items))

	for _, i := range items {
		m, ok := i.(*Message)
		if ok {
			messages = append(messages, m.JSON())
		}
	}

	send(w, http.StatusOK, Json{"queues": messages, "len": len(messages)})
	return
}

/*
func (h *httpApiHandler) DeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Method != "DELETE" {
		send(w, http.StatusMethodNotAllowed, Json{"error": "Not supported method"})
		return
	}

	queueName := mux.Vars(r)["queue"]
	id := mux.Vars(r)["id"]

	var msgId MessageID
	copy(msgId[:], id)

	err := h.broker.FinishMessage(queueName, msgId)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
*/

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
