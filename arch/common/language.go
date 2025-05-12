package common

type Language int

const (
	English Language = iota
	Spanish
	French
	German
	Italian
	Portuguese
	Russian
	Chinese
	Japanese
	Korean
	Hindi
	Arabic
	Turkish
	Dutch
	Swedish
	Norwegian
	Danish
	Finnish
	Greek
	Hebrew
	Thai
	Vietnamese
	Indonesian
	Malay
	Polish
	Ukrainian
	Czech
	Hungarian
	Romanian
	Bulgarian
)

type LanguageDetail struct {
	id          string
	DisplayName string
	NativeName  string
}

var languageDetails = map[Language]LanguageDetail{
	English:    {id: "en", DisplayName: "English", NativeName: "English"},
	Spanish:    {id: "es", DisplayName: "Spanish", NativeName: "Español"},
	French:     {id: "fr", DisplayName: "French", NativeName: "Français"},
	German:     {id: "de", DisplayName: "German", NativeName: "Deutsch"},
	Italian:    {id: "it", DisplayName: "Italian", NativeName: "Italiano"},
	Portuguese: {id: "pt", DisplayName: "Portuguese", NativeName: "Português"},
	Russian:    {id: "ru", DisplayName: "Russian", NativeName: "Русский"},
	Chinese:    {id: "zh", DisplayName: "Chinese", NativeName: "中文"},
	Japanese:   {id: "ja", DisplayName: "Japanese", NativeName: "日本語"},
	Korean:     {id: "ko", DisplayName: "Korean", NativeName: "한국어"},
	Hindi:      {id: "hi", DisplayName: "Hindi", NativeName: "हिन्दी"},
	Arabic:     {id: "ar", DisplayName: "Arabic", NativeName: "العربية"},
	Turkish:    {id: "tr", DisplayName: "Turkish", NativeName: "Türkçe"},
	Dutch:      {id: "nl", DisplayName: "Dutch", NativeName: "Nederlands"},
	Swedish:    {id: "sv", DisplayName: "Swedish", NativeName: "Svenska"},
	Norwegian:  {id: "no", DisplayName: "Norwegian", NativeName: "Norsk"},
	Danish:     {id: "da", DisplayName: "Danish", NativeName: "Dansk"},
	Finnish:    {id: "fi", DisplayName: "Finnish", NativeName: "Suomi"},
	Greek:      {id: "el", DisplayName: "Greek", NativeName: "Ελληνικά"},
	Hebrew:     {id: "he", DisplayName: "Hebrew", NativeName: "עברית"},
	Thai:       {id: "th", DisplayName: "Thai", NativeName: "ไทย"},
	Vietnamese: {id: "vi", DisplayName: "Vietnamese", NativeName: "Tiếng Việt"},
	Indonesian: {id: "id", DisplayName: "Indonesian", NativeName: "Bahasa Indonesia"},
	Malay:      {id: "ms", DisplayName: "Malay", NativeName: "Bahasa Melayu"},
	Polish:     {id: "pl", DisplayName: "Polish", NativeName: "Polski"},
	Ukrainian:  {id: "uk", DisplayName: "Ukrainian", NativeName: "Українська"},
	Czech:      {id: "cs", DisplayName: "Czech", NativeName: "Čeština"},
	Hungarian:  {id: "hu", DisplayName: "Hungarian", NativeName: "Magyar"},
	Romanian:   {id: "ro", DisplayName: "Romanian", NativeName: "Română"},
	Bulgarian:  {id: "bg", DisplayName: "Bulgarian", NativeName: "Български"},
}

func (l Language) String() string {
	return languageDetails[l].DisplayName
}

func (l Language) ID() string {
	return languageDetails[l].id
}

func (l Language) NativeName() string {
	return languageDetails[l].NativeName
}

func (l Language) ToDetail() LanguageDetail {
	return languageDetails[l]
}

func AllLanguages() []Language {
	var all []Language
	for lang := range languageDetails {
		all = append(all, lang)
	}
	return all
}

func GetLanguageByID(id string) Language {
	for lang, detail := range languageDetails {
		if detail.id == id {
			return lang
		}
	}
	return English // Default to English if not found
}
