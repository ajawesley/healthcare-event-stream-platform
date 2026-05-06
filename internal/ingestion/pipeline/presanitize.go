package pipeline

import (
	"context"
	"strings"

	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

// PreSanitize removes framing artifacts that interfere with detection and parsing.
// It does NOT perform format-specific normalization.
func PreSanitize(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	// No context available here → use background context
	ctx := context.Background()
	log := observability.WithTrace(ctx)

	s := string(payload)

	log.Debug("presanitize_start",
		zap.Int("raw_len", len(payload)),
		zap.String("raw_preview", preview(s)),
	)

	// 1. Trim framing quotes + whitespace
	before := s
	s = strings.Trim(s, "\" \t\r\n")
	if s != before {
		log.Debug("presanitize_trimmed_quotes_whitespace")
	}

	// -------------------------
	// HL7: CR handling
	// -------------------------

	// Reduce quadruple-escaped CR → double-escaped CR
	before = s
	s = strings.ReplaceAll(s, `\\\\r`, `\\r`)
	if s != before {
		log.Debug("presanitize_reduce_quadruple_cr")
	}

	// Reduce double-escaped CR → single-escaped CR
	before = s
	s = strings.ReplaceAll(s, `\\r`, `\r`)
	if s != before {
		log.Debug("presanitize_reduce_double_cr")
	}

	// Convert literal "\r" → real CR
	before = s
	s = strings.ReplaceAll(s, `\r`, "\r")
	if s != before {
		log.Debug("presanitize_literal_cr_to_real")
	}

	// -------------------------
	// X12: LF handling
	// -------------------------

	// Reduce quadruple-escaped LF → double-escaped LF
	before = s
	s = strings.ReplaceAll(s, `\\\\n`, `\\n`)
	if s != before {
		log.Debug("presanitize_reduce_quadruple_lf")
	}

	// Reduce double-escaped LF → single-escaped LF
	before = s
	s = strings.ReplaceAll(s, `\\n`, `\n`)
	if s != before {
		log.Debug("presanitize_reduce_double_lf")
	}

	// Convert literal "\n" → real LF
	before = s
	s = strings.ReplaceAll(s, `\n`, "\n")
	if s != before {
		log.Debug("presanitize_literal_lf_to_real")
	}

	// -------------------------
	// Normalize line endings
	// -------------------------

	// Normalize CRLF → CR
	before = s
	s = strings.ReplaceAll(s, "\r\n", "\r")
	if s != before {
		log.Debug("presanitize_normalize_crlf_to_cr")
	}

	// Normalize LF-only → CR (X12 does not use LF)
	before = s
	s = strings.ReplaceAll(s, "\n", "\r")
	if s != before {
		log.Debug("presanitize_normalize_lf_to_cr")
	}

	// Remove UTF‑8 BOM if present
	before = s
	s = strings.TrimPrefix(s, "\uFEFF")
	if s != before {
		log.Debug("presanitize_removed_bom")
	}

	log.Debug("presanitize_complete",
		zap.Int("final_len", len(s)),
		zap.String("final_preview", preview(s)),
	)

	return []byte(s)
}

// preview returns a safe, truncated preview for logs.
func preview(s string) string {
	if len(s) <= 200 {
		return s
	}
	return s[:200] + "...(truncated)"
}
