package api

import (
	"chat/user"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func UserChannelsHandler(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["user"]

	list, err := user.GetChannels(username)
	if err != nil {
		handleError(err, w)
		return
	}
	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		handleError(err, w)
		return
	}

}

func UsersHandler(w http.ResponseWriter, r *http.Request) {

	list, err := user.List()
	if err != nil {
		handleError(err, w)
		return
	}
	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		handleError(err, w)
		return
	}
}

func handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf(`{"err": "%s"}`, err.Error())))
}
