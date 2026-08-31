package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/wailsapp/wails/v3/pkg/application"

	"sing-share/internal/bpf"
)

const configExtension = ".json"

// ShareProfile contains the generated BPF and only the metadata needed by the UI.
type ShareProfile struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
	Size     int    `json:"size"`
}

// ShareService owns filesystem access and sing-box BPF handling.
type ShareService struct {
	startupPath string
	startupOnce sync.Once
}

func NewShareService(startupPath string) *ShareService {
	return &ShareService{startupPath: startupPath}
}

func (s *ShareService) OpenConfig() (*ShareProfile, error) {
	path, err := application.Get().Dialog.OpenFile().
		SetTitle("Choose sing-box configuration").
		AddFilter("sing-box JSON configuration", "*.json").
		PromptForSingleSelection()
	if err != nil {
		return nil, fmt.Errorf("open configuration dialog: %w", err)
	}
	if path == "" {
		return nil, nil
	}
	return s.LoadConfig(path)
}

func (s *ShareService) LoadConfig(path string) (*ShareProfile, error) {
	if !strings.EqualFold(filepath.Ext(path), configExtension) {
		return nil, fmt.Errorf("choose a .json configuration file")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect configuration: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("configuration path is not a regular file")
	}

	config, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read configuration: %w", err)
	}
	defer clear(config)
	if !utf8.Valid(config) {
		return nil, fmt.Errorf("configuration is not valid UTF-8")
	}
	if !json.Valid(config) {
		return nil, fmt.Errorf("configuration is not valid JSON")
	}

	name := deriveProfileName(path)
	data, err := bpf.EncodeLocal(name, config)
	if err != nil {
		return nil, fmt.Errorf("encode BPF profile: %w", err)
	}
	return &ShareProfile{
		Name:     name,
		Filename: filepath.Base(path),
		Data:     data,
		Size:     len(data),
	}, nil
}

// StartupProfile returns the command-line profile at most once.
func (s *ShareService) StartupProfile() (*ShareProfile, error) {
	var profile *ShareProfile
	var loadErr error
	s.startupOnce.Do(func() {
		if s.startupPath != "" {
			profile, loadErr = s.LoadConfig(s.startupPath)
		}
	})
	return profile, loadErr
}

// SaveBPF writes the exact bytes currently held by the frontend. The bool is
// false only when the native save dialog is cancelled.
func (s *ShareService) SaveBPF(data []byte, suggestedName string) (bool, error) {
	if _, err := bpf.DecodeLocal(data); err != nil {
		return false, fmt.Errorf("refusing to save invalid BPF: %w", err)
	}

	filename := safeBPFFilename(suggestedName)
	path, err := application.Get().Dialog.SaveFile().
		SetFilename(filename).
		AddFilter("sing-box profile", "*.bpf").
		PromptForSingleSelection()
	if err != nil {
		return false, fmt.Errorf("open save dialog: %w", err)
	}
	if path == "" {
		return false, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".bpf") {
		path += ".bpf"
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return false, fmt.Errorf("write BPF profile: %w", err)
	}
	return true, nil
}

func deriveProfileName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func safeBPFFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "" || name == "." {
		name = "profile"
	}
	if !strings.EqualFold(filepath.Ext(name), ".bpf") {
		name += ".bpf"
	}
	return name
}
