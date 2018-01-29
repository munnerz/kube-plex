package kubeplex

import (
	"github.com/stretchr/testify/assert"
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"github.com/munnerz/kube-plex/pkg/testutils"
	"os"
	"testing"
)

func TestCommandExecution(t *testing.T){
	testutils.CanaryCommand()
}

func TestRunPlexTranscodeJob(t *testing.T) {
	filename := testutils.RandomPath()

	env := []string{"TEST_CANARY=0", "OUTPUT_FILE=" + filename}
	args := []string{
		os.Args[0], "-test.run=TestCommandExecution", "--", "my", "arguments",
	}

	ptj := GeneratePlexTranscodeJob(args, env)

	state, stderr := RunPlexTranscodeJob(&ptj)
	assert.Equal(t, ptjv1.PlexTranscodeStateCompleted, state, "Should be completed")
	assert.Equal(t, "", stderr, "Stderr should be empty")

	output, err := testutils.ReadJson(filename)
	assert.Equal(t, nil, err, "err should be empty")
	assert.Equal(t, env, output["environment"], "environment variables should match")
	assert.Equal(t, args, output["args"], "args should match")
}

func TestRunPlexTranscodeJobCommandFails(t *testing.T) {
	filename := testutils.RandomPath()

	// TEST_CANARY=1 -- return failure
	env := []string{"TEST_CANARY=1", "OUTPUT_FILE=" + filename}
	args := []string{
		os.Args[0], "-test.run=TestCommandExecution", "--", "my", "arguments",
	}

	ptj := GeneratePlexTranscodeJob(args, env)

	state, stderr := RunPlexTranscodeJob(&ptj)
	assert.Equal(t, ptjv1.PlexTranscodeStateFailed, state, "Should be failed")
	assert.Equal(t, "", stderr, "Stderr should be empty")

	output, err := testutils.ReadJson(filename)
	assert.Equal(t, nil, err, "err should be empty")
	assert.Equal(t, env, output["environment"], "environment variables should match")
	assert.Equal(t, args, output["args"], "args should match")
}
