package auth

import (
	"errors"
	"fmt"

	"github.com/spritz-finance/spritz-cli/internal/config"
)

// AccessMode identifies the principal that will own a device grant.
// User access is the existing Spritz account flow. Developer access is a
// workspace/integrator flow and must never silently fall back to user access.
type AccessMode string

const (
	AccessUser      AccessMode = "user"
	AccessDeveloper AccessMode = "developer"
)

var ErrDeveloperAccessNotEnabled = errors.New("developer access device grants are not enabled")

type DeveloperAccessNotEnabledError struct {
	ConsoleURL string
}

func (e *DeveloperAccessNotEnabledError) Error() string {
	return fmt.Sprintf(
		"%s by this Spritz platform deployment; a human workspace administrator must enroll and accept the Developer Terms at %s (an AI agent must not accept terms or own a Production credential)",
		ErrDeveloperAccessNotEnabled.Error(),
		e.ConsoleURL,
	)
}

func (e *DeveloperAccessNotEnabledError) Unwrap() error {
	return ErrDeveloperAccessNotEnabled
}

func ParseAccessMode(value string) (AccessMode, error) {
	mode := AccessMode(value)
	switch mode {
	case AccessUser, AccessDeveloper:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid access mode %q: expected user or developer", value)
	}
}

func developerAccessNotEnabled() error {
	return &DeveloperAccessNotEnabledError{ConsoleURL: config.DeveloperConsoleURL()}
}
