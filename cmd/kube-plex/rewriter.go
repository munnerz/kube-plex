package main

import "strings"

type rewriter struct {
	pmsInternalAddress string
}

// rewriteEnv rewrites environment variables to be passed to the transcoder
func (r rewriter) Env(in []string) []string {
	return in
}

// Args rewrites argument list to use kube-plex specific values
func (r rewriter) Args(args []string) []string {
	out := make([]string, len(args))
	copy(out, args)
	for i, v := range args {
		switch v {
		case "-progressurl", "-manifest_name", "-segment_list":
			out[i+1] = strings.Replace(out[i+1], "http://127.0.0.1:32400", r.pmsInternalAddress, 1)
		case "-loglevel", "-loglevel_plex":
			out[i+1] = "debug"
		}
	}
	return out
}
