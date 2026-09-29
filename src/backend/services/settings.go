package services

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/services/encoding"
)

type Settings struct {
	ProxyURL                string           `json:"proxyUrl"`
	Theme                   string           `json:"theme"`
	PreferCompatibleFormats bool             `json:"preferCompatibleFormats"`
	DefaultEncoding         encoding.Request `json:"defaultEncoding"`
	KeepOriginal            bool             `json:"keepOriginal"`
	CacheSize               int              `json:"cacheSize"`
	MaxConcurrentDownloads  int              `json:"maxConcurrentDownloads"`
	MaxConcurrentEncodes    int              `json:"maxConcurrentEncodes"`
}

func (s Settings) Validate(caps *encoding.Capabilities) error {
	if !slices.Contains([]string{"light", "dark", "system"}, s.Theme) {
		return fmt.Errorf("theme must be light, dark or system")
	}
	if s.CacheSize < 0 || s.CacheSize > 5000 {
		return fmt.Errorf("cache size must be between 0 and 5000")
	}
	if s.MaxConcurrentDownloads < 1 || s.MaxConcurrentDownloads > 16 {
		return fmt.Errorf("concurrent downloads must be between 1 and 16")
	}
	if s.MaxConcurrentEncodes < 1 || s.MaxConcurrentEncodes > 8 {
		return fmt.Errorf("concurrent encodes must be between 1 and 8")
	}
	return s.DefaultEncoding.Validate(caps)
}

type SettingsService struct {
	queries   *database.Queries
	mu        sync.RWMutex
	cache     *Settings
	listeners []func(Settings)
}

func NewSettingsService(queries *database.Queries) *SettingsService {
	return &SettingsService{queries: queries}
}

// OnChange registers a callback invoked after settings are saved.
func (s *SettingsService) OnChange(fn func(Settings)) {
	s.listeners = append(s.listeners, fn)
}

func (s *SettingsService) Get(ctx context.Context) (Settings, error) {
	s.mu.RLock()
	if s.cache != nil {
		cached := *s.cache
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	row, err := s.queries.GetSettings(ctx)
	if err != nil {
		return Settings{}, err
	}
	settings := fromRow(row)
	s.mu.Lock()
	s.cache = &settings
	s.mu.Unlock()
	return settings, nil
}

// MustGet returns settings, falling back to defaults if the database fails.
func (s *SettingsService) MustGet(ctx context.Context) Settings {
	settings, err := s.Get(ctx)
	if err != nil {
		return Settings{Theme: "system", PreferCompatibleFormats: true, KeepOriginal: true, CacheSize: 100,
			MaxConcurrentDownloads: 3, MaxConcurrentEncodes: 1, DefaultEncoding: encoding.Request{Goal: encoding.GoalCompatible}}
	}
	return settings
}

func (s *SettingsService) Update(ctx context.Context, settings Settings) (Settings, error) {
	defaultEncoding, err := json.Marshal(settings.DefaultEncoding)
	if err != nil {
		return Settings{}, err
	}
	row, err := s.queries.UpdateSettings(ctx, database.UpdateSettingsParams{
		ProxyUrl:                settings.ProxyURL,
		Theme:                   settings.Theme,
		PreferCompatibleFormats: settings.PreferCompatibleFormats,
		DefaultEncoding:         string(defaultEncoding),
		KeepOriginal:            settings.KeepOriginal,
		CacheSize:               int64(settings.CacheSize),
		MaxConcurrentDownloads:  int64(settings.MaxConcurrentDownloads),
		MaxConcurrentEncodes:    int64(settings.MaxConcurrentEncodes),
	})
	if err != nil {
		return Settings{}, err
	}

	result := fromRow(row)
	s.mu.Lock()
	s.cache = &result
	s.mu.Unlock()
	for _, fn := range s.listeners {
		fn(result)
	}
	return result, nil
}

func fromRow(row database.Setting) Settings {
	var req encoding.Request
	if err := json.Unmarshal([]byte(row.DefaultEncoding), &req); err != nil || req.Goal == "" {
		req = encoding.Request{Goal: encoding.GoalCompatible}
	}
	return Settings{
		ProxyURL:                row.ProxyUrl,
		Theme:                   row.Theme,
		PreferCompatibleFormats: row.PreferCompatibleFormats,
		DefaultEncoding:         req,
		KeepOriginal:            row.KeepOriginal,
		CacheSize:               int(row.CacheSize),
		MaxConcurrentDownloads:  int(row.MaxConcurrentDownloads),
		MaxConcurrentEncodes:    int(row.MaxConcurrentEncodes),
	}
}
