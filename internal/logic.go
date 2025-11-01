package internal

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sherif-fanous/xmltv"
)

const MAX_LENGTH_IN_DAYS = 7
const MIN_LENGTH_IN_DAYS = 1
const MIN_OFFSET_IN_DAYS = 0
const ONE_DAY = time.Hour * 24

type Logic struct {
	GetDescriptionFunc  func() string
	GetCapabilitiesFunc func() []string
	GetContentFunc      func(request Request) (string, error)
}

type Request struct {
	URLs         []string
	LengthInDays int
	OffsetInDays int
	Quiet        bool
	GetNowFunc   func() time.Time
}

func GetLogic() Logic {
	return Logic{
		GetDescriptionFunc:  getDescription,
		GetCapabilitiesFunc: getCapabilities,
		GetContentFunc:      getContent,
	}
}

func getDescription() string {
	return "Tvheadend XMLTV URL Generator"
}

func getCapabilities() []string {
	return []string{"baseline"}
}

func getContent(r Request) (string, error) {
	xmlDocuments := make([]xmltv.EPG, 0, len(r.URLs))
	for _, url := range r.URLs {
		xmlDocument, err := getContentForUrl(url, r.LengthInDays, r.OffsetInDays, r.GetNowFunc())
		if err != nil {
			return "", err
		}
		xmlDocuments = append(xmlDocuments, xmlDocument)
	}

	mergedDocument, err := mergeXml(xmlDocuments, r.GetNowFunc())
	if err != nil {
		return "", err
	}

	return serializeXml(mergedDocument)
}

func mergeXml(xmlDocuments []xmltv.EPG, now time.Time) (xmltv.EPG, error) {
	result := xmltv.EPG{
		Date:       &xmltv.Time{Time: now},
		Channels:   make([]xmltv.Channel, 0),
		Programmes: make([]xmltv.Programme, 0),
	}

	for _, doc := range xmlDocuments {
		result.Channels = append(result.Channels, doc.Channels...)
		result.Programmes = append(result.Programmes, doc.Programmes...)
	}

	return result, nil
}

func timeIsInRange(xmltvTime *xmltv.Time, start time.Time, end time.Time) bool {
	time := xmltvTime.Time
	return time.After(start) && time.Before(end)
}

func programmeIsInRange(programme xmltv.Programme, start time.Time, end time.Time) bool {
	startIsInRange := timeIsInRange(&programme.Start, start, end)
	stopIsInRange := timeIsInRange(programme.Stop, start, end)
	fullOverlap := programme.Start.Time.Before(start) && programme.Stop.After(end)
	return startIsInRange || stopIsInRange || fullOverlap
}

func getContentForUrl(url string, lengthInDays int, offsetInDays int, now time.Time) (xmltv.EPG, error) {
	var epg xmltv.EPG

	body, err := downloadXml(url)
	if err != nil {
		return epg, err
	}

	// Parse body into xmltv structure
	if err := xml.Unmarshal(body, &epg); err != nil {
		return epg, err
	}

	// Remove all epg.Programme entries where the whole programme is outside the requested range
	earliest, latest := calculateProgrammeTimeRange(now, offsetInDays, lengthInDays)
	filtered := make([]xmltv.Programme, 0, len(epg.Programmes))
	for _, p := range epg.Programmes {
		if programmeIsInRange(p, earliest, latest) {
			filtered = append(filtered, p)
			continue
		}
	}
	epg.Programmes = filtered

	// Marshal back to XML
	return epg, nil
}

func downloadXml(url string) ([]byte, error) {
	// send GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	// read body from response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// decode gzip if needed
	if len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b { // Gzip magic number
		content, err := unzipGzipData(body)
		if err != nil {
			return nil, err
		}
		return content, nil
	}

	// return body as is
	return body, nil
}

func unzipGzipData(data []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(data)
	gzipReader, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()
	return io.ReadAll(gzipReader)
}

func serializeXml(epg xmltv.EPG) (string, error) {
	var header []byte
	header = fmt.Appendln(header, `<?xml version="1.0" encoding="UTF-8"?>`)
	header = fmt.Appendln(header, `<!DOCTYPE tv SYSTEM "xmltv.dtd">`)
	out, err := xml.MarshalIndent(epg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(header) + string(out), nil
}

func calculateProgrammeTimeRange(now time.Time, offsetInDays int, lengthInDays int) (time.Time, time.Time) {
	offsetDuration := time.Duration(max(offsetInDays, MIN_OFFSET_IN_DAYS)) * ONE_DAY
	lengthDuration := time.Duration(MAX_LENGTH_IN_DAYS) * ONE_DAY
	if lengthInDays > 0 {
		lengthDuration = time.Duration(max(lengthInDays, MIN_LENGTH_IN_DAYS)) * ONE_DAY
	}
	earliest := now.Truncate(ONE_DAY).Add(offsetDuration)
	latest := earliest.Add(lengthDuration)
	return earliest, latest
}
