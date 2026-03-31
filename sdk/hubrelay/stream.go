package hubrelay

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Stream struct {
	response *http.Response
	scanner  *bufio.Scanner
	chunk    StreamChunk
	result   CommandResult
	err      error
	done     bool
}

func newStream(response *http.Response) *Stream {
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &Stream{
		response: response,
		scanner:  scanner,
	}
}

func (s *Stream) Next() bool {
	if s == nil || s.done {
		return false
	}

	eventName := ""
	data := ""
	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			if eventName == "" {
				continue
			}
			return s.handleEvent(eventName, data)
		}
		switch {
		case strings.HasPrefix(line, "event: "):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		}
	}

	if err := s.scanner.Err(); err != nil {
		s.err = err
	}
	s.done = true
	return false
}

func (s *Stream) handleEvent(eventName, data string) bool {
	switch eventName {
	case "chunk":
		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			s.err = err
			s.done = true
			return false
		}
		s.chunk = chunk
		return true
	case "done":
		_ = json.Unmarshal([]byte(data), &s.result)
		s.done = true
		return false
	case "error":
		_ = json.Unmarshal([]byte(data), &s.result)
		if s.result.Message != "" {
			s.err = errors.New(s.result.Message)
		} else {
			s.err = errors.New("hubrelay stream returned error")
		}
		s.done = true
		return false
	default:
		return false
	}
}

func (s *Stream) Chunk() StreamChunk {
	return s.chunk
}

func (s *Stream) Result() (CommandResult, error) {
	return s.result, s.err
}

func (s *Stream) Close() error {
	if s == nil || s.response == nil || s.response.Body == nil {
		return nil
	}
	return s.response.Body.Close()
}
