package common

import (
	"fmt"
	"time"
)

type TimeZone int

const (
	UTC TimeZone = iota

	// North America
	AmericaNewYork
	AmericaChicago
	AmericaDenver
	AmericaLosAngeles
	AmericaAnchorage
	AmericaAdak
	PacificHonolulu
	AmericaPhoenix
	AmericaToronto
	AmericaVancouver
	AmericaMexicoCity

	// South America
	AmericaSaoPaulo
	AmericaBuenosAires
	AmericaSantiago
	AmericaBogota
	AmericaLima

	// Europe
	EuropeLondon
	EuropeParis
	EuropeBerlin
	EuropeMadrid
	EuropeRome
	EuropeMoscow
	EuropeAthens
	EuropeIstanbul

	// Asia
	AsiaDubai
	AsiaKolkata
	AsiaShanghai
	AsiaTokyo
	AsiaSeoul
	AsiaSingapore
	AsiaHongKong
	AsiaBangkok
	AsiaJakarta
	AsiaKarachi
	AsiaTelAviv

	// Australia & Pacific
	AustraliaSydney
	AustraliaMelbourne
	AustraliaPerth
	AustraliaBrisbane
	PacificAuckland

	// Africa
	AfricaCairo
	AfricaJohannesburg
	AfricaLagos
	AfricaNairobi
	AfricaCasablanca
)

func (tz TimeZone) String() string {
	return timeZoneDetails[tz].DisplayName
}

func (tz TimeZone) ID() string {
	return timeZoneDetails[tz].id
}

func (tz TimeZone) Timezone() string {
	return timeZoneDetails[tz].Timezone
}

func (tz TimeZone) ToDetail() TimeZoneDetail {
	return timeZoneDetails[tz]
}

func (tz TimeZone) Location() (*time.Location, error) {
	return time.LoadLocation(tz.ID())
}

type TimeZoneDetail struct {
	id          string
	DisplayName string
	Timezone    string // Format like "UTC+2", "UTC-5", etc.
}

