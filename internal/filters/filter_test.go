package filters_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/prymitive/karma/internal/alertmanager"
	"github.com/prymitive/karma/internal/filters"
	"github.com/prymitive/karma/internal/models"

	log "github.com/sirupsen/logrus"
)

type filterTest struct {
	Expression string
	IsValid    bool
	IsMatch    bool
	Alert      models.Alert
	Silence    models.Silence
}

var tests = []filterTest{
	{
		Expression: "@state=active",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    false,
	},
	{
		Expression: "@state!=active",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    true,
	},
	{
		Expression: "@state=suppressed",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		IsMatch:    true,
	},
	{
		Expression: "@state!=suppressed",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		IsMatch:    false,
	},
	{
		Expression: "@state=xx",
		IsValid:    false,
	},
	{
		Expression: "@state=:xx",
		IsValid:    false,
	},
	{
		Expression: "@state==xx",
		IsValid:    false,
	},
	{
		Expression: "@state=~true",
		IsValid:    false,
	},
	{
		Expression: "@state=~false",
		IsValid:    false,
	},
	{
		Expression: "@state=suppressed",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", InhibitedBy: []string{"999"}},
		IsMatch:    true,
	},
	{
		Expression: "@state=suppressed",
		IsValid:    true,
		Alert:      models.Alert{State: "active"},
		IsMatch:    false,
	},
	{
		Expression: "@state!=suppressed",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", InhibitedBy: []string{"999"}},
		IsMatch:    false,
	},

	{
		Expression: "@silence_id=abcdef",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		IsMatch:    false,
	},
	{
		Expression: "@silence_id=abcdef",
		IsValid:    true,
		Alert:      models.Alert{State: "active"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_id=abcdef",
		IsValid:    true,
		Alert:      models.Alert{State: "active", SilencedBy: []string{"abcdef"}},
		IsMatch:    false,
	},
	{
		Expression: "@silence_id=abcdef",
		IsValid:    true,
		Alert:      models.Alert{State: "unprocessed"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_id=abcdef",
		IsValid:    true,
		Alert:      models.Alert{State: "unprocessed", SilencedBy: []string{"abcdef"}},
		IsMatch:    false,
	},
	{
		Expression: "@silence_id=abcdef",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"abcdef"}},
		IsMatch:    true,
	},
	{
		Expression: "@silence_id!=abcdef",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"abcdef"}},
		IsMatch:    false,
	},
	{
		Expression: "@silence_id!=abcdef",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		IsMatch:    true,
	},
	{
		Expression: "@silence_id=~cde",
		IsValid:    false,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"abcdef"}},
		IsMatch:    false,
	},
	{
		Expression: "@silence_id!~abc",
		IsValid:    false,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"zwd"}},
		IsMatch:    false,
	},

	{
		Expression: "@silence_jira=1",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", JiraID: "1"},
		IsMatch:    true,
	},
	{
		Expression: "@silence_jira=2",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_jira!=3",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", JiraID: "x"},
		IsMatch:    true,
	},
	{
		Expression: "@silence_jira!=4",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", JiraID: "4"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_jira!=5",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1"},
		IsMatch:    true,
	},
	{
		Expression: "@silence_jira=~abc",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", JiraID: "xxabcxx"},
		IsMatch:    true,
	},
	{
		Expression: "@silence_jira=~abc",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", JiraID: "xxx"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_jira=~",
		IsValid:    false,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", JiraID: "xxx"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_jira~=",
		IsValid:    false,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", JiraID: "xxx"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_jira~=1",
		IsValid:    false,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", JiraID: "xxx"},
		IsMatch:    false,
	},

	{
		Expression: "@silence_author=john",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", CreatedBy: "john"},
		IsMatch:    true,
	},
	{
		Expression: "@silence_author=john",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", CreatedBy: "bob"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_author!=john",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", CreatedBy: "bob"},
		IsMatch:    true,
	},
	{
		Expression: "@silence_author!=john",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", CreatedBy: "john"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_author!=john",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1"},
		IsMatch:    true,
	},
	{
		Expression: "@silence_author=~",
		IsValid:    false,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_author===x",
		IsValid:    false,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1"},
		IsMatch:    false,
	},
	{
		Expression: "@silence_author=!!xxx",
		IsValid:    false,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1"},
		IsMatch:    false,
	},

	{
		Expression: "@age<1h",
		IsValid:    true,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Minute * -55)},
		IsMatch:    true,
	},
	{
		Expression: "@age>1h",
		IsValid:    true,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Hour * -2)},
		IsMatch:    true,
	},
	{
		Expression: "@age<-1h",
		IsValid:    true,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Minute * -55)},
		IsMatch:    true,
	},
	{
		Expression: "@age>-1h",
		IsValid:    true,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Hour * -2)},
		IsMatch:    true,
	},
	{
		Expression: "@age=1h",
		IsValid:    false,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Minute * -55)},
		IsMatch:    false,
	},
	{
		Expression: "@age=~1h",
		IsValid:    false,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Minute * -55)},
		IsMatch:    false,
	},
	{
		Expression: "@age>",
		IsValid:    false,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Minute * -55)},
		IsMatch:    false,
	},
	{
		Expression: "@age<",
		IsValid:    false,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Minute * -55)},
		IsMatch:    false,
	},
	{
		Expression: "@age>a",
		IsValid:    false,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Minute * -55)},
		IsMatch:    false,
	},
	{
		Expression: "@age<10v",
		IsValid:    false,
		Alert:      models.Alert{StartsAt: time.Now().Add(time.Minute * -55)},
		IsMatch:    false,
	},

	{
		Expression: "node=vps1",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps1"}},
		IsMatch:    true,
	},
	{
		Expression: "node=vps1",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    false,
	},
	{
		Expression: "node!=vps1",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps1"}},
		IsMatch:    false,
	},
	{
		Expression: "node!=vps1",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps2"}},
		IsMatch:    true,
	},
	{
		Expression: "node=~vps",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps1"}},
		IsMatch:    true,
	},
	{
		Expression: "node!~vps",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps1"}},
		IsMatch:    false,
	},
	{
		Expression: "node!~abc",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps1"}},
		IsMatch:    true,
	},
	{
		Expression: "node!~",
		IsValid:    false,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps1"}},
		IsMatch:    false,
	},
	{
		Expression: "node=",
		IsValid:    false,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps1"}},
		IsMatch:    false,
	},
	{
		Expression: "node===",
		IsValid:    false,
		Alert:      models.Alert{Labels: map[string]string{"node": "vps1"}},
		IsMatch:    false,
	},

	{
		Expression: "abc",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"key": "abc"}},
		IsMatch:    true,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"key": "XXXabcx"}},
		IsMatch:    true,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert:      models.Alert{Labels: map[string]string{"abc": "xxxab"}},
		IsMatch:    false,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert: models.Alert{
			Annotations: models.Annotations{
				models.Annotation{Name: "key", Value: "abc"},
			},
		},
		IsMatch: true,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert: models.Alert{
			Annotations: models.Annotations{
				models.Annotation{Name: "key", Value: "ccc abc"},
			},
		},
		IsMatch: true,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert: models.Alert{
			Annotations: models.Annotations{
				models.Annotation{Name: "abc", Value: "zzz"},
			},
		},
		IsMatch: false,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", Comment: "abc"},
		IsMatch:    true,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", Comment: "abcxxx"},
		IsMatch:    true,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", Comment: "ABCD"},
		IsMatch:    true,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert:      models.Alert{State: "suppressed", SilencedBy: []string{"1"}},
		Silence:    models.Silence{ID: "1", Comment: "xzc"},
		IsMatch:    false,
	},
	{
		Expression: "abc",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    false,
	},
	{
		Expression: "^abb[****].*****",
		IsValid:    false,
	},
	{
		Expression: "@silenced=true",
		IsValid:    false,
	},
	{
		Expression: "@silenced!=false",
		IsValid:    false,
	},
	{
		Expression: "@silenced=~false",
		IsValid:    false,
	},
	{
		Expression: "@inhibited=true",
		IsValid:    false,
	},
	{
		Expression: "@inhibited!=false",
		IsValid:    false,
	},
	{
		Expression: "@inhibited=~false",
		IsValid:    false,
	},
	{
		Expression: "@alertmanager=test",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    true,
	},
	{
		Expression: "@alertmanager=abc",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    false,
	},
	{
		Expression: "@alertmanager=~tes",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    true,
	},
	{
		Expression: "@alertmanager=~000",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    false,
	},
	{
		Expression: "@alertmanager!=tes",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    true,
	},
	{
		Expression: "@alertmanager!~abc",
		IsValid:    true,
		Alert:      models.Alert{},
		IsMatch:    true,
	},
	{
		Expression: "@receiver=by-name",
		IsValid:    true,
		Alert: models.Alert{
			Receiver: "by-name",
		},
		IsMatch: true,
	},
	{
		Expression: "@receiver=by-name",
		IsValid:    true,
		Alert: models.Alert{
			Receiver: "by-not-name",
		},
		IsMatch: false,
	},
	{
		Expression: "@receiver=~name",
		IsValid:    true,
		Alert: models.Alert{
			Receiver: "by-not-name",
		},
		IsMatch: true,
	},
}

