// Package logger provides a logger which will write log entries to Plex Media Server
package logger

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-logr/logr"
)

// Log levels known by Plex
const (
	PLEX_LOG_LEVEL_ERROR = iota
	PLEX_LOG_LEVEL_WARNING
	PLEX_LOG_LEVEL_INFO
	PLEX_LOG_LEVEL_DEBUG
	PLEX_LOG_LEVEL_VERBOSE
)

// NewPlexLogger returns a PlexLogger instance that has URL preset
//
// URL should be the base url for Plex Media Server, `/log` path and relevant
// query parameters will be added
func NewPlexLogger(token, plexurl string) (*PlexLogger, error) {
	u, err := url.Parse(plexurl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url: %v", err)
	}

	u.Path = strings.TrimSuffix(u.Path, "/") + "/log"

	q := u.Query()
	if q.Get("source") == "" {
		q.Add("source", "KubePlex")
	}
	u.RawQuery = q.Encode()

	return &PlexLogger{
		plexURL:   u,
		plexToken: token,
	}, nil
}

// PlexLogger is a single instance of Plex which is used for logging
type PlexLogger struct {
	plexURL   *url.URL // Plex url, includes plex source. (http://127.0.0.1:32400/?source=Transcoder)
	plexToken string   // Plex token for authentication

	name      string
	keyValues map[string]interface{}
}

// Enabled tests whether this Logger is enabled.
// For now we assume that the logger is always enabled
func (*PlexLogger) Enabled() bool {
	return true
}

// Info level logs are written directly to Plex
func (l *PlexLogger) Info(msg string, kvs ...interface{}) {
	l.send(PLEX_LOG_LEVEL_INFO, msg, kvs...)
}

// Error logs will have the error message passed as error key
func (l *PlexLogger) Error(err error, msg string, kvs ...interface{}) {
	if err != nil {
		kvs = append(kvs, "error", err)
	}
	l.send(PLEX_LOG_LEVEL_ERROR, msg, kvs...)
}

// We only support single verbosity
func (l *PlexLogger) V(_ int) logr.Logger {
	return l
}

// WithName adds an element to the logger name
func (l *PlexLogger) WithName(name string) logr.Logger {
	return &PlexLogger{
		plexURL:   l.plexURL,
		plexToken: l.plexToken,
		name:      l.name + "." + name,
		keyValues: l.keyValues,
	}
}

// WithValues adds key value pairs to the logger
func (l *PlexLogger) WithValues(kvs ...interface{}) logr.Logger {
	newMap := make(map[string]interface{}, len(l.keyValues)+len(kvs)/2)
	for k, v := range l.keyValues {
		newMap[k] = v
	}
	for i := 0; i < len(kvs); i += 2 {
		newMap[kvs[i].(string)] = kvs[i+1]
	}

	return &PlexLogger{
		plexURL:   l.plexURL,
		plexToken: l.getPlexToken(),
		name:      l.name,
		keyValues: newMap,
	}
}

// send message to PMS. Wrap all key value pairs to a text string since Plex has
// no concept of metadata other than message level.
//
// The request includes Plex token if it's available through the environment
func (l *PlexLogger) send(level int, msg string, kvs ...interface{}) {
	kvmsg := []string{}
	for k, v := range l.keyValues {
		kvmsg = append(kvmsg, fmt.Sprintf("%s:%+v", k, v))
	}
	for i := 0; i < len(kvs); i += 2 {
		kvmsg = append(kvmsg, fmt.Sprintf("%s:%+v", kvs[i], kvs[i+1]))
	}

	if len(kvmsg) > 0 {
		msg = fmt.Sprintf("%s %+v", msg, kvmsg)
	}

	u := l.getURL()
	q := u.Query()
	q.Set("level", fmt.Sprintf("%d", level))
	q.Set("message", msg)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		// We have an error, but no place to report it. Bail out!
		return
	}

	plexToken := l.getPlexToken()
	if plexToken != "" {
		req.Header.Add("X-Plex-Token", plexToken)
	}
	req.Header.Add("User-Agent", "PlexLogger")

	// Ignore results
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR! %v", err)
	}
}

// getURL returns either the set URL or the default URL if unset
func (l *PlexLogger) getURL() url.URL {
	if l.plexURL == nil {
		u, _ := url.Parse("http://127.0.0.1:32400/log?source=KubePlex")
		return *u
	}
	return *l.plexURL
}

// getPlexToken returns the plex token from struct if it exists or falls back to
// X_PLEX_TOKEN environment variable
func (l *PlexLogger) getPlexToken() string {
	if l.plexToken != "" {
		return l.plexToken
	}
	return os.Getenv("X_PLEX_TOKEN")
}