package player

import (
	"os"
	"strings"
	"testing"
)

func TestLanguageCodes(t *testing.T) {
	tests := map[string]string{
		"en":    "en,eng",
		"fa":    "fa,fas,per",
		"zh-tw": "chi,zh,zho",
		"pt_BR": "por,pt",
		"nb":    "nb,no,nob,nor",
		"xx":    "",
	}
	for locale, want := range tests {
		if got := LanguageCodes(locale); got != want {
			t.Errorf("LanguageCodes(%q) = %q, want %q", locale, got, want)
		}
	}
}

// Every language the interface is translated into can pick audio.
func TestLanguageCodesCoverLocales(t *testing.T) {
	files, err := os.ReadDir("../app/locales")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		locale := strings.TrimSuffix(f.Name(), ".json")
		if LanguageCodes(locale) == "" {
			t.Errorf("no language codes for locale %q", locale)
		}
	}
}