func TestFilters(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	am, err := alertmanager.NewAlertmanager("test", "http://localhost", alertmanager.WithRequestTimeout(time.Second))
	if err != nil {
		t.Error(err)
	}
	for _, ft := range tests {
		alert := models.Alert(ft.Alert)
		alert.Alertmanager = []models.AlertmanagerInstance{
			{
				Name:       am.Name,
				Silences:   map[string]*models.Silence{},
				SilencedBy: []string{},
			},
		}
		if ft.Silence.ID != "" {
			alert.Alertmanager[0].Silences[ft.Silence.ID] = &ft.Silence
			alert.Alertmanager[0].SilencedBy = append(alert.Alertmanager[0].SilencedBy, ft.Silence.ID)
		}

		f := filters.NewFilter(ft.Expression)
		if f == nil {
			t.Errorf("[%s] No filter found", ft.Expression)
		}
		if f.GetHits() != 0 {
			t.Errorf("[%s] Hits = %#v after init(), expected 0", ft.Expression, f.GetHits())
		}
		if f.GetIsValid() != ft.IsValid {
			t.Errorf("[%s] GetIsValid() returned %#v while %#v was expected", ft.Expression, f.GetIsValid(), ft.IsValid)
		}
		if f.GetIsValid() {
			m := f.Match(&alert, 0)
			if m != ft.IsMatch {
				j, _ := json.Marshal(ft.Alert)
				s, _ := json.Marshal(ft.Silence)
				t.Errorf("[%s] Match() returned %#v while %#v was expected\nalert used: %s\nsilence used: %s", ft.Expression, m, ft.IsMatch, j, s)
			}
			if ft.IsMatch && f.GetHits() != 1 {
				t.Errorf("[%s] GetHits() returned %#v after match, expected 1", ft.Expression, f.GetHits())
			}
			if !ft.IsMatch && f.GetHits() != 0 {
				t.Errorf("[%s] GetHits() returned %#v after non-match, expected 0", ft.Expression, f.GetHits())
			}
			if f.GetRawText() != ft.Expression {
				t.Errorf("[%s] GetRawText() returned %#v != %s passed as the expression", ft.Expression, f.GetRawText(), ft.Expression)
			}
		}
	}
}

