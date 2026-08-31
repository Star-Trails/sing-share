package bpf

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"unicode/utf8"
)

const (
	MessageTypeProfileContent byte  = 0x03
	ProfileFormatVersion      byte  = 0x01
	ProfileTypeLocal          int32 = 0
)

// EncodeLocal encodes a local profile using sing-box's current BPF wire format.
// Keep this in sync with experimental/libbox/profile_import.go in sing-box.
func EncodeLocal(name string, config []byte) ([]byte, error) {
	if !utf8.ValidString(name) {
		return nil, fmt.Errorf("profile name is not valid UTF-8")
	}
	if !utf8.Valid(config) {
		return nil, fmt.Errorf("configuration is not valid UTF-8")
	}

	var output bytes.Buffer
	output.WriteByte(MessageTypeProfileContent)
	output.WriteByte(ProfileFormatVersion)

	compressed := gzip.NewWriter(&output)
	if err := writeString(compressed, []byte(name)); err != nil {
		return nil, fmt.Errorf("write profile name: %w", err)
	}
	if err := binary.Write(compressed, binary.BigEndian, ProfileTypeLocal); err != nil {
		return nil, fmt.Errorf("write profile type: %w", err)
	}
	if err := writeString(compressed, config); err != nil {
		return nil, fmt.Errorf("write configuration: %w", err)
	}
	if err := compressed.Close(); err != nil {
		return nil, fmt.Errorf("finish gzip payload: %w", err)
	}

	return output.Bytes(), nil
}

func writeString(buffer *gzip.Writer, value []byte) error {
	var prefix [binary.MaxVarintLen64]byte
	length := binary.PutUvarint(prefix[:], uint64(len(value)))
	if _, err := buffer.Write(prefix[:length]); err != nil {
		return err
	}
	_, err := buffer.Write(value)
	return err
}
