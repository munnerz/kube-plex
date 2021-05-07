// Package logger provides a logger which will write log entries to Plex Media Server
package logger

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func TestPlexLogger_getURL(t *testing.T) {
	l := &PlexLogger{}
	want, _ := url.Parse("http://127.0.0.1:32400/log")

	if got := l.getURL(); !reflect.DeepEqual(got, *want) {
		t.Errorf("PlexLogger.getURL() when empty = %v, want %v", got, *want)
	}

	want, _ = url.Parse("http://test:1234/log")
	l = &PlexLogger{plexURL: want}

	got := l.getURL()
	if !reflect.DeepEqual(got, *want) {
		t.Errorf("PlexLogger.getURL() when plexURL set = %v, want %v", got, *want)
	}
}

func TestPlexLogger_send(t *testing.T) {
	type args struct {
		level int
		msg   string
		kvs   []interface{}
	}
	tests := []struct {
		name      string
		keyValues map[string]interface{}
		token     string
		args      args
		level     int
		message   string
	}{
		{"plain message is sent", map[string]interface{}{}, "PTOKEN", args{level: PlexLogInfo, msg: "test"}, PlexLogInfo, "test"},
		{"level is respected", map[string]interface{}{}, "PTOKEN", args{level: PlexLogError, msg: "test"}, PlexLogError, "test"},
		{"message is sent when there is no token", map[string]interface{}{}, "", args{level: PlexLogInfo, msg: "test"}, PlexLogInfo, "test"},
		{"struct kv is sent", map[string]interface{}{"key": "value", "key2": "value2"}, "PTOKEN", args{level: PlexLogInfo, msg: "test"}, PlexLogInfo, "test [key:value key2:value2]"},
		{"message args are sent", map[string]interface{}{}, "PTOKEN", args{level: PlexLogInfo, msg: "test", kvs: []interface{}{"key", 1, "key2", "value2"}}, PlexLogInfo, "test [key:1 key2:value2]"},
		{"message args and struct kv are sent", map[string]interface{}{"skey": 2, "skey2": "svalue2"}, "PTOKEN", args{level: PlexLogInfo, msg: "test", kvs: []interface{}{"key", 1, "key2", "value2"}}, PlexLogInfo, "test [skey:2 skey2:svalue2 key:1 key2:value2]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("PlexLogger.send() sent %v want %v", r.Method, http.MethodGet)
				}

				msg := r.URL.Query().Get("message")
				if msg != tt.message {
					t.Errorf("PlexLogger.send() received message = %v want = %v", msg, tt.message)
				}

				token := r.Header.Get("X-Plex-Token")
				if token != tt.token {
					t.Errorf("PlexLogger.send() token header mismatch, got = %v want = %v", token, tt.token)
				}

				level := r.URL.Query().Get("level")
				l, err := strconv.Atoi(level)
				if err != nil {
					t.Errorf("PlexLogger.send() sent invalid level = %v: %v", level, err)
					return
				}

				if l != tt.level {
					t.Errorf("PlexLogger.send() received with level = %v want = %v", level, tt.level)
				}
			}))
			defer ts.Close()

			u, _ := url.Parse(ts.URL)

			l := &PlexLogger{
				plexURL:   u,
				plexToken: tt.token,
				keyValues: tt.keyValues,
				name:      "test",
			}
			l.send(tt.args.level, tt.args.msg, tt.args.kvs...)
		})
	}
}
