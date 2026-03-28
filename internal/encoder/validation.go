package encoder

import (
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

// validates mac address format
func ValidateMAC(addr string) error {
	cleaned := strings.ReplaceAll(addr, ":", "")
	if len(cleaned) != 12 {
		return fmt.Errorf("invalid MAC: expected 12 hex chars, got %d", len(cleaned))
	}
	for i, c := range cleaned {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return fmt.Errorf("invalid MAC: non-hex character at position %d", i)
		}
	}
	return nil
}