package pirsch

import (
	"fmt"
	"strings"
	"time"
)

const (
	// PlatformDesktop filters for everything on desktops.
	PlatformDesktop = "desktop"

	// PlatformMobile filters for everything on mobile devices.
	PlatformMobile = "mobile"

	// PlatformUnknown filters for everything where the platform is unspecified.
	PlatformUnknown = "unknown"
)

// NullClient is a placeholder for no client (0).
var NullClient = int64(0)

// Filter are all fields that can be used to filter the result sets.
// Fields can be inverted by adding a "!" in front of the string.
type Filter struct {
	// ClientID is the optional.
	ClientID int64

	// Timezone sets the timezone used to interpret dates and times.
	// It will be set to UTC by default.
	Timezone *time.Location

	// From is the start date of the selected period.
	From time.Time

	// To is the end date of the selected period.
	To time.Time

	// Day is an exact match for the result set ("on this day").
	Day time.Time

	// Start is the start date and time of the selected period.
	Start time.Time

	// Path filters for the path.
	// Note that if this and PathPattern are both set, Path will be preferred.
	Path string

	// PathPattern filters for the path using a (ClickHouse supported) regex pattern.
	// Note that if this and Path are both set, Path will be preferred.
	// Examples for useful patterns (all case-insensitive, * is used for every character but slashes, ** is used for all characters including slashes):
	//  (?i)^/path/[^/]+$ // matches /path/*
	//  (?i)^/path/[^/]+/.* // matches /path/*/**
	//  (?i)^/path/[^/]+/slashes$ // matches /path/*/slashes
	//  (?i)^/path/.+/slashes$ // matches /path/**/slashes
	PathPattern string

	// Language filters for the ISO language code.
	Language string

	// Country filters for the ISO country code.
	Country string

	// Referrer filters for the referrer.
	Referrer string

	// OS filters for the operating system.
	OS string

	// OSVersion filters for the operating system version.
	OSVersion string

	// Browser filters for the browser.
	Browser string

	// BrowserVersion filters for the browser version.
	BrowserVersion string

	// Platform filters for the platform (desktop, mobile, unknown).
	Platform string

	// ScreenClass filters for the screen class.
	ScreenClass string

	// UTMSource filters for the utm_source query parameter.
	UTMSource string

	// UTMMedium filters for the utm_medium query parameter.
	UTMMedium string

	// UTMCampaign filters for the utm_campaign query parameter.
	UTMCampaign string

	// UTMContent filters for the utm_content query parameter.
	UTMContent string

	// UTMTerm filters for the utm_term query parameter.
	UTMTerm string

	// EventName filters for an event by its name.
	EventName string

	// EventMetaKey filters for an event meta key.
	// This must be used together with an EventName.
	EventMetaKey string

	// Limit limits the number of results. Less or equal to zero means no limit.
	Limit int

	// IncludeTitle indicates whether the Analyzer.Pages, Analyzer.EntryPages, and Analyzer.ExitPages should contain the page title or not.
	IncludeTitle bool

	// IncludeAvgTimeOnPage indicates whether Analyzer.Pages and Analyzer.EntryPages should contain the average time on page or not.
	IncludeAvgTimeOnPage bool

	// MaxTimeOnPageSeconds is an optional maximum for the time spent on page.
	// Visitors who are idle artificially increase the average time spent on a page, this option can be used to limit the effect.
	// Set to 0 to disable this option (default).
	MaxTimeOnPageSeconds int
}

// NewFilter creates a new filter for given client ID.
func NewFilter(clientID int64) *Filter {
	return &Filter{
		ClientID: clientID,
	}
}

