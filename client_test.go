package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("tcp://127.0.0.1:9000", nil)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NoError(t, client.DB.Ping())
}

func TestClient_SaveHit(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{
			ClientID:                  1,
			Fingerprint:               "fp",
			Time:                      time.Now(),
			Session:                   time.Now(),
			PreviousTimeOnPageSeconds: 42,
			UserAgent:                 "ua",
			Path:                      "/path",
			Title:                     "title",
			Language:                  "en",
			Referrer:                  "ref",
			ReferrerName:              "ref_name",
			ReferrerIcon:              "ref_icon",
			OS:                        "os",
			OSVersion:                 "10",
			Browser:                   "browser",
			BrowserVersion:            "89",
			CountryCode:               "en",
			Desktop:                   true,
			Mobile:                    false,
			ScreenWidth:               1920,
			ScreenHeight:              1080,
			ScreenClass:               "XL",
		},
		{
			Fingerprint: "fp",
			Time:        time.Now().UTC(),
			UserAgent:   "ua",
			Path:        "/path",
		},
	}))
}

func TestClient_SaveEvent(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveEvents([]Event{
		{
			Hit: Hit{
				ClientID:                  1,
				Fingerprint:               "fp",
				Time:                      time.Now(),
				Session:                   time.Now(),
				PreviousTimeOnPageSeconds: 42,
				UserAgent:                 "ua",
				Path:                      "/path",
				Title:                     "title",
				Language:                  "en",
				Referrer:                  "ref",
				ReferrerName:              "ref_name",
				ReferrerIcon:              "ref_icon",
				OS:                        "os",
				OSVersion:                 "10",
				Browser:                   "browser",
				BrowserVersion:            "89",
				CountryCode:               "en",
				Desktop:                   true,
				Mobile:                    false,
				ScreenWidth:               1920,
				ScreenHeight:              1080,
				ScreenClass:               "XL",
			},
			Name:            "event_name",
			DurationSeconds: 21,
			MetaKeys:        []string{"meta", "keys"},
			MetaValues:      []string{"some", "values"},
		},
		{
			Hit: Hit{
				Fingerprint: "fp",
				Time:        time.Now().UTC(),
				UserAgent:   "ua",
				Path:        "/path",
			},
			Name: "different_event",
		},
	}))
}

func TestClient_Session(t *testing.T) {
	cleanupDB()
	fp := "session_fp"
	now := time.Now().UTC().Add(-time.Second * 20)
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{
			ClientID:    1,
			Fingerprint: fp,
			Time:        now.Add(-time.Second * 20),
			Session:     now.Add(-time.Second * 20),
			UserAgent:   "ua",
			Path:        "/path1",
		},
		{
			ClientID:    1,
			Fingerprint: fp,
			Time:        now,
			Session:     now,
			UserAgent:   "ua",
			Path:        "/path2",
		},
		{
			ClientID:    1,
			Fingerprint: fp,
			Time:        now.Add(-time.Second * 10),
			Session:     now.Add(-time.Second * 10),
			UserAgent:   "ua",
			Path:        "/path3",
		},
	}))
	session, err := dbClient.Session(1, fp, time.Now().UTC().Add(-time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, "/path2", session.Path)
	assert.Equal(t, now.Unix(), session.Time.Unix())
	assert.Equal(t, now.Unix(), session.Session.Unix())
}
