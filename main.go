package main

import (
	"encoding/json"
	"fmt"
	"github.com/stianeikeland/go-rpio/v4"
	"log"
	"net/http"
	"regexp"
)

type AtmHandler struct {
	s *AtmService
}

type AtmService struct {
}

type AtmAction struct {
	Action     string `json:"action"`
	CheckingId string `json:"checking_id"`
	Item       string `json:"item"`
}

const (
	Open = "open"
)

func (h *AtmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && regexp.MustCompile(`execute`).MatchString(r.URL.Path):
		h.Execute(w, r)
		return
	default:
		http.NotFound(w, r)
		return
	}
}

func (h *AtmHandler) Execute(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var p AtmAction
	err := decoder.Decode(&p)
	if err != nil {
		WriteStatusWithMessageHandler(w, http.StatusInternalServerError, "Could not decode JSON")
		return
	} else {
		err := h.s.PerformAction(p)
		if err != nil {
			WriteStatusWithMessageHandler(w, http.StatusInternalServerError, "Could not perform action")
			return
		}
		log.Printf("Action performed: %v\n", p)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *AtmService) PerformAction(action AtmAction) error {
	log.Printf("Action: %v\n", action)
	if action.Action == Open {
		err := rpio.Open()
		if err != nil {
			log.Printf("Error: %v\n", err)
			return err
		}
		pin := rpio.Pin(17)
		pin.Output()
		return nil
	}
	return fmt.Errorf("unknown action: %s", action.Action)
}

func WriteStatusWithMessageHandler(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(message))
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", &AtmHandler{
		s: &AtmService{},
	})
	log.Println("Server running on :3001")
	err := http.ListenAndServe(":3001", mux)
	if err != nil {
		log.Fatal(err)
	}
}
