package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/google/uuid"
)

// OpenAIAuthConnection is a safe management DTO. Credentials are never serialized.
// AccountID owns security, concurrency and billing; ID identifies only this grant.
// A subject is meaningful only with its issuer/client and verification flag.
type OpenAIAuthConnection struct {
	Legacy                bool       `json:"legacy"`
	ID                    string     `json:"id"`
	AccountID             int64      `json:"account_id"`
	Kind                  string     `json:"kind"`
	Label                 string     `json:"label"`
	Status                string     `json:"status"`
	Issuer                string     `json:"issuer"`
	ClientID              string     `json:"client_id"`
	Subject               string     `json:"subject,omitempty"`
	IdentityVerified      bool       `json:"identity_verified"`
	Workspace             string     `json:"workspace,omitempty"`
	Resource              string     `json:"resource,omitempty"`
	Audience              string     `json:"audience,omitempty"`
	Scopes                []string   `json:"scopes"`
	Capabilities          []string   `json:"capabilities"`
	HostID                string     `json:"host_id,omitempty"`
	Revision              int64      `json:"revision"`
	Preferred             bool       `json:"preferred"`
	CooldownUntil         *time.Time `json:"cooldown_until,omitempty"`
	CredentialsCiphertext string     `json:"-"`
}

type OpenAIAuthConnectionRepository interface {
	AuthConnectionsManaged(context.Context, int64) (bool, error)
	ResolveAuthConnectionAffinity(context.Context, int64, int64, string, string) (string, error)
	UseLegacyAuthConnections(context.Context, int64) error
	ListAuthConnections(context.Context, int64) ([]*OpenAIAuthConnection, error)
	CreateAuthConnection(context.Context, *OpenAIAuthConnection) error
	CASAuthConnection(context.Context, *OpenAIAuthConnection, int64) (bool, error)
	SelectAuthConnection(context.Context, int64, string, int64) (bool, error)
	AcquireAuthConnectionLease(context.Context, int64, string, string) (bool, error)
	ReleaseAuthConnectionLease(context.Context, int64, string, string) error
}

type CreateOpenAIAuthConnectionInput struct {
	Kind   string `json:"kind" binding:"required"`
	Label  string `json:"label" binding:"required"`
	APIKey string `json:"api_key"`
}

