package pipeline_test

import (
	"strings"
	"testing"

	"github.com/ajawes/hesp/internal/ingestion/pipeline"
	"github.com/ajawes/hesp/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ------------------------------------------------------------
// Observability initialization for tests
// ------------------------------------------------------------
func init() {
	observability.NewLogger("hesp-ecs", "test")
	observability.InitMetrics("hesp-ecs", "test")
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
}

func TestPreSanitize(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		// ------------------------------------------------------------
		// HL7 TEST
		// ------------------------------------------------------------
		{
			name:     "HL7 escaped CR sequences become real CR",
			raw:      "MSH|^~\\\\&|ADT|HOSPITAL|HESP|INGEST|202605041200||ADT^A01|123456|P|2.5\\rPID|1||12345^^^HOSPITAL^MR||DOE^JOHN||19800101|M|||123 MAIN ST^^RALEIGH^NC^27612||5551234567\\rPV1|1|I",
			expected: "MSH|^~\\\\&|ADT|HOSPITAL|HESP|INGEST|202605041200||ADT^A01|123456|P|2.5\rPID|1||12345^^^HOSPITAL^MR||DOE^JOHN||19800101|M|||123 MAIN ST^^RALEIGH^NC^27612||5551234567\rPV1|1|I",
		},

		// ------------------------------------------------------------
		// X12 TEST
		// ------------------------------------------------------------
		{
			name: "X12 escaped LF sequences become CR-delimited segments",
			raw: `ISA*00*          *00*          *ZZ*SENDERID       *ZZ*RECEIVERID     *260504*1200*^*00501*000000905*1*T*:~\n` +
				`GS*HC*SENDERID*RECEIVERID*20260504*1200*1*X*005010X222A1~\n` +
				`ST*837*0001*005010X222A1~\n` +
				`BHT*0019*00*0123*20260504*1200*CH~\n` +
				`NM1*41*2*SENDER BILLING SERVICE*****46*123456789~\n` +
				`SE*6*0001~\n` +
				`GE*1*1~\n` +
				`IEA*1*000000905~`,
			expected: "ISA*00*          *00*          *ZZ*SENDERID       *ZZ*RECEIVERID     *260504*1200*^*00501*000000905*1*T*:~\r" +
				"GS*HC*SENDERID*RECEIVERID*20260504*1200*1*X*005010X222A1~\r" +
				"ST*837*0001*005010X222A1~\r" +
				"BHT*0019*00*0123*20260504*1200*CH~\r" +
				"NM1*41*2*SENDER BILLING SERVICE*****46*123456789~\r" +
				"SE*6*0001~\r" +
				"GE*1*1~\r" +
				"IEA*1*000000905~",
		},

		// ------------------------------------------------------------
		// FHIR PATIENT TEST
		// ------------------------------------------------------------
		{
			name: "FHIR Patient JSON normalized LF→CR",
			raw: `{
                "resourceType": "Patient",
                "id": "example",
                "name": [{
                    "use": "official",
                    "family": "Doe",
                    "given": ["Jane"]
                }],
                "gender": "female",
                "birthDate": "1990-01-01"
            }`,
			expected: "{\r                \"resourceType\": \"Patient\",\r                \"id\": \"example\",\r                \"name\": [{\r                    \"use\": \"official\",\r                    \"family\": \"Doe\",\r                    \"given\": [\"Jane\"]\r                }],\r                \"gender\": \"female\",\r                \"birthDate\": \"1990-01-01\"\r            }",
		},

		// ------------------------------------------------------------
		// FHIR OBSERVATION TEST
		// ------------------------------------------------------------
		{
			name: "FHIR Observation JSON normalized LF→CR",
			raw: `{
                "resourceType": "Observation",
                "id": "obs-001",
                "status": "final",
                "code": {
                    "coding": [{
                        "system": "http://loinc.org",
                        "code": "718-7",
                        "display": "Hemoglobin [Mass/volume] in Blood"
                    }]
                },
                "subject": { "reference": "Patient/example" },
                "effectiveDateTime": "2026-05-04T12:00:00Z",
                "valueQuantity": {
                    "value": 13.5,
                    "unit": "g/dL",
                    "system": "http://unitsofmeasure.org",
                    "code": "g/dL"
                }
            }`,
			expected: "{\r                \"resourceType\": \"Observation\",\r                \"id\": \"obs-001\",\r                \"status\": \"final\",\r                \"code\": {\r                    \"coding\": [{\r                        \"system\": \"http://loinc.org\",\r                        \"code\": \"718-7\",\r                        \"display\": \"Hemoglobin [Mass/volume] in Blood\"\r                    }]\r                },\r                \"subject\": { \"reference\": \"Patient/example\" },\r                \"effectiveDateTime\": \"2026-05-04T12:00:00Z\",\r                \"valueQuantity\": {\r                    \"value\": 13.5,\r                    \"unit\": \"g/dL\",\r                    \"system\": \"http://unitsofmeasure.org\",\r                    \"code\": \"g/dL\"\r                }\r            }",
		},

		// ------------------------------------------------------------
		// GENERIC TEST — AETNA SCENARIO
		// ------------------------------------------------------------
		{
			name: "Aetna eligibility decision event (generic JSON normalized LF→CR)",
			raw: `{
                "eventType": "eligibility_decision",
                "memberId": "AET123456789",
                "planId": "AET-PPO-2026",
                "decision": "approved",
                "effectiveDate": "2026-06-01",
                "requestedService": {
                    "serviceCode": "PT-THERAPY",
                    "description": "Physical Therapy Evaluation"
                },
                "utilization": {
                    "remainingVisits": 8,
                    "limit": 12
                },
                "notes": "Auto-approved based on clinical ruleset v4.2"
            }`,
			expected: "{\r                \"eventType\": \"eligibility_decision\",\r                \"memberId\": \"AET123456789\",\r                \"planId\": \"AET-PPO-2026\",\r                \"decision\": \"approved\",\r                \"effectiveDate\": \"2026-06-01\",\r                \"requestedService\": {\r                    \"serviceCode\": \"PT-THERAPY\",\r                    \"description\": \"Physical Therapy Evaluation\"\r                },\r                \"utilization\": {\r                    \"remainingVisits\": 8,\r                    \"limit\": 12\r                },\r                \"notes\": \"Auto-approved based on clinical ruleset v4.2\"\r            }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := string(pipeline.PreSanitize([]byte(tt.raw)))

			if strings.TrimSpace(out) != strings.TrimSpace(tt.expected) {
				t.Fatalf("\n--- Sanitization Mismatch ---\nExpected:\n%q\nGot:\n%q\n", tt.expected, out)
			}
		})
	}
}
