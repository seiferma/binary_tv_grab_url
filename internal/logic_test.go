package internal

import (
	_ "embed"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sherif-fanous/xmltv"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/533.xml
var xmltvSingleChannelWeek string

//go:embed testdata/533.xml.gz
var xmltvSingleChannelWeekGzip string

//go:embed testdata/533-20251005.xml
var xmltvSingleChannelFirstDay string

//go:embed testdata/533-20251006.xml
var xmltvSingleChannelSecondDay string

//go:embed testdata/533-exceptFirstTwo.xml
var xmltvSingleChannelSkipFirstTwoDays string

// 2025-10-05 12:00:00 UTC
var TEST_NOW = time.Date(2025, time.October, 5, 12, 00, 00, 0, time.UTC)

func TestGetContentFromUrlWithoutRestrictions(t *testing.T) {
	testWithExample(xmltvSingleChannelWeek, func(url string) {
		content, err := getContentForUrl(url, 0, 0, TEST_NOW)
		assert.Nil(t, err)
		assertXmlEquals(t, xmltvSingleChannelWeek, content)
	})
}

func TestGetContentFromUrlWithGzipWithoutRestrictions(t *testing.T) {
	testWithExample(xmltvSingleChannelWeekGzip, func(url string) {
		content, err := getContentForUrl(url, 0, 0, TEST_NOW)
		assert.Nil(t, err)
		assertXmlEquals(t, xmltvSingleChannelWeek, content)
	})
}

func TestGetContentFromUrlWithDayLimit(t *testing.T) {
	testWithExample(xmltvSingleChannelWeek, func(url string) {
		content, err := getContentForUrl(url, 1, 0, TEST_NOW)
		assert.Nil(t, err)
		assertXmlEquals(t, xmltvSingleChannelFirstDay, content)
	})
}

func TestGetContentFromUrlWithOffset(t *testing.T) {
	testWithExample(xmltvSingleChannelWeek, func(url string) {
		content, err := getContentForUrl(url, 0, 2, TEST_NOW)
		assert.Nil(t, err)
		assertXmlEquals(t, xmltvSingleChannelSkipFirstTwoDays, content)
	})
}

func TestGetContentFromUrlWithOffsetAndLimit(t *testing.T) {
	testWithExample(xmltvSingleChannelWeek, func(url string) {
		content, err := getContentForUrl(url, 1, 1, TEST_NOW)
		assert.Nil(t, err)
		assertXmlEquals(t, xmltvSingleChannelSecondDay, content)
	})
}

func TestGetError(t *testing.T) {
	testWithErrorCode(http.StatusInternalServerError, func(url string) {
		_, err := getContentForUrl(url, 0, 0, TEST_NOW)
		assert.NotNil(t, err)
	})
}

func testWithExample(example string, testFunction func(string)) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, example)
	}))
	testFunction(ts.URL)
	defer ts.Close()
}

func testWithErrorCode(statusCode int, testFunction func(string)) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
	testFunction(ts.URL)
	defer ts.Close()
}

func assertXmlEquals(t *testing.T, expected string, actual string) {
	expectedNormalized, err := normalizeXmlTv(expected)
	if err != nil {
		t.Errorf("Error normalizing expected XML: %v", err)
		return
	}
	actualNormalized, err := normalizeXmlTv(actual)
	if err != nil {
		t.Errorf("Error normalizing actual XML: %v", err)
		return
	}
	assert.Equal(t, expectedNormalized, actualNormalized)
}

func normalizeXmlTv(xmlString string) (string, error) {
	var epg xmltv.EPG
	if err := xml.Unmarshal([]byte(xmlString), &epg); err != nil {
		return "", err
	}
	return serializeXml(epg)
}
