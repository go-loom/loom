package server

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/loom.v1/config"
	"io/ioutil"
	"net/http"
)

type Json map[string]interface{}

func PushHandler(w http.ResponseWriter, r *http.Request) {
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

	msg, err := broker.PushMessage(queueName, &job)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	send(w, http.StatusCreated, msg.JSON())

	return
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		send(w, http.StatusMethodNotAllowed, Json{"error": "Not supported method"})
		return
	}

	queueName := mux.Vars(r)["queue"]
	id := mux.Vars(r)["id"]

	var msgId MessageID
	copy(msgId[:], id)

	msg, err := broker.GetMessage(queueName, msgId)
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

/*
func DeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Method != "DELETE" {
		send(w, http.StatusMethodNotAllowed, Json{"error": "Not supported method"})
		return
	}

	queueName := mux.Vars(r)["queue"]
	id := mux.Vars(r)["id"]

	var msgId MessageID
	copy(msgId[:], id)

	err := broker.FinishMessage(queueName, msgId)
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
