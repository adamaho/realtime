// TODO: implement redis cache for previous data for cases where there is a lot of data
// TODO: figure out how to write some tests

package realtime

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mattbaird/jsonpatch"
)

type session struct {
	Clients []*client
	Data    json.RawMessage
}

func newSession(data json.RawMessage) session {
	return session{Clients: make([]*client, 0), Data: data}
}

func (s *session) addClient(ch *chan []byte) uuid.UUID {
	clientID := uuid.New()
	client := client{clientID: clientID, Channel: ch}
	s.Clients = append(s.Clients, &client)
	return clientID
}

func (s *session) removeClient(clientID uuid.UUID) int {
	newClients := make([]*client, 0)

	for _, c := range s.Clients {
		if c.clientID != clientID {
			newClients = append(newClients, c)
		}
	}

	s.Clients = newClients
	return len(s.Clients)
}

type client struct {
	clientID uuid.UUID
	Channel  *chan []byte
}

type Realtime struct {
	sessions map[string]*session
}

// New creates a new instance of realtime
func New() Realtime {
	return Realtime{sessions: make(map[string]*session, 0)}
}

func (rt *Realtime) getOrCreateSession(sessionID string, data json.RawMessage) *session {
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

type responseOptions struct {
	bufferSize          int
	streamRequestHeader string
}

// ResponseOptions returns a struct of options to be passed to the Response instance method on Realtime
func ResponseOptions(options ...func(*responseOptions)) *responseOptions {
	opts := &responseOptions{
		bufferSize:          10,
		streamRequestHeader: "x-stream",
	}
	for _, o := range options {
		o(opts)
	}
	return opts
}

// WithBufferSize returns a function that sets the bufferSize for the Response with the given size
//
// Default: 10
func WithBufferSize(size int) func(*responseOptions) {
	return func(opts *responseOptions) {
		opts.bufferSize = size
	}
}

// WithStreamRequestHeader returns a function that sets the streamRequestHeader for the Response with the given header.
// When `true` is passed as the value for the header, Response will return a stream of jsonpatch updates when the data changes
//
// Default: "x-stream"
func WithStreamRequestHeader(header string) func(*responseOptions) {
	return func(opts *responseOptions) {
		opts.streamRequestHeader = header
	}
}

// Response takes the provided data, sessionID, and options to produce an api response.
// If the ResponseOptions streamRequestHeader is set to true the provided data will be returned, the connection to the client
// will remain open, and subsequent updates to the data via the SendMessage method will be send to the client in the form of non-deliminated
// jsonpatch's.
// If the streamRequestHeader is falsy, the data will be immediately returned and the client connection will be closed. This mimics the
// traditional json api approach.
func (rt *Realtime) Response(w http.ResponseWriter, r *http.Request, data json.RawMessage, sessionID string, options *responseOptions) error {
	ctx := r.Context()

	stream := r.Header.Get(options.streamRequestHeader)
	if stream != "true" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return nil
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("failed to get http.Flusher")
	}

	ch := make(chan []byte, options.bufferSize)
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
			return nil
		case value := <-ch:
			if len(value) > 0 {
				fmt.Fprintf(w, "%s\n", value)
			}
			flusher.Flush()
		}
	}
}

func (rt *Realtime) createPatch(target json.RawMessage, sessionID string) (json.RawMessage, error) {
	session, ok := rt.sessions[sessionID]

	if !ok {
		return nil, fmt.Errorf("failed to get session for sessionID: %s. Are you sure you have a Response for that sessionID?", sessionID)
	}

	patch, err := jsonpatch.CreatePatch(session.Data, target)
	if err != nil {
		return nil, err
	}

	patchJson, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	session.Data = target

	return patchJson, nil
}

// Send message creates a new json patch from the diff of the provided target json and the current
// data in memory that is stored under the provided sessionID.
// This is typically called after performing a mutation (PUT, PATCH, POST, DELETE) to send the patch to all connected clients
// for the provided sessionID
func (rt *Realtime) SendMessage(target json.RawMessage, sessionID string) error {
	session, ok := rt.sessions[sessionID]

	if !ok {
		return fmt.Errorf("failed to get session for sessionID: %s. Are you sure you have a Response for that sessionID?", sessionID)
	}

	patch, err := rt.createPatch(target, sessionID)
	if err != nil {
		return err
	}

	if len(patch) != 0 {
		for _, client := range session.Clients {
			*client.Channel <- patch
		}
	}

	return nil
}
