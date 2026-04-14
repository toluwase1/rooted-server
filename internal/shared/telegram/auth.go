package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// ValidateInitData verifies the Telegram Mini App init data.
// This ensures the request actually came from Telegram and hasn't been tampered with.
// See: https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func ValidateInitData(initData string, botToken string) (bool, error) {
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return false, fmt.Errorf("parsing init data: %w", err)
	}

	receivedHash := parsed.Get("hash")
	if receivedHash == "" {
		return false, fmt.Errorf("missing hash")
	}

	// Remove hash from the data
	parsed.Del("hash")

	// Sort remaining params alphabetically
	var keys []string
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build data-check-string
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, parsed.Get(k)))
	}
	dataCheckString := strings.Join(parts, "\n")

	// HMAC-SHA256 with secret key derived from bot token
	secretKey := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	hash := hmacSHA256(secretKey, []byte(dataCheckString))
	computedHash := hex.EncodeToString(hash)

	return computedHash == receivedHash, nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
