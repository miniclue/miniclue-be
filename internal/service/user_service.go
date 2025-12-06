package service

import (
	"context"
	"errors"
	"fmt"

	"app/internal/model"
	"app/internal/repository"

	"github.com/rs/zerolog"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)

type UserService interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	Get(ctx context.Context, id string) (*model.User, error)
	GetRecentLecturesWithCount(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, int, error)
	GetCourses(ctx context.Context, userID string) ([]model.Course, error)
	StoreAPIKey(ctx context.Context, userID, provider, apiKey string) error
	DeleteAPIKey(ctx context.Context, userID, provider string) error
	ListModels(ctx context.Context, userID string) ([]ProviderModels, error)
	SetModelPreference(ctx context.Context, userID, provider, modelName string, enabled bool) error
}

type userService struct {
	userRepo           repository.UserRepository
	courseRepo         repository.CourseRepository
	lectureRepo        repository.LectureRepository
	secretManagerSvc   SecretManagerService
	openAIValidator    OpenAIValidator
	geminiValidator    GeminiValidator
	anthropicValidator AnthropicValidator
	xaiValidator       XAIValidator
	deepseekValidator  DeepSeekValidator
	userLogger         zerolog.Logger
}

type ModelToggle struct {
	ID      string
	Name    string
	Enabled bool
}

// ProviderModels represents a provider and its models with enabled state.
type ProviderModels struct {
	Provider string
	Models   []ModelToggle
}

type catalogEntry struct {
	ID   string
	Name string
}

// providerOrder defines the order in which providers should be displayed
var providerOrder = []string{
	"openai",
	"gemini",
	"anthropic",
	"xai",
	"deepseek",
}

// curatedModelCatalog holds a static list of models per provider.
// These are placeholders and can be updated later without changing the API surface.
var curatedModelCatalog = map[string][]catalogEntry{
	"openai": {
		{ID: "gpt-5.1", Name: "GPT-5.1"},
		{ID: "gpt-5.1-chat-latest", Name: "GPT-5.1 chat latest"},
		{ID: "gpt-5", Name: "GPT-5"},
		{ID: "gpt-5-chat-latest", Name: "GPT-5 chat latest"},
		{ID: "gpt-5-mini", Name: "GPT-5 mini"},
		{ID: "gpt-5-nano", Name: "GPT-5 nano"},
		{ID: "gpt-4.1", Name: "GPT-4.1"},
		{ID: "gpt-4.1-mini", Name: "GPT-4.1 mini"},
		{ID: "gpt-4.1-nano", Name: "GPT-4.1 nano"},
		{ID: "gpt-4o", Name: "GPT-4o"},
		{ID: "gpt-4o-mini", Name: "GPT-4o mini"},
	},
	"gemini": {
		{ID: "gemini-3-pro-preview", Name: "Gemini 3 Pro Preview"},
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro"},
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash"},
		{ID: "gemini-2.5-flash-lite", Name: "Gemini 2.5 Flash Lite"},
	},
	"anthropic": {
		{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5"},
		{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5"},
	},
	"xai": {
		{ID: "grok-4-1-fast-reasoning", Name: "Grok 4.1 Fast (Reasoning)"},
		{ID: "grok-4-1-fast-non-reasoning", Name: "Grok 4.1 Fast (Non-reasoning)"},
	},
	"deepseek": {
		{ID: "deepseek-chat", Name: "DeepSeek-V3.2 (Non-thinking Mode)"},
		{ID: "deepseek-reasoner", Name: "DeepSeek-V3.2 (Thinking Mode)"},
	},
}

func NewUserService(userRepo repository.UserRepository, courseRepo repository.CourseRepository, lectureRepo repository.LectureRepository, secretManagerSvc SecretManagerService, openAIValidator OpenAIValidator, geminiValidator GeminiValidator, anthropicValidator AnthropicValidator, xaiValidator XAIValidator, deepseekValidator DeepSeekValidator, logger zerolog.Logger) UserService {
	return &userService{
		userRepo:           userRepo,
		courseRepo:         courseRepo,
		lectureRepo:        lectureRepo,
		secretManagerSvc:   secretManagerSvc,
		openAIValidator:    openAIValidator,
		geminiValidator:    geminiValidator,
		anthropicValidator: anthropicValidator,
		xaiValidator:       xaiValidator,
		deepseekValidator:  deepseekValidator,
		userLogger:         logger.With().Str("service", "UserService").Logger(),
	}
}

func (s *userService) Create(ctx context.Context, u *model.User) (*model.User, error) {
	// Check if user already exists first
	_, err := s.userRepo.GetUserByID(ctx, u.UserID)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		s.userLogger.Error().Err(err).Str("user_id", u.UserID).Msg("Failed to check if user exists")
		return nil, err
	}

	// Create/update user in database
	err = s.userRepo.CreateUser(ctx, u)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", u.UserID).Msg("Failed to create/update user")
		return nil, err
	}

	return u, nil
}

