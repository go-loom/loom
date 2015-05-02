package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
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
		"value":   string(msg.Value),
		"created": msg.Created,
		"state":   msg.State,
	})

	return
}

func GetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	queueName := ps.ByName("queue")
	id := ps.ByName("id")

	var msgId MessageID
	copy(msgId[:], id)

	msg, err := broker.GetMessage(queueName, msgId)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	send(w, http.StatusCreated, Json{
		"id":      string(msg.ID[:]),
		"value":   string(msg.Value),
		"created": msg.Created,
		"state":   msg.State,
	})

	return
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
