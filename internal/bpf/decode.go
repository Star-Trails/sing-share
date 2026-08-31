package bpf

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf8"
)

// LocalProfile is the verified content of a local-profile BPF message.
type LocalProfile struct {
	Name   string
	Config []byte
}

// DecodeLocal verifies and decodes sing-box's current local-profile BPF format.
func DecodeLocal(data []byte) (*LocalProfile, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("BPF message is truncated")
	}
	if data[0] != MessageTypeProfileContent {
		return nil, fmt.Errorf("unsupported BPF message type 0x%02x", data[0])
	}
	if data[1] != ProfileFormatVersion {
		return nil, fmt.Errorf("unsupported BPF profile version 0x%02x", data[1])
	}

	compressed, err := gzip.NewReader(bytes.NewReader(data[2:]))
	if err != nil {
		return nil, fmt.Errorf("open gzip payload: %w", err)
	}
	defer compressed.Close()

	payload := bufio.NewReader(compressed)
	name, err := readString(payload)
	if err != nil {
		return nil, fmt.Errorf("read profile name: %w", err)
	}

	var profileType int32
	if err := binary.Read(payload, binary.BigEndian, &profileType); err != nil {
		return nil, fmt.Errorf("read profile type: %w", err)
	}
	if profileType != ProfileTypeLocal {
		return nil, fmt.Errorf("unsupported profile type %d", profileType)
	}

	config, err := readBytes(payload)
	if err != nil {
		return nil, fmt.Errorf("read configuration: %w", err)
	}
	if _, err := payload.ReadByte(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("unexpected data after local profile")
		}
		return nil, fmt.Errorf("finish gzip payload: %w", err)
	}
	if !utf8.ValidString(name) {
		return nil, fmt.Errorf("profile name is not valid UTF-8")
	}
	if !utf8.Valid(config) {
		return nil, fmt.Errorf("configuration is not valid UTF-8")
	}

	return &LocalProfile{Name: name, Config: config}, nil
}

func readString(reader *bufio.Reader) (string, error) {
	value, err := readBytes(reader)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func readBytes(reader *bufio.Reader) ([]byte, error) {
	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, err
	}
	if length > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("length %d exceeds platform capacity", length)
	}
	value := make([]byte, int(length))
	if _, err := io.ReadFull(reader, value); err != nil {
		return nil, err
	}
	return value, nil
}
