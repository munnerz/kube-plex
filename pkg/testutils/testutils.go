package testutils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"math/rand"
	"path/filepath"
	"strconv"
	"os"
)

func RandomName() string {
  letters := "abcdefghijklmnopqrstuvwxyz"
	var buffer bytes.Buffer

	for i:=0; i<16; i+=1 {
		buffer.WriteByte(letters[rand.Intn(len(letters))])
	}

	return buffer.String()
}

func RandomPath() string {
    return filepath.Join("/tmp/", RandomName())
}

func ReadJson(filename string) (result map[string][]string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}

	r := bufio.NewReader(f)
	err = json.NewDecoder(r).Decode(&result)
	return
}

func CanaryCommand() {
	if os.Getenv("TEST_CANARY") == "" {
		return
	}

	f, _ := os.Create(os.Getenv("OUTPUT_FILE"))
	w := bufio.NewWriter(f)

	cwd, _ := os.Getwd()

	json.NewEncoder(w).Encode(map[string][]string{
		"environment": os.Environ(),
		"args": os.Args,
		"cwd": []string{cwd},
	})

	w.Flush()

	exit, _ := strconv.Atoi(os.Getenv("TEST_CANARY"))
	os.Exit(exit)
}
