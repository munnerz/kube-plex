package kubeplex

import (
	"os"
	"strings"
)

var pmsInternalAddress = os.Getenv("PMS_INTERNAL_ADDRESS")

func RewriteArgs(in []string) {
	for i, v := range in {
		switch v {
		case "-progressurl", "-manifest_name", "-segment_list":
			in[i+1] = strings.Replace(in[i+1], "http://127.0.0.1:32400", pmsInternalAddress, 1)
		case "-loglevel", "-loglevel_plex":
			in[i+1] = "debug"
		}
	}
}
