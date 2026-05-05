package pipeline

import "strings"

// PreSanitize removes framing artifacts that interfere with detection and parsing.
// It does NOT perform format-specific normalization.
func PreSanitize(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	s := string(payload)

	// 1. Trim framing quotes + whitespace
	s = strings.Trim(s, "\" \t\r\n")

	// -------------------------
	// HL7: CR handling
	// -------------------------

	// Reduce quadruple-escaped CR → double-escaped CR
	// "\\\\r" → "\\r"
	s = strings.ReplaceAll(s, `\\\\r`, `\\r`)

	// Reduce double-escaped CR → single-escaped CR
	// "\\r" → "\r"
	s = strings.ReplaceAll(s, `\\r`, `\r`)

	// Convert literal "\r" → real CR
	s = strings.ReplaceAll(s, `\r`, "\r")

	// -------------------------
	// X12: LF handling
	// -------------------------

	// Reduce quadruple-escaped LF → double-escaped LF
	// "\\\\n" → "\\n"
	s = strings.ReplaceAll(s, `\\\\n`, `\\n`)

	// Reduce double-escaped LF → single-escaped LF
	// "\\n" → "\n"
	s = strings.ReplaceAll(s, `\\n`, `\n`)

	// Convert literal "\n" → real LF
	s = strings.ReplaceAll(s, `\n`, "\n")

	// -------------------------
	// Normalize line endings
	// -------------------------

	// Normalize CRLF → CR
	s = strings.ReplaceAll(s, "\r\n", "\r")

	// Normalize LF-only → CR (X12 does not use LF)
	s = strings.ReplaceAll(s, "\n", "\r")

	// 7. Remove UTF‑8 BOM if present
	s = strings.TrimPrefix(s, "\uFEFF")

	return []byte(s)
}