func (filter *Filter) validate() {
	if filter.Timezone == nil {
		filter.Timezone = time.UTC
	}

	if !filter.From.IsZero() {
		filter.From = filter.toDate(filter.From)
	} else {
		filter.From = filter.From.In(time.UTC)
	}

	if !filter.To.IsZero() {
		filter.To = filter.toDate(filter.To)
	} else {
		filter.To = filter.To.In(time.UTC)
	}

	if !filter.Day.IsZero() {
		filter.Day = filter.toDate(filter.Day)
	} else {
		filter.Day = filter.Day.In(time.UTC)
	}

	if !filter.Start.IsZero() {
		filter.Start = time.Date(filter.Start.Year(), filter.Start.Month(), filter.Start.Day(), filter.Start.Hour(), filter.Start.Minute(), filter.Start.Second(), 0, time.UTC)
	}

	if !filter.To.IsZero() && filter.From.After(filter.To) {
		filter.From, filter.To = filter.To, filter.From
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if !filter.To.IsZero() && filter.To.After(today) {
		filter.To = today
	}

	if filter.Path != "" && filter.PathPattern != "" {
		filter.PathPattern = ""
	}

	if filter.Limit < 0 {
		filter.Limit = 0
	}
}

func (filter *Filter) table() string {
	if filter.EventName != "" {
		return "event"
	}

	return "hit"
}

func (filter *Filter) queryTime() ([]interface{}, string) {
	args := make([]interface{}, 0, 5)
	args = append(args, filter.ClientID)
	timezone := filter.Timezone.String()
	var sqlQuery strings.Builder
	sqlQuery.WriteString("client_id = ? ")

	if !filter.From.IsZero() {
		args = append(args, filter.From)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') >= toDate(?, '%s') ", timezone, timezone))
	}

	if !filter.To.IsZero() {
		args = append(args, filter.To)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') <= toDate(?, '%s') ", timezone, timezone))
	}

	if !filter.Day.IsZero() {
		args = append(args, filter.Day)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') = toDate(?, '%s') ", timezone, timezone))
	}

	if !filter.Start.IsZero() {
		args = append(args, filter.Start)
		sqlQuery.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') >= toDateTime(?, '%s') ", timezone, timezone))
	}

	return args, sqlQuery.String()
}

func (filter *Filter) queryFields() ([]interface{}, string) {
	args := make([]interface{}, 0, 16)
	fields := make([]string, 0, 16)
	filter.appendQuery(&fields, &args, "path", filter.Path)
	filter.appendQuery(&fields, &args, "language", filter.Language)
	filter.appendQuery(&fields, &args, "country_code", filter.Country)
	filter.appendQuery(&fields, &args, "referrer", filter.Referrer)
	filter.appendQuery(&fields, &args, "os", filter.OS)
	filter.appendQuery(&fields, &args, "os_version", filter.OSVersion)
	filter.appendQuery(&fields, &args, "browser", filter.Browser)
	filter.appendQuery(&fields, &args, "browser_version", filter.BrowserVersion)
	filter.appendQuery(&fields, &args, "screen_class", filter.ScreenClass)
	filter.appendQuery(&fields, &args, "utm_source", filter.UTMSource)
	filter.appendQuery(&fields, &args, "utm_medium", filter.UTMMedium)
	filter.appendQuery(&fields, &args, "utm_campaign", filter.UTMCampaign)
	filter.appendQuery(&fields, &args, "utm_content", filter.UTMContent)
	filter.appendQuery(&fields, &args, "utm_term", filter.UTMTerm)
	filter.appendQuery(&fields, &args, "event_name", filter.EventName)

	if filter.Platform != "" {
		if strings.HasPrefix(filter.Platform, "!") {
			platform := filter.Platform[1:]

			if platform == PlatformDesktop {
				fields = append(fields, "desktop != 1 ")
			} else if platform == PlatformMobile {
				fields = append(fields, "mobile != 1 ")
			} else {
				fields = append(fields, "(desktop = 1 OR mobile = 1) ")
			}
		} else {
			if filter.Platform == PlatformDesktop {
				fields = append(fields, "desktop = 1 ")
			} else if filter.Platform == PlatformMobile {
				fields = append(fields, "mobile = 1 ")
			} else {
				fields = append(fields, "desktop = 0 AND mobile = 0 ")
			}
		}
	}

	if filter.PathPattern != "" {
		if strings.HasPrefix(filter.PathPattern, "!") {
			args = append(args, filter.PathPattern[1:])
			fields = append(fields, `match("path", ?) = 0`)
		} else {
			args = append(args, filter.PathPattern)
			fields = append(fields, `match("path", ?) = 1`)
		}
	}

	return args, strings.Join(fields, "AND ")
}

func (filter *Filter) withFill() ([]interface{}, string) {
	if !filter.From.IsZero() && !filter.To.IsZero() {
		timezone := filter.Timezone.String()
		return []interface{}{filter.From, filter.To}, fmt.Sprintf("WITH FILL FROM toDate(?, '%s') TO toDate(?, '%s')+1 ", timezone, timezone)
	}

	return nil, ""
}

func (filter *Filter) withLimit() string {
	if filter.Limit > 0 {
		return fmt.Sprintf("LIMIT %d ", filter.Limit)
	}

	return ""
}

func (filter *Filter) query() ([]interface{}, string) {
	args, query := filter.queryTime()
	fieldArgs, queryFields := filter.queryFields()
	args = append(args, fieldArgs...)

	if queryFields != "" {
		query += "AND " + queryFields
	}

	return args, query
}

func (filter *Filter) appendQuery(fields *[]string, args *[]interface{}, field, value string) {
	if value != "" {
		if strings.HasPrefix(value, "!") {
			*args = append(*args, value[1:])
			*fields = append(*fields, fmt.Sprintf("%s != ? ", field))
		} else {
			*args = append(*args, value)
			*fields = append(*fields, fmt.Sprintf("%s = ? ", field))
		}
	}
}

func (filter *Filter) toDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func (filter *Filter) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}
