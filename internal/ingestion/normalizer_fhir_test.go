package ingestion

import (
	"testing"
)

func TestFHIRNormalizer(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		env       envelope
		expectErr bool
		verify    func(t *testing.T, ce *CanonicalEvent)
	}{
		{
			name: "Patient basic extraction",
			raw: `{
                "resourceType": "Patient",
                "id": "pat123",
                "name": [{
                    "given": ["John"],
                    "family": "Doe"
                }]
            }`,
			env: envelope{EventID: "evt1", SourceSystem: "test"},
			verify: func(t *testing.T, ce *CanonicalEvent) {
				if ce.Patient == nil {
					t.Fatalf("expected patient to be populated")
				}
				if ce.Patient.ID != "pat123" {
					t.Fatalf("expected patient id pat123, got %s", ce.Patient.ID)
				}
				if ce.Patient.FirstName != "John" {
					t.Fatalf("expected first name John, got %s", ce.Patient.FirstName)
				}
				if ce.Patient.LastName != "Doe" {
					t.Fatalf("expected last name Doe, got %s", ce.Patient.LastName)
				}
			},
		},
		{
			name: "Encounter basic extraction",
			raw: `{
                "resourceType": "Encounter",
                "id": "enc789",
                "class": { "code": "AMB" }
            }`,
			env: envelope{EventID: "evt2", SourceSystem: "test"},
			verify: func(t *testing.T, ce *CanonicalEvent) {
				if ce.Encounter == nil {
					t.Fatalf("expected encounter to be populated")
				}
				if ce.Encounter.ID != "enc789" {
					t.Fatalf("expected encounter id enc789, got %s", ce.Encounter.ID)
				}
				if ce.Encounter.Type != "AMB" {
					t.Fatalf("expected encounter type AMB, got %s", ce.Encounter.Type)
				}
			},
		},
		{
			name: "Observation basic extraction",
			raw: `{
                "resourceType": "Observation",
                "code": {
                    "coding": [{
                        "code": "12345-6"
                    }]
                },
                "valueString": "98.6"
            }`,
			env: envelope{EventID: "evt3", SourceSystem: "test"},
			verify: func(t *testing.T, ce *CanonicalEvent) {
				if ce.Observation == nil {
					t.Fatalf("expected observation to be populated")
				}
				if ce.Observation.Code != "12345-6" {
					t.Fatalf("expected code 12345-6, got %s", ce.Observation.Code)
				}
				if ce.Observation.Value != "98.6" {
					t.Fatalf("expected value 98.6, got %v", ce.Observation.Value)
				}
			},
		},
		{
			name: "Unsupported resourceType",
			raw: `{
                "resourceType": "Medication",
                "id": "med001"
            }`,
			env:       envelope{EventID: "evt4", SourceSystem: "test"},
			expectErr: true,
		},
	}

	n := NewFHIRNormalizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce, err := n.Normalize(tt.raw, tt.env)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, ce)
			}
		})
	}
}