type limitFilterTest struct {
	Expression string
	IsValid    bool
	IsMatch    []bool
	Hits       int
}

var limitTests = []limitFilterTest{
	{
		Expression: "@limit=3",
		IsValid:    true,
		IsMatch:    []bool{true, true, true},
		Hits:       0,
	},
	{
		Expression: "@limit=1",
		IsValid:    true,
		IsMatch:    []bool{true, false, false},
		Hits:       2,
	},
	{
		Expression: "@limit=5",
		IsValid:    true,
		IsMatch:    []bool{true, true, true, true, true, false, false, false},
		Hits:       3,
	},
	{
		Expression: "@limit=0",
		IsValid:    false,
	},
	{
		Expression: "@limit=abc",
		IsValid:    false,
	},
	{
		Expression: "@limit==0",
		IsValid:    false,
	},
	{
		Expression: "@limit>0",
		IsValid:    false,
	},
	{
		Expression: "@limit<0",
		IsValid:    false,
	},
}

func TestLimitFilter(t *testing.T) {
	for _, ft := range limitTests {
		f := filters.NewFilter(ft.Expression)
		if f == nil {
			t.Errorf("[%s] No filter found", ft.Expression)
		}
		if f.GetHits() != 0 {
			t.Errorf("[%s] Hits = %#v after init(), expected 0", ft.Expression, f.GetHits())
		}
		if f.GetIsValid() != ft.IsValid {
			t.Errorf("[%s] GetIsValid() returned %#v while %#v was expected", ft.Expression, f.GetIsValid(), ft.IsValid)
		}
		if f.GetIsValid() {
			alert := models.Alert{}
			var index int
			for _, isMatch := range ft.IsMatch {
				m := f.Match(&alert, index)
				if m != isMatch {
					t.Errorf("[%s] Match() returned %#v while %#v was expected, index %d", ft.Expression, m, isMatch, index)
				}
				if f.GetRawText() != ft.Expression {
					t.Errorf("[%s] GetRawText() returned %#v != %s passed as the expression", ft.Expression, f.GetRawText(), ft.Expression)
				}
				index++
			}
			if f.GetHits() != ft.Hits {
				t.Errorf("[%s] GetHits() returned %#v hits, expected %d", ft.Expression, f.GetHits(), ft.Hits)
			}
		}
	}
}
