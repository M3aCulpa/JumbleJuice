package encoder

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// validates ipv4 address format
func ValidateIPv4(addr string) error {
	parts := strings.Split(addr, ".")
	if len(parts) != 4 {
		return fmt.Errorf("invalid IPv4: expected 4 octets, got %d", len(parts))
	}
	for i, part := range parts {
		octet, err := strconv.Atoi(part)
		if err != nil {
			return fmt.Errorf("invalid IPv4 octet %d: %s", i, part)
		}
		if octet < 0 || octet > 255 {
			return fmt.Errorf("IPv4 octet %d out of range: %d", i, octet)
		}
	}
	return nil
}

// validates mac address format (colon or dash delimited)
func ValidateMAC(addr string) error {
	cleaned := strings.ReplaceAll(addr, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	if len(cleaned) != 12 {
		return fmt.Errorf("invalid MAC: expected 12 hex chars, got %d", len(cleaned))
	}
	if _, err := hex.DecodeString(cleaned); err != nil {
		return fmt.Errorf("invalid MAC: %w", err)
	}
	return nil
}