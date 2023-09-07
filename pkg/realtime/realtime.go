package realtime

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mattbaird/jsonpatch"
)

// Realtime supports http handlers. The first is `Stream` which supports both a plain json API response
// and a streaming jsonpatch response.
//
// `sessions` allow for colocating data and clients that are subscribe to the stream and are able to receive patches.
type Realtime struct {
	sessions map[string]*Session
}

// Creates a new instance of realtime
func New() Realtime {
	return Realtime{sessions: make(map[string]*Session, 0)}
}

// Creates a new session
func (rt *Realtime) createSession(sessionID string) *Session {
	session := newSession()
	rt.sessions[sessionID] = &session
	return &session
}

// Removes a session
func (rt *Realtime) removeSession(sessionID string) {
	delete(rt.sessions, sessionID)
}

// TODO: allow data to be an interface and dont force it to be a raw json message
// TODO: allow options and make it so that stream can be passed as a boolean instead of assuming the header
// TODO: specify storage method for previous data (in memory or redis)
// TODO: create standard response type with data struct
// TODO: write some good comments on all of this stuff

// Handles creating a stream and channel. If header
// is not set, the raw json message is returned to the user like a standard REST api.
func (rt *Realtime) Stream(w http.ResponseWriter, r *http.Request, d json.RawMessage, sessionID string, stream bool) {
	ctx := r.Context()

	if !stream {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(d)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := make(chan []byte, 10)
	defer close(ch)

	// check if there is already a session for the current sessionID
	session, ok := rt.sessions[sessionID]

	if !ok {
		session = rt.createSession(sessionID)
	}

	clientID := session.addClient(&ch)
	w.Header().Set("Content-Type", "application/json+ndjsonpatch")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", d)
	flusher.Flush()

	session.Data = d

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

// Creates a new json patch
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

// Creates a new json patch
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

// Holds all of the active client connections
type Session struct {
	Clients []*Client
	Data    json.RawMessage
}

// Creates a new Session
func newSession() Session {
	return Session{Clients: make([]*Client, 0), Data: nil}
}

// Adds a new client to the Session
func (s *Session) addClient(ch *chan []byte) uuid.UUID {
	clientID := uuid.New()
	client := Client{clientID: clientID, Channel: ch}
	s.Clients = append(s.Clients, &client)
	return clientID
}

// Removes a client from the Session
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

// Represents a single client connection
type Client struct {
	clientID uuid.UUID
	Channel  *chan []byte
}

// The response structure of a realtime api
type data struct {
	Data interface{} `json:"data"`
}
