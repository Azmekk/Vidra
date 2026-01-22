package services

import (
	"context"
	"sync"

	"github.com/Azmekk/Vidra/backend/gen/database"
)

type SettingsDTO struct {
	ProxyUrl          string `json:"proxyUrl"`
	DefaultReEncode   bool   `json:"defaultReEncode"`
	DefaultVideoCodec string `json:"defaultVideoCodec"`
	DefaultAudioCodec string `json:"defaultAudioCodec"`
	DefaultCrf        int    `json:"defaultCrf"`
	Theme             string `json:"theme"`
}

type SettingsService struct {
	queries *database.Queries
	cache   *SettingsDTO
	mu      sync.RWMutex
}

func NewSettingsService(queries *database.Queries) *SettingsService {
	return &SettingsService{
		queries: queries,
	}
}

func mapSettingToDTO(s database.Setting) SettingsDTO {
	return SettingsDTO{
		ProxyUrl:          s.ProxyUrl,
		DefaultReEncode:   s.DefaultReEncode,
		DefaultVideoCodec: s.DefaultVideoCodec,
		DefaultAudioCodec: s.DefaultAudioCodec,
		DefaultCrf:        int(s.DefaultCrf),
		Theme:             s.Theme,
	}
}

func (s *SettingsService) GetSettings(ctx context.Context) (SettingsDTO, error) {
	s.mu.RLock()
	if s.cache != nil {
		cached := *s.cache
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	setting, err := s.queries.GetSettings(ctx)
	if err != nil {
		return SettingsDTO{}, err
	}

	dto := mapSettingToDTO(setting)

	s.mu.Lock()
	s.cache = &dto
	s.mu.Unlock()

	return dto, nil
}

func (s *SettingsService) UpdateSettings(ctx context.Context, dto SettingsDTO) (SettingsDTO, error) {
	setting, err := s.queries.UpdateSettings(ctx, database.UpdateSettingsParams{
		ProxyUrl:          dto.ProxyUrl,
		DefaultReEncode:   dto.DefaultReEncode,
		DefaultVideoCodec: dto.DefaultVideoCodec,
		DefaultAudioCodec: dto.DefaultAudioCodec,
		DefaultCrf:        int32(dto.DefaultCrf),
		Theme:             dto.Theme,
	})
	if err != nil {
		return SettingsDTO{}, err
	}

	result := mapSettingToDTO(setting)

	s.mu.Lock()
	s.cache = &result
	s.mu.Unlock()

	return result, nil
}

func (s *SettingsService) GetProxyURL(ctx context.Context) string {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return ""
	}
	return settings.ProxyUrl
}
