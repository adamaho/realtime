package realtime

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mattbaird/jsonpatch"
)

type Session struct {
	Clients []*Client
	Data    json.RawMessage
}

func newSession(data json.RawMessage) Session {
	return Session{Clients: make([]*Client, 0), Data: data}
}

func (s *Session) addClient(ch *chan []byte) uuid.UUID {
	clientID := uuid.New()
	client := Client{clientID: clientID, Channel: ch}
	s.Clients = append(s.Clients, &client)
	return clientID
}

func (s *Session) removeClient(clientID uuid.UUID) int {
	newClients := make([]*Client, 0)

	for _, c := range s.Clients {
		if c.clientID != clientID {
			newClients = append(newClients, c)
		}
	}

	s.Clients = newClients
	return len(s.Clients)
}

type Client struct {
	clientID uuid.UUID
	Channel  *chan []byte
}

type Realtime struct {
	sessions map[string]*Session
}

func New() Realtime {
	return Realtime{sessions: make(map[string]*Session, 0)}
}

func (rt *Realtime) getOrCreateSession(sessionID string, data json.RawMessage) *Session {
	session, ok := rt.sessions[sessionID]

	if !ok {
		session := newSession(data)
		rt.sessions[sessionID] = &session
		return &session
	}

	return session
}

func (rt *Realtime) removeSession(sessionID string) {
	delete(rt.sessions, sessionID)
}

// TODO: allow options and make it so that stream can be passed as a boolean instead of assuming the header
// TODO: specify storage method for previous data (in memory or redis)
// TODO: write some good comments on all of this stuff
// TODO: figure out how to write some tests
// TODO: tighten up the api and make sure things that are public should be public
func (rt *Realtime) Stream(w http.ResponseWriter, r *http.Request, data json.RawMessage, sessionID string, stream bool) {
	ctx := r.Context()

	if !stream {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := make(chan []byte, 10)
	defer close(ch)

	session := rt.getOrCreateSession(sessionID, data)

	clientID := session.addClient(&ch)
	w.Header().Set("Content-Type", "application/json+ndjsonpatch")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", data)
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			count := session.removeClient(clientID)
			if count == 0 {
				rt.removeSession(sessionID)
			}
			return
		case value := <-ch:
			if len(value) > 0 {
				fmt.Fprintf(w, "%s\n", value)
			}
			flusher.Flush()
		}
	}
}

func (rt *Realtime) CreatePatch(target json.RawMessage, sessionID string) (json.RawMessage, error) {
	session, ok := rt.sessions[sessionID]

	if !ok {
		return nil, fmt.Errorf("Failed to get session for sessionID: %s", sessionID)
	}

	patch, _ := jsonpatch.CreatePatch(session.Data, target)
	patchJson, err := json.Marshal(patch)

	if err != nil {
		log.Print("Failed to marshal json for patch")
		return nil, err
	}

	session.Data = target

	return patchJson, nil
}

func (rt *Realtime) PublishMsg(msg json.RawMessage, sessionID string) error {
	session, ok := rt.sessions[sessionID]

	if !ok {
		return fmt.Errorf("Failed to find session with sessionID: %s", sessionID)
	}

	for _, client := range session.Clients {
		*client.Channel <- msg
	}

	return nil
}
