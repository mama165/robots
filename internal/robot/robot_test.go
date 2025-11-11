package robot

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSplitSecret(t *testing.T) {
	ass := assert.New(t)
	input := "Hidden beneath the old oak tree, golden coins patiently await discovery."
	expected := []string{"Hidden", "beneath", "the", "old", "oak", "tree,", "golden", "coins", "patiently", "await", "discovery."}

	secretManager := SecretManager{Config: Config{}}
	result := secretManager.SplitSecret(input)
	ass.Equal(expected, result)
}

func TestCreateRobots(t *testing.T) {
	ass := assert.New(t)
	words := []string{"a", "b", "c", "d"}
	secretManager := SecretManager{Config: Config{NbrOfRobots: 3, BufferSize: 10}}
	robots := secretManager.CreateRobots(words)

	ass.Equal(3, len(robots))
	total := 0
	for _, r := range robots {
		ass.Equal(10, cap(r.Inbox))
		total += len(r.Words)
	}
	ass.Equal(total, len(words))
}

func TestIsSecretRebuilt(t *testing.T) {
	ass := assert.New(t)
	completed := IsSecretCompleted([]string{"a", "b", "c."}, ".")
	ass.True(completed)
}

func TestExchangeAtLeastOneMessage(t *testing.T) {
	ass := assert.New(t)
	secretManager := SecretManager{Config: Config{
		MaxAttempts: 2,
	}}
	r1 := Robot{
		ID:    0,
		Words: []string{"alpha", "gamma"},
		Inbox: make(chan Inbox, 10),
	}
	r2 := Robot{
		ID:    1,
		Words: []string{"beta"},
		Inbox: make(chan Inbox, 10),
	}
	sent := secretManager.ExchangeMessage(r1, r2)

	ass.GreaterOrEqual(sent, 1)
	ass.LessOrEqual(sent, 2)
}

func TestWriteSecret(t *testing.T) {
	ass := assert.New(t)
	tmpFile := "test_secret.txt"
	secretManager := SecretManager{Config: Config{OutputFile: tmpFile}}

	secret := "Hidden beneath the oak"
	err := secretManager.WriteSecret(secret)
	ass.NoError(err)

	// Checking the file existence
	_, err = os.Stat(tmpFile)
	ass.NoError(err)

	data, err := os.ReadFile(tmpFile)
	ass.NoError(err)
	ass.Equal(secret, string(data))

	err = os.Remove(tmpFile)
	ass.NoError(err)
}