func (s *userService) Get(ctx context.Context, id string) (*model.User, error) {
	u, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", id).Msg("Failed to get user by ID")
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *userService) GetCourses(ctx context.Context, userID string) ([]model.Course, error) {
	courses, err := s.courseRepo.GetCoursesByUserID(ctx, userID)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get courses by user ID")
		return nil, err
	}
	return courses, nil
}

func (s *userService) GetRecentLecturesWithCount(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, int, error) {
	// Get lectures with pagination
	lectures, err := s.lectureRepo.GetLecturesByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get recent lectures by user ID")
		return nil, 0, err
	}

	// Get total count
	totalCount, err := s.lectureRepo.CountLecturesByUserID(ctx, userID)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get lecture count by user ID")
		return nil, 0, err
	}

	return lectures, totalCount, nil
}

func (s *userService) StoreAPIKey(ctx context.Context, userID, provider, apiKey string) error {
	if apiKey == "" {
		return errors.New("API key cannot be empty")
	}
	if provider == "" {
		return errors.New("provider cannot be empty")
	}

	// Validate API key before storing based on provider
	var validationErr error
	switch provider {
	case "openai":
		validationErr = s.openAIValidator.ValidateAPIKey(ctx, apiKey)
	case "gemini":
		validationErr = s.geminiValidator.ValidateAPIKey(ctx, apiKey)
	case "anthropic":
		validationErr = s.anthropicValidator.ValidateAPIKey(ctx, apiKey)
	case "xai":
		validationErr = s.xaiValidator.ValidateAPIKey(ctx, apiKey)
	case "deepseek":
		validationErr = s.deepseekValidator.ValidateAPIKey(ctx, apiKey)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	if validationErr != nil {
		s.userLogger.Error().Err(validationErr).Str("user_id", userID).Str("provider", provider).Msg("API key validation failed")
		return fmt.Errorf("invalid API key: %w", validationErr)
	}

	// Store in Secret Manager with provider-specific naming
	err := s.secretManagerSvc.StoreUserAPIKey(ctx, userID, provider, apiKey)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Str("provider", provider).Msg("Failed to store API key in Secret Manager")
		return err
	}

	// Update the flag in database
	err = s.userRepo.UpdateAPIKeyFlag(ctx, userID, provider, true)
	if err != nil {
		s.userLogger.Error().
			Err(err).
			Str("user_id", userID).
			Str("provider", provider).
			Bool("context_cancelled", ctx.Err() != nil).
			Str("error_type", fmt.Sprintf("%T", err)).
			Msg("Failed to update API key flag in database")
		return err
	}

	return nil
}

func (s *userService) DeleteAPIKey(ctx context.Context, userID, provider string) error {
	if provider == "" {
		return errors.New("provider cannot be empty")
	}

	// Delete from Secret Manager
	err := s.secretManagerSvc.DeleteUserAPIKey(ctx, userID, provider)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Str("provider", provider).Msg("Failed to delete API key from Secret Manager")
		return err
	}

	// Update the flag in database to false
	err = s.userRepo.UpdateAPIKeyFlag(ctx, userID, provider, false)
	if err != nil {
		s.userLogger.Error().
			Err(err).
			Str("user_id", userID).
			Str("provider", provider).
			Bool("context_cancelled", ctx.Err() != nil).
			Str("error_type", fmt.Sprintf("%T", err)).
			Msg("Failed to update API key flag in database")
		return err
	}

	return nil
}

func (s *userService) ListModels(ctx context.Context, userID string) ([]ProviderModels, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user for models listing")
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	var result []ProviderModels
	// Iterate in defined order to ensure consistent provider sequence
	for _, provider := range providerOrder {
		models, exists := curatedModelCatalog[provider]
		if !exists {
			continue
		}

		hasKey := user.APIKeysProvided[provider]
		if !hasKey {
			continue
		}

		var toggles []ModelToggle
		for _, m := range models {
			enabled := false
			if user.ModelPreferences != nil {
				if providerPrefs, ok := user.ModelPreferences[provider]; ok {
					enabled = providerPrefs[m.ID]
				}
			}
			toggles = append(toggles, ModelToggle{
				ID:      m.ID,
				Name:    m.Name,
				Enabled: enabled,
			})
		}

		result = append(result, ProviderModels{
			Provider: provider,
			Models:   toggles,
		})
	}

	return result, nil
}

func (s *userService) SetModelPreference(ctx context.Context, userID, provider, modelName string, enabled bool) error {
	entries, ok := curatedModelCatalog[provider]
	if !ok {
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	// Validate model exists in catalog
	found := false
	for _, m := range entries {
		if m.ID == modelName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unsupported model for provider %s: %s", provider, modelName)
	}

	if err := s.userRepo.UpdateModelPreference(ctx, userID, provider, modelName, enabled); err != nil {
		s.userLogger.Error().
			Err(err).
			Str("user_id", userID).
			Str("provider", provider).
			Str("model", modelName).
			Bool("enabled", enabled).
			Msg("Failed to update model preference")
		return err
	}

	return nil
}
