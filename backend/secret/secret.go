package secret

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the unique identifier for Varys in the OS keychain
	ServiceName = "com.varys.app"

	// KeyAccountOpenAI is the account name for OpenAI API Key
	KeyAccountOpenAI = "openai_api_key"
	
	// KeyAccountTavily is the account name for Tavily API Key
	KeyAccountTavily = "tavily_api_key"
)

// Internal interface for storage to allow mocking
type secretStore interface {
	Get(service, account string) (string, error)
	Set(service, account, value string) error
	Delete(service, account string) error
}

type osKeyring struct{}

func (osKeyring) Get(s, a string) (string, error)    { return keyring.Get(s, a) }
func (osKeyring) Set(s, a, v string) error          { return keyring.Set(s, a, v) }
func (osKeyring) Delete(s, a string) error          { return keyring.Delete(s, a) }

var (
	// currentStore can be swapped in tests
	currentStore secretStore = osKeyring{}
)

// UseMockStore switches the backend to an in-memory storage for testing
func UseMockStore() {
	currentStore = &mockStore{data: make(map[string]string)}
}

// ResetStore restores the OS keyring as the backend
func ResetStore() {
	currentStore = osKeyring{}
}

type mockStore struct {
	data map[string]string
}

func (m *mockStore) Get(s, a string) (string, error) {
	val, ok := m.data[s+":"+a]
	if !ok {
		return "", keyring.ErrNotFound
	}
	return val, nil
}

func (m *mockStore) Set(s, a, v string) error {
	m.data[s+":"+a] = v
	return nil
}

func (m *mockStore) Delete(s, a string) error {
	delete(m.data, s+":"+a)
	return nil
}

// SetSecret stores a sensitive value in the OS secure storage
func SetSecret(account, value string) error {
	if value == "" {
		return nil
	}
	return currentStore.Set(ServiceName, account, value)
}

// GetSecret retrieves a sensitive value from the OS secure storage
func GetSecret(account string) (string, error) {
	val, err := currentStore.Get(ServiceName, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}
	return val, err
}

// DeleteSecret removes a secret from the OS secure storage
func DeleteSecret(account string) error {
	err := currentStore.Delete(ServiceName, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
