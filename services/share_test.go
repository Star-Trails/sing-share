package services

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"sing-share/internal/bpf"
)

func TestLoadConfigPreservesExactBytes(t *testing.T) {
	t.Parallel()

	config := []byte("{\n  \"log\": { \"level\": \"warn\" },\n  \"备注\": \"本地\"\n}\n")
	path := filepath.Join(t.TempDir(), "singapore-node.json")
	if err := os.WriteFile(path, config, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	profile, err := NewShareService("").LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if profile.Name != "singapore-node" {
		t.Fatalf("Name = %q, want %q", profile.Name, "singapore-node")
	}
	if profile.Filename != "singapore-node.json" {
		t.Fatalf("Filename = %q", profile.Filename)
	}
	if profile.Size != len(profile.Data) {
		t.Fatalf("Size = %d, len(Data) = %d", profile.Size, len(profile.Data))
	}

	decoded, err := bpf.DecodeLocal(profile.Data)
	if err != nil {
		t.Fatalf("DecodeLocal() error = %v", err)
	}
	if !bytes.Equal(decoded.Config, config) {
		t.Fatalf("config changed:\n got %q\nwant %q", decoded.Config, config)
	}
}

func TestLoadConfigRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	for name, fixture := range map[string]struct {
		filename string
		data     []byte
	}{
		"extension": {filename: "config.txt", data: []byte("{}")},
		"JSON":      {filename: "config.json", data: []byte("not JSON")},
		"UTF-8":     {filename: "config.json", data: []byte{0xff}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), fixture.filename)
			if err := os.WriteFile(path, fixture.data, 0o600); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			if _, err := NewShareService("").LoadConfig(path); err == nil {
				t.Fatal("LoadConfig() accepted invalid input")
			}
		})
	}
}
