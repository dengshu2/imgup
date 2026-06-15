package config

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// KeychainService is the keyring service name under which imgup stores secrets.
// On macOS this maps to a generic-password item; go-keyring shells out to the
// system "security" tool, so no cgo is required.
const KeychainService = "imgup"

// ErrSecretNotFound is returned when no secret is stored for a profile.
var ErrSecretNotFound = errors.New("secret not found in keychain")

// GetSecret reads the secret access key for a profile from the OS keyring.
func GetSecret(profile string) (string, error) {
	s, err := keyring.Get(KeychainService, profile)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrSecretNotFound
	}
	return s, err
}

// SetSecret stores the secret access key for a profile in the OS keyring.
func SetSecret(profile, secret string) error {
	return keyring.Set(KeychainService, profile, secret)
}

// DeleteSecret removes a profile's secret from the OS keyring. A missing entry
// is not treated as an error.
func DeleteSecret(profile string) error {
	err := keyring.Delete(KeychainService, profile)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
