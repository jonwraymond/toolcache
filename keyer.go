package toolcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// Keyer derives cache keys from tool input.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Determinism: identical inputs must yield stable keys.
// - Errors: invalid inputs should return a descriptive error.
type Keyer interface {
	Key(toolID string, input any) (string, error)
}

type DefaultKeyer struct{}

func NewDefaultKeyer() *DefaultKeyer {
	return &DefaultKeyer{}
}

func (k *DefaultKeyer) Key(toolID string, input any) (string, error) {
	canonical, err := canonicalJSON(input)
	if err != nil {
		return "", fmt.Errorf("toolcache: failed to canonicalize input: %w", err)
	}

	hash := sha256.Sum256(canonical)
	hashHex := hex.EncodeToString(hash[:8])

	return fmt.Sprintf("toolcache:%s:%s", toolID, hashHex), nil
}

func canonicalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeCanonical(buf *bytes.Buffer, v any) error {
	switch val := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case float64:
		buf.WriteString(fmt.Sprintf("%v", val))
	case int:
		buf.WriteString(fmt.Sprintf("%d", val))
	case int64:
		buf.WriteString(fmt.Sprintf("%d", val))
	case string:
		writeJSONString(buf, val)
	case []any:
		buf.WriteByte('[')
		for i, elem := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonical(buf, elem); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeJSONString(buf, k)
			buf.WriteByte(':')
			if err := writeCanonical(buf, val[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	return nil
}

func writeJSONString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if r < 0x20 {
				buf.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else {
				buf.WriteRune(r)
			}
		}
	}
	buf.WriteByte('"')
}
