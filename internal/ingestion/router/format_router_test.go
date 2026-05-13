package router

import (
	"context"
	"errors"
	"testing"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/ingestion/normalizer"
	"github.com/ajawes/hesp/internal/ingestion/transformer"
	"github.com/ajawes/hesp/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

// ------------------------------------------------------------
// Observability initialization for tests
// ------------------------------------------------------------
func init() {
	observability.NewLogger("hesp-ecs", "test")
	observability.InitMetrics("hesp-ecs", "test")
	otel.SetTracerProvider(noop.NewTracerProvider())
}

// ------------------------------------------------------------
// Fake Detector
// ------------------------------------------------------------
type fakeDetector struct {
	format config.Format
}

func (f fakeDetector) Detect(_ []byte) config.Format {
	return f.format
}

// ------------------------------------------------------------
// Fake Normalizer (ctx-aware)
// ------------------------------------------------------------
type fakeNormalizer struct {
	out *models.NormalizedEvent
	err error
}

func (f fakeNormalizer) Normalize(_ context.Context, _ []byte, _ api.Envelope) (*models.NormalizedEvent, error) {
	return f.out, f.err
}

type fakeNormalizationRouter struct {
	norm fakeNormalizer
	err  error
}

func (r fakeNormalizationRouter) NormalizerFor(_ config.Format) (normalizer.Normalizer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.norm, nil
}

// ------------------------------------------------------------
// Fake Transformer (ctx-aware)
// ------------------------------------------------------------
type fakeTransformer struct {
	out *models.CanonicalEvent
	err error
}

func (f fakeTransformer) Transform(_ context.Context, _ *models.NormalizedEvent, _ api.Envelope) (*models.CanonicalEvent, error) {
	return f.out, f.err
}

type fakeTransformationRouter struct {
	xfm fakeTransformer
	err error
}

func (r fakeTransformationRouter) TransformerFor(_ config.Format) (transformer.Transformer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.xfm, nil
}

// ------------------------------------------------------------
// Fake Dispatcher (ctx-aware)
// ------------------------------------------------------------
type fakeDispatcher struct {
	called bool
	err    error
}

func (d *fakeDispatcher) Dispatch(_ context.Context, _ *models.CanonicalEvent, _ api.Envelope, _ []byte) error {
	d.called = true
	return d.err
}

// ------------------------------------------------------------
// ⭐ NEW: Fake Compliance Guard
// ------------------------------------------------------------
type mockComplianceGuard struct {
	err error
}

func (m *mockComplianceGuard) Apply(_ context.Context, _ *models.CanonicalEvent) error {
	return m.err
}

// ------------------------------------------------------------
// Tests
// ------------------------------------------------------------
func TestFormatRouter(t *testing.T) {
	tests := []struct {
		name           string
		format         config.Format
		normErr        error
		normLookupErr  error
		xfmErr         error
		xfmLookupErr   error
		dispatchErr    error
		expectErr      bool
		expectDispatch bool
	}{
		// Success cases
		{"hl7_success", config.FormatHL7, nil, nil, nil, nil, nil, false, true},
		{"x12_success", config.FormatX12, nil, nil, nil, nil, nil, false, true},
		{"fhir_success", config.FormatFHIR, nil, nil, nil, nil, nil, false, true},
		{"generic_success", config.FormatGeneric, nil, nil, nil, nil, nil, false, true},

		// Failure cases
		{"normalizer_lookup_error", config.FormatHL7, nil, errors.New("no norm"), nil, nil, nil, true, false},
		{"normalizer_error", config.FormatHL7, errors.New("norm fail"), nil, nil, nil, nil, true, false},
		{"transformer_lookup_error", config.FormatHL7, nil, nil, nil, errors.New("no xfm"), nil, true, false},
		{"transformer_error", config.FormatHL7, nil, nil, errors.New("xfm fail"), nil, nil, true, false},
		{"dispatcher_error", config.FormatHL7, nil, nil, nil, nil, errors.New("dispatch fail"), true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fd := &fakeDispatcher{err: tt.dispatchErr}

			r := NewFormatRouter(
				WithDetector(fakeDetector{format: tt.format}),
				WithNormalizationRouter(fakeNormalizationRouter{
					norm: fakeNormalizer{
						out: &models.NormalizedEvent{
							Format: tt.format,
							Fields: map[string]any{},
						},
						err: tt.normErr,
					},
					err: tt.normLookupErr,
				}),
				WithTransformationRouter(fakeTransformationRouter{
					xfm: fakeTransformer{
						out: &models.CanonicalEvent{
							EventID: "abc123",
							Format:  tt.format,
						},
						err: tt.xfmErr,
					},
					err: tt.xfmLookupErr,
				}),
				WithComplianceGuard(&mockComplianceGuard{}), // ⭐ REQUIRED
				WithDispatcher(fd),
			)

			env := api.Envelope{
				EventID:      "abc123",
				EventType:    "test",
				SourceSystem: "unit",
			}

			_, err := r.Route(context.Background(), []byte("raw"), env)

			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if fd.called != tt.expectDispatch {
				t.Fatalf("dispatcher called=%v, expected=%v", fd.called, tt.expectDispatch)
			}
		})
	}
}
