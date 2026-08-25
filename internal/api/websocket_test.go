package api

import (
	"encoding/json"
	"testing"
)

func TestLiveClient_parsePayload_ErrorResponse(t *testing.T) {
	client := NewLiveClient("token", "home")

	payload := []byte(`{
		"errors": [{"message": "user not authorized to access home", "extensions": {"code": "FORBIDDEN"}}],
		"data": {"liveMeasurement": null}
	}`)

	measurement, err := client.parsePayload(json.RawMessage(payload))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if measurement != nil {
		t.Errorf("expected measurement to be nil, got %v", measurement)
	}

	expectedErr := "user not authorized to access home"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestLiveClient_parsePayload_NilMeasurement(t *testing.T) {
	client := NewLiveClient("token", "home")

	payload := []byte(`{
		"data": {"liveMeasurement": null}
	}`)

	measurement, err := client.parsePayload(json.RawMessage(payload))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if measurement != nil {
		t.Errorf("expected measurement to be nil, got %v", measurement)
	}

	expectedErr := "no measurement data received"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestLiveClient_parsePayload_Success(t *testing.T) {
	client := NewLiveClient("token", "home")

	payload := []byte(`{
		"data": {
			"liveMeasurement": {
				"timestamp": "2023-10-01T12:00:00Z",
				"power": 1234
			}
		}
	}`)

	measurement, err := client.parsePayload(json.RawMessage(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if measurement == nil {
		t.Fatal("expected measurement, got nil")
	}
}
