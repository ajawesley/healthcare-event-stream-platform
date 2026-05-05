package pipeline

import "bytes"

// PreSanitize removes framing artifacts that interfere with detection and parsing.
// It does NOT perform format-specific normalization.
func PreSanitize(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	// 1. Remove framing quotes and whitespace
	clean := bytes.Trim(payload, "\" \t\r\n")

	// 2. Convert literal '\' 'r' into actual CR (Unicode U+000D)
	//    This is CRITICAL for HL7 segment splitting.
	//    The literal bytes are: 0x5C 0x72  →  '\' 'r'
	clean = bytes.ReplaceAll(clean, []byte{'\\', 'r'}, []byte{'\r'})

	// 3. Normalize CRLF → CR
	clean = bytes.ReplaceAll(clean, []byte("\r\n"), []byte("\r"))

	// 4. Normalize LF-only → CR
	clean = bytes.ReplaceAll(clean, []byte("\n"), []byte("\r"))

	// 5. Remove UTF‑8 BOM if present
	bom := []byte{0xEF, 0xBB, 0xBF}
	clean = bytes.TrimPrefix(clean, bom)

	return clean
}
