CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    proxy_url TEXT NOT NULL DEFAULT '',
    default_re_encode BOOLEAN NOT NULL DEFAULT true,
    default_video_codec TEXT NOT NULL DEFAULT 'libx264',
    default_audio_codec TEXT NOT NULL DEFAULT 'aac',
    default_crf INTEGER NOT NULL DEFAULT 23,
    theme TEXT NOT NULL DEFAULT 'system',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

INSERT INTO settings (id) VALUES (1) ON CONFLICT DO NOTHING;

CREATE TRIGGER update_settings_updated_at
    BEFORE UPDATE ON settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