type UpdateOpenAIAuthConnectionInput struct {
	Revision int64  `json:"revision" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

func (s *OpenAIOAuthReauthService) connectionRepository() (OpenAIAuthConnectionRepository, error) {
	if s == nil || s.repo == nil || s.accounts == nil || s.encryptor == nil {
		return nil, infraerrors.ServiceUnavailable("OPENAI_CONNECTIONS_UNAVAILABLE", "Authorization connection storage is unavailable")
	}
	r, ok := s.repo.(OpenAIAuthConnectionRepository)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("OPENAI_CONNECTIONS_UNAVAILABLE", "Authorization connection storage is unavailable")
	}
	return r, nil
}

func (s *OpenAIOAuthReauthService) connectionAccount(ctx context.Context, id int64) (*Account, error) {
	a, err := s.accounts.GetAccount(ctx, id)
	if err != nil || a == nil {
		return nil, infraerrors.NotFound("OPENAI_CONNECTION_ACCOUNT_NOT_FOUND", "Account not found")
	}
	if a.Platform != PlatformOpenAI || a.IsCredentialShadow() || (a.Type != AccountTypeOAuth && a.Type != AccountTypeAPIKey) {
		return nil, infraerrors.BadRequest("OPENAI_CONNECTION_ACCOUNT_INVALID", "Connections require an OpenAI parent account")
	}
	return a, nil
}

func (s *OpenAIOAuthReauthService) ListAuthConnections(ctx context.Context, id int64) ([]*OpenAIAuthConnection, error) {
	r, err := s.connectionRepository()
	if err != nil {
		return nil, err
	}
	if _, err = s.connectionAccount(ctx, id); err != nil {
		return nil, err
	}
	managed, err := r.ListAuthConnections(ctx, id)
	if err != nil {
		return nil, err
	}
	a, err := s.connectionAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	legacy, err := s.legacyAuthConnections(ctx, a)
	if err != nil {
		return nil, err
	}
	return append(managed, legacy...), nil
}

func (s *OpenAIOAuthReauthService) CreateAuthConnection(ctx context.Context, id int64, input CreateOpenAIAuthConnectionInput) (*OpenAIAuthConnection, error) {
	r, err := s.connectionRepository()
	if err != nil {
		return nil, err
	}
	if _, err = s.connectionAccount(ctx, id); err != nil {
		return nil, err
	}
	input.Label = strings.TrimSpace(input.Label)
	if input.Label == "" || len(input.Label) > 120 {
		return nil, infraerrors.BadRequest("OPENAI_CONNECTION_LABEL_INVALID", "Use a connection label of 1–120 bytes")
	}
	c := &OpenAIAuthConnection{ID: uuid.NewString(), AccountID: id, Kind: input.Kind, Label: input.Label, Status: "pending", Revision: 1, Scopes: []string{}, Capabilities: []string{}}
	switch input.Kind {
	case "codex", "bps", "siwc":
		if input.APIKey != "" {
			return nil, infraerrors.BadRequest("OPENAI_CONNECTION_CREDENTIAL_MISMATCH", "An OAuth connection cannot accept an API key")
		}
		c.Issuer = "https://auth.openai.com"
		if input.Kind == "siwc" {
			c.HostID = "urn:uuid:" + uuid.NewString()
			c.Resource = openai.SIWCResource
		}
	case "api_key":
		key := strings.TrimSpace(input.APIKey)
		if key == "" || len(key) > 16384 {
			return nil, infraerrors.BadRequest("OPENAI_CONNECTION_API_KEY_REQUIRED", "An API key is required")
		}
		if err = s.ensureDurableEncryption(); err != nil {
			return nil, err
		}
		raw, marshalErr := json.Marshal(map[string]any{"api_key": key})
		if marshalErr != nil {
			return nil, marshalErr
		}
		c.CredentialsCiphertext, err = s.encryptor.Encrypt(string(raw))
		if err != nil {
			return nil, errors.New("failed to encrypt connection credentials")
		}
		c.Status = "active"
		c.Capabilities = []string{"responses", "chat_completions", "embeddings"}
		c.Resource = "https://api.openai.com/v1"
	default:
		return nil, infraerrors.BadRequest("OPENAI_CONNECTION_KIND_INVALID", "Unknown authorization connection kind")
	}
	if err = r.CreateAuthConnection(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *OpenAIOAuthReauthService) authConnection(ctx context.Context, accountID int64, id string) (*OpenAIAuthConnection, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, infraerrors.BadRequest("OPENAI_CONNECTION_ID_INVALID", "Invalid connection ID")
	}
	connections, err := s.ListAuthConnections(ctx, accountID)
	if err != nil {
		return nil, err
	}
	for _, c := range connections {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, infraerrors.NotFound("OPENAI_CONNECTION_NOT_FOUND", "Authorization connection not found")
}

func (s *OpenAIOAuthReauthService) UpdateAuthConnection(ctx context.Context, accountID int64, id string, input UpdateOpenAIAuthConnectionInput) (*OpenAIAuthConnection, error) {
	c, err := s.authConnection(ctx, accountID, id)
	if err != nil {
		return nil, err
	}
	if c.Legacy {
		return nil, infraerrors.BadRequest("OPENAI_CONNECTION_LEGACY_READ_ONLY", "Use the existing account controls for a legacy registration")
	}
	if c.Revision != input.Revision {
		return nil, connectionChanged()
	}
	r, err := s.connectionRepository()
	if err != nil {
		return nil, err
	}
	var changed bool
	switch input.Action {
	case "select":
		if c.Status != "active" || len(c.Capabilities) == 0 {
			return nil, infraerrors.BadRequest("OPENAI_CONNECTION_NOT_CAPABLE", "This connection does not authorize inference")
		}
		changed, err = r.SelectAuthConnection(ctx, accountID, id, input.Revision)
	case "disable":
		c.Status = "disabled"
		changed, err = r.CASAuthConnection(ctx, c, input.Revision)
	case "enable":
		if c.Status != "disabled" || c.CredentialsCiphertext == "" {
			return nil, infraerrors.BadRequest("OPENAI_CONNECTION_NOT_AUTHORIZED", "Authorize this connection first")
		}
		c.Status = "active"
		changed, err = r.CASAuthConnection(ctx, c, input.Revision)
	case "unlink":
		// Local unbinding never claims remote revocation. Preserve registration
		// metadata for reauthorization; erase only this connection's token set.
		c.Status = "unlinked"
		c.CredentialsCiphertext = ""
		c.Capabilities = []string{}
		changed, err = r.CASAuthConnection(ctx, c, input.Revision)
	default:
		return nil, infraerrors.BadRequest("OPENAI_CONNECTION_ACTION_INVALID", "Unknown connection action")
	}
	if err != nil {
		return nil, err
	}
	if !changed {
		return nil, connectionChanged()
	}
	return s.authConnection(ctx, accountID, id)
}

func connectionChanged() error {
	return infraerrors.Conflict("OPENAI_CONNECTION_CHANGED", "Connection changed; reload before retrying")
}

// An explicit connection marker must never become an implicit Codex grant.
func validateOpenAIConnectionKind(a *Account) error {
	if a == nil || a.Platform != PlatformOpenAI {
		return nil
	}
	kind := a.GetExtraString("openai_connection_kind")
	if kind != "" && a.AuthConnectionID == "" {
		return infraerrors.BadRequest("OPENAI_CONNECTION_NOT_CAPABLE", "Explicit connections require a bound authorization record")
	}
	if kind == "codex" && a.GetCredential("client_id") != openai.ClientID || kind == "bps" && a.GetCredential("client_id") != openai.ExcelClientID || kind == "siwc" && (!a.IsSIWCEnabled() || a.GetCredential("client_id") == "" || a.GetCredential("client_id") == openai.SIWCDynamicClientID) || kind == "api_key" && a.Type != AccountTypeAPIKey {
		return infraerrors.BadRequest("OPENAI_CONNECTION_CLIENT_MISMATCH", "Connection credentials do not match the authorization client")
	}
	switch kind {
	case "", "codex", "bps", "siwc", "api_key":
		return nil
	default:
		return infraerrors.BadRequest("OPENAI_CONNECTION_NOT_CAPABLE", "Unknown or identity-only authorization cannot be used for inference")
	}
}
