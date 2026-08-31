package bpf

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"strings"
	"testing"
)

func TestEncodeLocalMatchesSingBoxWireLayout(t *testing.T) {
	t.Parallel()

	name := "新加坡"
	config := []byte("{\n  \"outbounds\": [{\"tag\": \"日本\"}]\n}\n")
	encoded, err := EncodeLocal(name, config)
	if err != nil {
		t.Fatalf("EncodeLocal() error = %v", err)
	}
	if len(encoded) < 2 || encoded[0] != 0x03 || encoded[1] != 0x01 {
		t.Fatalf("header = % x, want 03 01", encoded[:min(2, len(encoded))])
	}

	compressed, err := gzip.NewReader(bytes.NewReader(encoded[2:]))
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer compressed.Close()
	payload := bufio.NewReader(compressed)

	nameLength, err := binary.ReadUvarint(payload)
	if err != nil {
		t.Fatalf("read name length: %v", err)
	}
	if nameLength != uint64(len([]byte(name))) {
		t.Fatalf("name byte length = %d, want %d", nameLength, len([]byte(name)))
	}
	nameBytes := make([]byte, nameLength)
	if _, err := io.ReadFull(payload, nameBytes); err != nil {
		t.Fatalf("read name: %v", err)
	}
	if string(nameBytes) != name {
		t.Fatalf("name = %q, want %q", nameBytes, name)
	}

	var profileType int32
	if err := binary.Read(payload, binary.BigEndian, &profileType); err != nil {
		t.Fatalf("read profile type: %v", err)
	}
	if profileType != 0 {
		t.Fatalf("profile type = %d, want 0", profileType)
	}

	configLength, err := binary.ReadUvarint(payload)
	if err != nil {
		t.Fatalf("read config length: %v", err)
	}
	if configLength != uint64(len(config)) {
		t.Fatalf("config byte length = %d, want %d", configLength, len(config))
	}
	configBytes := make([]byte, configLength)
	if _, err := io.ReadFull(payload, configBytes); err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !bytes.Equal(configBytes, config) {
		t.Fatalf("config bytes changed:\n got %q\nwant %q", configBytes, config)
	}
	if _, err := payload.ReadByte(); err != io.EOF {
		t.Fatalf("payload has trailing data: %v", err)
	}
}

func TestEncodeLocalRoundTripUsesUTF8ByteLengths(t *testing.T) {
	t.Parallel()

	name := strings.Repeat("界", 50)
	config := []byte("{\"备注\":\"秘密🔐\",\"value\":1}")
	encoded, err := EncodeLocal(name, config)
	if err != nil {
		t.Fatalf("EncodeLocal() error = %v", err)
	}
	decoded, err := DecodeLocal(encoded)
	if err != nil {
		t.Fatalf("DecodeLocal() error = %v", err)
	}
	if decoded.Name != name {
		t.Fatalf("name = %q, want %q", decoded.Name, name)
	}
	if !bytes.Equal(decoded.Config, config) {
		t.Fatalf("config = %q, want %q", decoded.Config, config)
	}
}

func TestEncodeLocalRejectsInvalidUTF8(t *testing.T) {
	t.Parallel()

	if _, err := EncodeLocal("profile", []byte{0xff}); err == nil {
		t.Fatal("EncodeLocal() accepted invalid configuration UTF-8")
	}
	if _, err := EncodeLocal(string([]byte{0xff}), []byte("{}")); err == nil {
		t.Fatal("EncodeLocal() accepted invalid profile-name UTF-8")
	}
}

func TestDecodeLocalRejectsNonCurrentOrNonLocalMessages(t *testing.T) {
	t.Parallel()

	valid, err := EncodeLocal("profile", []byte("{}"))
	if err != nil {
		t.Fatalf("EncodeLocal() error = %v", err)
	}

	wrongType := bytes.Clone(valid)
	wrongType[0] = 0x02
	wrongVersion := bytes.Clone(valid)
	wrongVersion[1] = 0x00

	for name, data := range map[string][]byte{
		"truncated":    {0x03},
		"message type": wrongType,
		"version":      wrongVersion,
		"gzip payload": {0x03, 0x01, 0x00},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeLocal(data); err == nil {
				t.Fatal("DecodeLocal() accepted invalid BPF")
			}
		})
	}
}
