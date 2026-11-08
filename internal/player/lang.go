package player

import (
	"sort"
	"strings"

	"github.com/xmhrnx/giode/i18n"
)

// languages maps ISO 639-1 and 639-2 (B and T) codes to English names,
// for the languages track tags commonly use.
var languages = map[string]string{}

func init() {
	for _, l := range [][]string{
		{"English", "en", "eng"}, {"Arabic", "ar", "ara"}, {"Bengali", "bn", "ben"},
		{"Bulgarian", "bg", "bul"}, {"Catalan", "ca", "cat"}, {"Chinese", "zh", "chi", "zho"},
		{"Croatian", "hr", "hrv"}, {"Czech", "cs", "cze", "ces"}, {"Danish", "da", "dan"},
		{"Dutch", "nl", "dut", "nld"}, {"Estonian", "et", "est"}, {"Finnish", "fi", "fin"},
		{"French", "fr", "fre", "fra"}, {"German", "de", "ger", "deu"}, {"Greek", "el", "gre", "ell"},
		{"Hebrew", "he", "heb"}, {"Hindi", "hi", "hin"}, {"Hungarian", "hu", "hun"},
		{"Icelandic", "is", "ice", "isl"}, {"Indonesian", "id", "ind"}, {"Italian", "it", "ita"},
		{"Japanese", "ja", "jpn"}, {"Korean", "ko", "kor"}, {"Latvian", "lv", "lav"},
		{"Lithuanian", "lt", "lit"}, {"Malay", "ms", "may", "msa"}, {"Norwegian", "no", "nor"},
		{"Norwegian Bokmål", "nb", "nob"}, {"Persian", "fa", "per", "fas"}, {"Polish", "pl", "pol"},
		{"Portuguese", "pt", "por"}, {"Romanian", "ro", "rum", "ron"}, {"Russian", "ru", "rus"},
		{"Serbian", "sr", "srp"}, {"Slovak", "sk", "slo", "slk"}, {"Slovenian", "sl", "slv"},
		{"Spanish", "es", "spa"}, {"Swedish", "sv", "swe"}, {"Tamil", "ta", "tam"},
		{"Telugu", "te", "tel"}, {"Thai", "th", "tha"}, {"Turkish", "tr", "tur"},
		{"Ukrainian", "uk", "ukr"}, {"Urdu", "ur", "urd"}, {"Vietnamese", "vi", "vie"},
		{"Filipino", "fil", "tl", "tgl"}, {"Kurdish", "ku", "kur"}, {"Swahili", "sw", "swa"},
		{"Armenian", "hy", "arm", "hye"}, {"Georgian", "ka", "geo", "kat"}, {"Gujarati", "gu", "guj"},
		{"Kannada", "kn", "kan"}, {"Malayalam", "ml", "mal"}, {"Marathi", "mr", "mar"},
		{"Punjabi", "pa", "pan"},
		{"Undetermined", "und"},
	} {
		for _, code := range l[1:] {
			languages[code] = l[0]
		}
	}
}

// LanguageName returns the name of a language code in the interface's
// language, or the code itself when unknown. Regional tags ("pt-BR")
// keep their region.
func LanguageName(code string) string {
	base, region, _ := strings.Cut(strings.ToLower(code), "-")
	name, ok := languages[base]
	if !ok {
		return code
	}
	name = i18n.Intl(name)
	if region != "" {
		return name + " (" + strings.ToUpper(region) + ")"
	}
	return name
}

// LanguageCodes is an mpv language list for locale ("pt", "zh-tw"): every
// code tracks of its language may be tagged with, or "" when unknown.
func LanguageCodes(locale string) string {
	base, _, _ := strings.Cut(strings.ReplaceAll(strings.ToLower(locale), "_", "-"), "-")
	name, ok := languages[base]
	if !ok {
		return ""
	}
	var codes []string
	for code, n := range languages {
		// Norwegian tracks are as often tagged plain "nor" as Bokmål.
		if n == name || (name == "Norwegian Bokmål" && n == "Norwegian") {
			codes = append(codes, code)
		}
	}
	sort.Strings(codes)
	return strings.Join(codes, ",")
}
