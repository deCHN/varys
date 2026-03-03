package secret

import (
	"testing"
)

func TestSecretLifecycle(t *testing.T) {
	// Keyring might not be available in headless CI environments
	// We'll try to set a test key and see if it works
	testAccount := "test_api_key"
	testValue := "sk-test-12345"

	err := SetSecret(testAccount, testValue)
	if err != nil {
		t.Skipf("Keyring not available or access denied: %v", err)
	}

	val, err := GetSecret(testAccount)
	if err != nil {
		t.Errorf("Failed to get secret: %v", err)
	}
	if val != testValue {
		t.Errorf("Expected %s, got %s", testValue, val)
	}

	err = DeleteSecret(testAccount)
	if err != nil {
		t.Errorf("Failed to delete secret: %v", err)
	}

	// Verify deletion
	val, err = GetSecret(testAccount)
	if err != nil {
		t.Errorf("Error during post-delete check: %v", err)
	}
	if val != "" {
		t.Errorf("Secret should be empty after deletion, got %s", val)
	}
}
