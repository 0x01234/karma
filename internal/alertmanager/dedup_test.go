package alertmanager_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"

	"github.com/prymitive/karma/internal/alertmanager"
	"github.com/prymitive/karma/internal/config"
	"github.com/prymitive/karma/internal/mock"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.ErrorLevel)
	httpmock.Activate()
	for _, version := range mock.ListAllMocks() {
		name := fmt.Sprintf("dedup-mock-%s", version)
		uri := fmt.Sprintf("http://%s.localhost", version)
		am, err := alertmanager.NewAlertmanager(name, uri, alertmanager.WithRequestTimeout(time.Second))
		if err != nil {
			log.Fatal(err)
		}
		err = alertmanager.RegisterAlertmanager(am)
		if err != nil {
			panic(fmt.Sprintf("Failed to register Alertmanager instance %s: %s", am.Name, err))
		}

		mock.RegisterURL(fmt.Sprintf("%s/metrics", uri), version, "metrics")
		mock.RegisterURL(fmt.Sprintf("%s/api/v1/status", uri), version, "api/v1/status")
		mock.RegisterURL(fmt.Sprintf("%s/api/v2/status", uri), version, "api/v2/status")
		mock.RegisterURL(fmt.Sprintf("%s/api/v1/silences", uri), version, "api/v1/silences")
		mock.RegisterURL(fmt.Sprintf("%s/api/v2/silences", uri), version, "api/v2/silences")
		mock.RegisterURL(fmt.Sprintf("%s/api/v1/alerts/groups", uri), version, "api/v1/alerts/groups")
		mock.RegisterURL(fmt.Sprintf("%s/api/v2/alerts/groups", uri), version, "api/v2/alerts/groups")
	}
}

func pullAlerts() error {
	for _, am := range alertmanager.GetAlertmanagers() {
		err := am.Pull()
		if err != nil {
			return err
		}
	}
	return nil
}

func TestDedupAlerts(t *testing.T) {
	if err := pullAlerts(); err != nil {
		t.Error(err)
	}
	alertGroups := alertmanager.DedupAlerts()

	if len(alertGroups) != 10 {
		t.Errorf("Expected %d alert groups, got %d", 10, len(alertGroups))
	}

	totalAlerts := 0
	for _, ag := range alertGroups {
		totalAlerts += len(ag.Alerts)
	}
	if totalAlerts != 24 {
		t.Errorf("Expected %d total alerts, got %d", 24, totalAlerts)
	}
}

func TestDedupAlertsWithoutLabels(t *testing.T) {
	config.Config.Labels.Keep = []string{"xyz"}
	if err := pullAlerts(); err != nil {
		t.Error(err)
	}
	alertGroups := alertmanager.DedupAlerts()
	config.Config.Labels.Keep = []string{}

	if len(alertGroups) != 10 {
		t.Errorf("Expected %d alert groups, got %d", 10, len(alertGroups))
	}

	totalAlerts := 0
	for _, ag := range alertGroups {
		totalAlerts += len(ag.Alerts)
	}
	if totalAlerts != 24 {
		t.Errorf("Expected %d total alerts, got %d", 24, totalAlerts)
	}
}

func TestDedupAutocomplete(t *testing.T) {
	if err := pullAlerts(); err != nil {
		t.Error(err)
	}
	ac := alertmanager.DedupAutocomplete()
	mockCount := len(mock.ListAllMockURIs())
	// 56 hints for everything except @alertmanager and @silence_id
	// 4 hints for @silence_id 1 and 2
	// 2 hints per @alertmanager
	// 6 hints for silences in for each alertmanager
	// silence id might get duplicated so this check isn't very strict
	expected := 56 + 4 + mockCount*2 + mockCount*6
	if len(ac) <= int(float64(expected)*0.8) || len(ac) > expected {
		t.Errorf("Expected %d autocomplete hints, got %d", expected, len(ac))
	}
}

