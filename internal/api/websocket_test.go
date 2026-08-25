package api

import (
	"encoding/json"
	"testing"
)

func TestLiveClient_parsePayload(t *testing.T) {
	client := NewLiveClient("token", "home")

	tests := []struct {
		name        string
		payload     string
		expectErr   bool
		errContains string
	}{
		{
			name: "error response for invalid home ID",
			payload: `{
				"errors": [{"message": "user not authorized to access home"}],
				"data": {"liveMeasurement": null}
			}`,
			expectErr:   true,
			errContains: "invalid or non-existing home ID",
		},
		{
			name: "other graphql error",
			payload: `{
				"errors": [{"message": "some other error"}],
				"data": {"liveMeasurement": null}
			}`,
			expectErr:   true,
			errContains: "some other error",
		},
		{
			name: "nil measurement data",
			payload: `{
				"data": {"liveMeasurement": null}
			}`,
			expectErr:   true,
			errContains: "no measurement data received (invalid home ID?)",
		},
		{
			name: "success payload",
			payload: `{
				"data": {
					"liveMeasurement": {
						"timestamp": "2023-10-01T12:00:00Z",
						"power": 1234
					}
				}
			}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			measurement, err := client.parsePayload(json.RawMessage(tt.payload))

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if measurement != nil {
					t.Errorf("expected measurement to be nil, got %v", measurement)
				}
				if err.Error() != tt.errContains {
					t.Errorf("expected error %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if measurement == nil {
					t.Fatal("expected measurement, got nil")
				}
			}
		})
	}
}
