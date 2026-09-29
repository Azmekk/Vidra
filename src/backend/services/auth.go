package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidSetupCode   = errors.New("invalid setup code")
	ErrSetupDone          = errors.New("setup has already been completed")
	ErrUnauthenticated    = errors.New("not authenticated")
	ErrMalformedAPIToken  = errors.New("token does not start with " + apiTokenPrefix)
	ErrUnknownAPIToken    = errors.New("token is unknown or was deleted")
	usernameRe            = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,32}$`)
)

const (
	RememberDuration = 30 * 24 * time.Hour
	SessionDuration  = 12 * time.Hour
	touchInterval    = time.Hour
	apiTokenPrefix   = "vidra_"
)

type User = database.User

type Session struct {
	Token     string
	ExpiresAt time.Time
	Remember  bool
}

type AuthService struct {
	queries *database.Queries

	mu        sync.Mutex
	setupCode string
}

func NewAuthService(queries *database.Queries) *AuthService {
	return &AuthService{queries: queries}
}

// Init creates a one-time setup code when no user exists yet.
func (s *AuthService) Init(ctx context.Context) error {
	required, err := s.SetupRequired(ctx)
	if err != nil || !required {
		return err
	}
	s.mu.Lock()
	code := make([]byte, 4)
	_, _ = rand.Read(code)
	s.setupCode = strings.ToUpper(hex.EncodeToString(code))
	s.mu.Unlock()
	log.Printf("🔑 No account exists yet. Open Vidra and use setup code: %s\n", s.setupCode)
	return nil
}

func (s *AuthService) SetupRequired(ctx context.Context) (bool, error) {
	n, err := s.queries.CountUsers(ctx)
	return n == 0, err
}

func (s *AuthService) Setup(ctx context.Context, code, username, password, userAgent string) (User, Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	required, err := s.SetupRequired(ctx)
	if err != nil {
		return User{}, Session{}, err
	}
	if !required || s.setupCode == "" {
		return User{}, Session{}, ErrSetupDone
	}
	if subtle.ConstantTimeCompare([]byte(strings.ToUpper(strings.TrimSpace(code))), []byte(s.setupCode)) != 1 {
		return User{}, Session{}, ErrInvalidSetupCode
	}
	if err := ValidateCredentials(username, password); err != nil {
		return User{}, Session{}, err
	}
	user, err := s.queries.CreateUser(ctx, database.CreateUserParams{
		ID: NewID(), Username: username, PasswordHash: hashPassword(password),
	})
	if err != nil {
		return User{}, Session{}, err
	}
	s.setupCode = ""
	session, err := s.createSession(ctx, user.ID, true, userAgent)
	return user, session, err
}

func (s *AuthService) Login(ctx context.Context, username, password string, remember bool, userAgent string) (User, Session, error) {
	user, err := s.queries.GetUserByUsername(ctx, username)
	if err != nil {
		verifyPassword(dummyHash, password)
		return User{}, Session{}, ErrInvalidCredentials
	}
	if !verifyPassword(user.PasswordHash, password) {
		return User{}, Session{}, ErrInvalidCredentials
	}
	session, err := s.createSession(ctx, user.ID, remember, userAgent)
	return user, session, err
}

// Authenticate validates a session token, sliding its expiry forward.
func (s *AuthService) Authenticate(ctx context.Context, token string) (User, *Session, error) {
	now := time.Now()
	row, err := s.queries.GetSession(ctx, database.GetSessionParams{TokenHash: hashToken(token), Now: now.Unix()})
	if err != nil {
		return User{}, nil, ErrUnauthenticated
	}
	if now.Sub(time.Unix(row.Session.LastSeenAt, 0)) < touchInterval {
		return row.User, nil, nil
	}
	expires := now.Add(sessionLength(row.Session.Remember))
	if err := s.queries.TouchSession(ctx, database.TouchSessionParams{
		TokenHash: row.Session.TokenHash, LastSeenAt: now.Unix(), ExpiresAt: expires.Unix(),
	}); err != nil {
		return row.User, nil, nil
	}
	return row.User, &Session{Token: token, ExpiresAt: expires, Remember: row.Session.Remember}, nil
}

func (s *AuthService) AuthenticateAPIToken(ctx context.Context, token string) (User, error) {
	if !strings.HasPrefix(token, apiTokenPrefix) {
		return User{}, ErrMalformedAPIToken
	}
	hash := hashToken(token)
	user, err := s.queries.GetUserByAPIToken(ctx, hash)
	if err != nil {
		return User{}, ErrUnknownAPIToken
	}
	_ = s.queries.TouchAPIToken(ctx, hash)
	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.queries.DeleteSession(ctx, hashToken(token))
}

// ChangePassword updates the password and signs out every other session.
func (s *AuthService) ChangePassword(ctx context.Context, user User, currentToken, current, next string) error {
	if !verifyPassword(user.PasswordHash, current) {
		return ErrInvalidCredentials
	}
	if err := ValidateCredentials(user.Username, next); err != nil {
		return err
	}
	if err := s.queries.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{ID: user.ID, PasswordHash: hashPassword(next)}); err != nil {
		return err
	}
	return s.queries.DeleteOtherSessions(ctx, database.DeleteOtherSessionsParams{UserID: user.ID, TokenHash: hashToken(currentToken)})
}

func (s *AuthService) CreateAPIToken(ctx context.Context, userID, name string) (database.ApiToken, string, error) {
	token := apiTokenPrefix + randomToken(32)
	row, err := s.queries.CreateAPIToken(ctx, database.CreateAPITokenParams{
		ID: NewID(), UserID: userID, Name: name, TokenHash: hashToken(token),
	})
	return row, token, err
}

func (s *AuthService) ListAPITokens(ctx context.Context, userID string) ([]database.ApiToken, error) {
	return s.queries.ListAPITokens(ctx, userID)
}

func (s *AuthService) DeleteAPIToken(ctx context.Context, userID, id string) error {
	return s.queries.DeleteAPIToken(ctx, database.DeleteAPITokenParams{ID: id, UserID: userID})
}

// PruneSessions removes expired sessions once a day.
func (s *AuthService) PruneSessions(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		if err := s.queries.DeleteExpiredSessions(ctx, time.Now().Unix()); err != nil {
			slog.Warn("failed to prune sessions", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *AuthService) createSession(ctx context.Context, userID string, remember bool, userAgent string) (Session, error) {
	token := randomToken(32)
	now := time.Now()
	expires := now.Add(sessionLength(remember))
	err := s.queries.CreateSession(ctx, database.CreateSessionParams{
		TokenHash: hashToken(token), UserID: userID, ExpiresAt: expires.Unix(), LastSeenAt: now.Unix(),
		Remember: remember, UserAgent: truncate(userAgent, 255),
	})
	return Session{Token: token, ExpiresAt: expires, Remember: remember}, err
}

func ValidateCredentials(username, password string) error {
	if !usernameRe.MatchString(username) {
		return fmt.Errorf("username must be 3-32 letters, digits, dots, dashes or underscores")
	}
	if len(password) < 8 || len(password) > 256 {
		return fmt.Errorf("password must be between 8 and 256 characters")
	}
	return nil
}

func sessionLength(remember bool) time.Duration {
	if remember {
		return RememberDuration
	}
	return SessionDuration
}

const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 2
	argonKeyLen  = 32
)

var dummyHash = hashPassword("vidra-timing-equaliser")

func hashPassword(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))
}

func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var memory, iterations uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.ToValidUTF8(s[:n], "")
}