func TestDedupColors(t *testing.T) {
	os.Setenv("LABELS_COLOR_UNIQUE", "cluster instance @receiver")
	os.Setenv("ALERTMANAGER_URI", "http://localhost")
	config.Config.Read()
	if err := pullAlerts(); err != nil {
		t.Error(err)
	}
	colors := alertmanager.DedupColors()
	expected := 3
	if len(colors) != expected {
		t.Errorf("Expected %d color keys, got %d", expected, len(colors))
	}
}

func TestStripReceivers(t *testing.T) {
	os.Setenv("RECEIVERS_STRIP", "by-name by-cluster-service")
	os.Setenv("ALERTMANAGER_URI", "http://localhost")
	config.Config.Read()
	if err := pullAlerts(); err != nil {
		t.Error(err)
	}
	alerts := alertmanager.DedupAlerts()
	if len(alerts) > 0 {
		t.Errorf("Expected no alerts after stripping all receivers, got %d", len(alerts))
	}
}

func TestClearData(t *testing.T) {
	log.SetLevel(log.PanicLevel)
	httpmock.Activate()
	for _, version := range mock.ListAllMocks() {
		name := fmt.Sprintf("clear-data-mock-%s", version)
		uri := fmt.Sprintf("http://localhost/clear/%s", version)
		am, _ := alertmanager.NewAlertmanager(name, uri, alertmanager.WithRequestTimeout(time.Second))

		mock.RegisterURL(fmt.Sprintf("%s/metrics", uri), version, "metrics")
		_ = am.Pull()
		if am.Version() != "" {
			t.Errorf("[%s] Got non-empty version string: %s", am.Name, am.Version())
		}
		if am.Error() == "" {
			t.Errorf("[%s] Got empty error string", am.Name)
		}
		if len(am.Silences()) != 0 {
			t.Errorf("[%s] Get %d silences", am.Name, len(am.Silences()))
		}
		if len(am.Alerts()) != 0 {
			t.Errorf("[%s] Get %d alerts", am.Name, len(am.Alerts()))
		}
		if len(am.KnownLabels()) != 0 {
			t.Errorf("[%s] Get %d known labels", am.Name, len(am.KnownLabels()))
		}

		mock.RegisterURL(fmt.Sprintf("%s/api/v1/status", uri), version, "api/v1/status")
		mock.RegisterURL(fmt.Sprintf("%s/api/v2/status", uri), version, "api/v2/status")
		mock.RegisterURL(fmt.Sprintf("%s/api/v1/silences", uri), version, "api/v1/silences")
		mock.RegisterURL(fmt.Sprintf("%s/api/v2/silences", uri), version, "api/v2/silences")
		_ = am.Pull()
		if am.Version() != "" {
			t.Errorf("[%s] Got non-empty version string: %s", am.Name, am.Version())
		}
		if am.Error() == "" {
			t.Errorf("[%s] Got empty error string", am.Name)
		}
		if len(am.Silences()) != 0 {
			t.Errorf("[%s] Get %d silences", am.Name, len(am.Silences()))
		}
		if len(am.Alerts()) != 0 {
			t.Errorf("[%s] Get %d alerts", am.Name, len(am.Alerts()))
		}
		if len(am.KnownLabels()) != 0 {
			t.Errorf("[%s] Get %d known labels", am.Name, len(am.KnownLabels()))
		}

		mock.RegisterURL(fmt.Sprintf("%s/api/v1/alerts/groups", uri), version, "api/v1/alerts/groups")
		mock.RegisterURL(fmt.Sprintf("%s/api/v2/alerts/groups", uri), version, "api/v2/alerts/groups")
		_ = am.Pull()
		if am.Version() == "" {
			t.Errorf("[%s] Got empty version string", am.Name)
		}
		if am.Error() != "" {
			t.Errorf("[%s] Got non-empty error string: %s", am.Name, am.Error())
		}
		if len(am.Silences()) == 0 {
			t.Errorf("[%s] Get %d silences", am.Name, len(am.Silences()))
		}
		if len(am.Alerts()) == 0 {
			t.Errorf("[%s] Get %d alerts", am.Name, len(am.Alerts()))
		}
		if len(am.KnownLabels()) == 0 {
			t.Errorf("[%s] Get %d known labels", am.Name, len(am.KnownLabels()))
		}
	}
}