var timeZoneDetails = map[TimeZone]TimeZoneDetail{
	UTC: {id: "UTC", DisplayName: "Coordinated Universal Time", Timezone: "UTC+0"},

	// North America
	AmericaNewYork:    {id: "America/New_York", DisplayName: "Eastern Time (US & Canada)", Timezone: "UTC-5/UTC-4"},
	AmericaChicago:    {id: "America/Chicago", DisplayName: "Central Time (US & Canada)", Timezone: "UTC-6/UTC-5"},
	AmericaDenver:     {id: "America/Denver", DisplayName: "Mountain Time (US & Canada)", Timezone: "UTC-7/UTC-6"},
	AmericaLosAngeles: {id: "America/Los_Angeles", DisplayName: "Pacific Time (US & Canada)", Timezone: "UTC-8/UTC-7"},
	AmericaAnchorage:  {id: "America/Anchorage", DisplayName: "Alaska", Timezone: "UTC-9/UTC-8"},
	AmericaAdak:       {id: "America/Adak", DisplayName: "Hawaii-Aleutian", Timezone: "UTC-10/UTC-9"},
	PacificHonolulu:   {id: "Pacific/Honolulu", DisplayName: "Hawaii", Timezone: "UTC-10"},
	AmericaPhoenix:    {id: "America/Phoenix", DisplayName: "Arizona", Timezone: "UTC-7"},
	AmericaToronto:    {id: "America/Toronto", DisplayName: "Toronto", Timezone: "UTC-5/UTC-4"},
	AmericaVancouver:  {id: "America/Vancouver", DisplayName: "Vancouver", Timezone: "UTC-8/UTC-7"},
	AmericaMexicoCity: {id: "America/Mexico_City", DisplayName: "Mexico City", Timezone: "UTC-6/UTC-5"},

	// South America
	AmericaSaoPaulo:    {id: "America/Sao_Paulo", DisplayName: "São Paulo", Timezone: "UTC-3"},
	AmericaBuenosAires: {id: "America/Argentina/Buenos_Aires", DisplayName: "Buenos Aires", Timezone: "UTC-3"},
	AmericaSantiago:    {id: "America/Santiago", DisplayName: "Santiago", Timezone: "UTC-4/UTC-3"},
	AmericaBogota:      {id: "America/Bogota", DisplayName: "Bogotá", Timezone: "UTC-5"},
	AmericaLima:        {id: "America/Lima", DisplayName: "Lima", Timezone: "UTC-5"},

	// Europe
	EuropeLondon:   {id: "Europe/London", DisplayName: "London", Timezone: "UTC+0/UTC+1"},
	EuropeParis:    {id: "Europe/Paris", DisplayName: "Paris, Central European Time", Timezone: "UTC+1/UTC+2"},
	EuropeBerlin:   {id: "Europe/Berlin", DisplayName: "Berlin", Timezone: "UTC+1/UTC+2"},
	EuropeMadrid:   {id: "Europe/Madrid", DisplayName: "Madrid", Timezone: "UTC+1/UTC+2"},
	EuropeRome:     {id: "Europe/Rome", DisplayName: "Rome", Timezone: "UTC+1/UTC+2"},
	EuropeMoscow:   {id: "Europe/Moscow", DisplayName: "Moscow", Timezone: "UTC+3"},
	EuropeAthens:   {id: "Europe/Athens", DisplayName: "Athens", Timezone: "UTC+2/UTC+3"},
	EuropeIstanbul: {id: "Europe/Istanbul", DisplayName: "Istanbul", Timezone: "UTC+3"},

	// Asia
	AsiaDubai:     {id: "Asia/Dubai", DisplayName: "Dubai", Timezone: "UTC+4"},
	AsiaKolkata:   {id: "Asia/Kolkata", DisplayName: "Mumbai, New Delhi", Timezone: "UTC+5:30"},
	AsiaShanghai:  {id: "Asia/Shanghai", DisplayName: "Beijing, Shanghai", Timezone: "UTC+8"},
	AsiaTokyo:     {id: "Asia/Tokyo", DisplayName: "Tokyo", Timezone: "UTC+9"},
	AsiaSeoul:     {id: "Asia/Seoul", DisplayName: "Seoul", Timezone: "UTC+9"},
	AsiaSingapore: {id: "Asia/Singapore", DisplayName: "Singapore", Timezone: "UTC+8"},
	AsiaHongKong:  {id: "Asia/Hong_Kong", DisplayName: "Hong Kong", Timezone: "UTC+8"},
	AsiaBangkok:   {id: "Asia/Bangkok", DisplayName: "Bangkok", Timezone: "UTC+7"},
	AsiaJakarta:   {id: "Asia/Jakarta", DisplayName: "Jakarta", Timezone: "UTC+7"},
	AsiaKarachi:   {id: "Asia/Karachi", DisplayName: "Karachi", Timezone: "UTC+5"},
	AsiaTelAviv:   {id: "Asia/Tel_Aviv", DisplayName: "Tel Aviv", Timezone: "UTC+2/UTC+3"},

	// Australia & Pacific
	AustraliaSydney:    {id: "Australia/Sydney", DisplayName: "Sydney", Timezone: "UTC+10/UTC+11"},
	AustraliaMelbourne: {id: "Australia/Melbourne", DisplayName: "Melbourne", Timezone: "UTC+10/UTC+11"},
	AustraliaPerth:     {id: "Australia/Perth", DisplayName: "Perth", Timezone: "UTC+8"},
	AustraliaBrisbane:  {id: "Australia/Brisbane", DisplayName: "Brisbane", Timezone: "UTC+10"},
	PacificAuckland:    {id: "Pacific/Auckland", DisplayName: "Auckland", Timezone: "UTC+12/UTC+13"},

	// Africa
	AfricaCairo:        {id: "Africa/Cairo", DisplayName: "Cairo", Timezone: "UTC+2"},
	AfricaJohannesburg: {id: "Africa/Johannesburg", DisplayName: "Johannesburg", Timezone: "UTC+2"},
	AfricaLagos:        {id: "Africa/Lagos", DisplayName: "Lagos", Timezone: "UTC+1"},
	AfricaNairobi:      {id: "Africa/Nairobi", DisplayName: "Nairobi", Timezone: "UTC+3"},
	AfricaCasablanca:   {id: "Africa/Casablanca", DisplayName: "Casablanca", Timezone: "UTC+0/UTC+1"},
}

func AllTimeZones() []TimeZone {
	var all []TimeZone
	for tz := range timeZoneDetails {
		all = append(all, tz)
	}
	return all
}

func (tz TimeZone) GetCurrentOffset() (string, error) {
	loc, err := tz.Location()
	if err != nil {
		return "", err
	}

	_, offset := time.Now().In(loc).Zone()

	hours := offset / 3600
	minutes := (offset % 3600) / 60

	if minutes == 0 {
		if hours >= 0 {
			return fmt.Sprintf("UTC+%d", hours), nil
		}
		return fmt.Sprintf("UTC%d", hours), nil // Negative sign already included in hours
	}

	if hours >= 0 {
		return fmt.Sprintf("UTC+%d:%02d", hours, minutes), nil
	}
	return fmt.Sprintf("UTC%d:%02d", hours, minutes), nil // Negative sign already included in hours
}

func GetTimeZone(gmt string) TimeZone {
	if gmt == "Unknown" {
		return UTC
	}

	offsetMap := make(map[string][]TimeZone)

	for tz := range timeZoneDetails {
		_, err := tz.GetCurrentOffset()
		if err == nil {
			if len(gmt) >= 3 && gmt[0:3] == "GMT" {
				utcOffset := "UTC" + gmt[3:]
				offsetMap[utcOffset] = append(offsetMap[utcOffset], tz)
			}
		}
	}

	utcFormat := ""
	if len(gmt) >= 3 && gmt[0:3] == "GMT" {
		utcFormat = "UTC" + gmt[3:]
	} else {
		return UTC // Return UTC for malformed input
	}

	if zones, found := offsetMap[utcFormat]; found && len(zones) > 0 {
		return zones[0]
	}

	return UTC
}
