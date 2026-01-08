package ics

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type Generator interface {
	Generate(event *models.Event, rsvpURL string) ([]byte, error)
}

type generator struct{}

func NewGenerator() Generator {
	return &generator{}
}

func (g *generator) Generate(event *models.Event, rsvpURL string) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//TinyRSVP//EN\r\n")
	buf.WriteString("METHOD:REQUEST\r\n")
	buf.WriteString("BEGIN:VEVENT\r\n")

	uid := fmt.Sprintf("%d@tinyrsvp", event.ID)
	buf.WriteString(fmt.Sprintf("UID:%s\r\n", uid))

	buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICSTime(time.Now())))
	buf.WriteString(fmt.Sprintf("DTSTART;TZID=%s:%s\r\n", event.Timezone, formatICSTime(event.StartTime)))

	if event.EndTime != nil {
		buf.WriteString(fmt.Sprintf("DTEND;TZID=%s:%s\r\n", event.Timezone, formatICSTime(*event.EndTime)))
	}

	buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICS(event.Title)))

	if event.Location != nil {
		buf.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(*event.Location)))
	}

	if event.Description != nil {
		desc := fmt.Sprintf("%s\\n\\nRSVP: %s", *event.Description, rsvpURL)
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(desc)))
	} else {
		desc := fmt.Sprintf("RSVP: %s", rsvpURL)
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(desc)))
	}

	buf.WriteString("STATUS:CONFIRMED\r\n")
	buf.WriteString(fmt.Sprintf("SEQUENCE:%d\r\n", event.ICSSequence))

	buf.WriteString("BEGIN:VALARM\r\n")
	buf.WriteString("TRIGGER:-PT24H\r\n")
	buf.WriteString("ACTION:DISPLAY\r\n")
	buf.WriteString(fmt.Sprintf("DESCRIPTION:Reminder: %s tomorrow\r\n", escapeICS(event.Title)))
	buf.WriteString("END:VALARM\r\n")

	buf.WriteString("END:VEVENT\r\n")
	buf.WriteString("END:VCALENDAR\r\n")

	return buf.Bytes(), nil
}

func formatICSTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

func escapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
