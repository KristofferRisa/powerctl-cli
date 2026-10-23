package commands

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kristofferrisa/powerctl-cli/internal/models"
)

func TestConsumptionRange(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}

	tests := []struct {
		name     string
		last     int
		lastSet  bool
		from, to string
		wantLast int
		wantFrom time.Time
		wantTo   time.Time
		wantErr  string
		wantTZ   bool // whether the home's time zone must be looked up
	}{
		{name: "default last", last: 7, wantLast: 7},
		{name: "last above one page", last: 2000, lastSet: true, wantLast: 2000},
		{name: "last zero", last: 0, lastSet: true, wantErr: "must be greater than 0"},
		{
			name: "from date in home time zone", last: 7, from: "2026-01-01",
			wantFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, oslo), wantTZ: true,
		},
		{
			name: "from and to dates", last: 7, from: "2026-01-01", to: "2026-07-01",
			wantFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, oslo),
			wantTo:   time.Date(2026, 7, 1, 0, 0, 0, 0, oslo), wantTZ: true,
		},
		{
			name: "RFC 3339 needs no time zone lookup", last: 7,
			from: "2026-01-01T00:00:00Z", to: "2026-01-02T12:00:00+02:00",
			wantFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantTo:   time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
		},
		{name: "to without from", last: 7, to: "2026-07-01", wantErr: "--to requires --from"},
		{name: "from with last", last: 30, lastSet: true, from: "2026-01-01", wantErr: "cannot be used together"},
		{name: "to before from", last: 7, from: "2026-07-01", to: "2026-01-01", wantErr: "must be after --from"},
		{name: "to equal to from", last: 7, from: "2026-07-01", to: "2026-07-01", wantErr: "must be after --from"},
		{name: "bad from", last: 7, from: "01/01/2026", wantErr: `invalid --from "01/01/2026"`},
		{name: "bad to", last: 7, from: "2026-01-01", to: "July", wantErr: `invalid --to "July"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookups := 0
			homeLocation := func() (*time.Location, error) {
				lookups++
				return oslo, nil
			}

			r, err := consumptionRange(tt.last, tt.lastSet, tt.from, tt.to, homeLocation)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if r.Last != tt.wantLast {
				t.Errorf("Last = %d, want %d", r.Last, tt.wantLast)
			}
			if !r.From.Equal(tt.wantFrom) {
				t.Errorf("From = %s, want %s", r.From, tt.wantFrom)
			}
			if !r.To.Equal(tt.wantTo) {
				t.Errorf("To = %s, want %s", r.To, tt.wantTo)
			}
			if wantLookups := map[bool]int{true: 1, false: 0}[tt.wantTZ]; lookups != wantLookups {
				t.Errorf("time zone lookups = %d, want %d", lookups, wantLookups)
			}
		})
	}
}

func TestConsumptionRange_TimeZoneError(t *testing.T) {
	_, err := consumptionRange(7, false, "2026-01-01", "", func() (*time.Location, error) {
		return nil, errors.New("boom")
	})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error = %v, want one wrapping boom", err)
	}
}

func TestShortHistoryNotice(t *testing.T) {
	if got := shortHistoryNotice(nil); !strings.Contains(got, "no Consumption History") {
		t.Errorf("notice = %q", got)
	}

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	got := shortHistoryNotice([]models.ConsumptionNode{{From: from}, {From: from.Add(time.Hour)}})
	if !strings.Contains(got, "2026-03-01T00:00:00Z") || !strings.Contains(got, "2 Nodes") {
		t.Errorf("notice = %q", got)
	}
}
