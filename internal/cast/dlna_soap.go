package cast

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// soapCall calls action on the UPnP service at controlURL of the device
// called name, and returns the response body.
func soapCall(ctx context.Context, name, controlURL, service, action, args string) (string, error) {
	body := `<?xml version="1.0" encoding="utf-8"?>` +
		`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">` +
		`<s:Body><u:` + action + ` xmlns:u="` + service + `">` + args + `</u:` + action + `></s:Body></s:Envelope>`
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, controlURL, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPACTION", `"`+service+"#"+action+`"`)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cast: %s: %w", name, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		msg := tag(string(data), "errorDescription")
		if msg == "" {
			msg = resp.Status
		}
		return "", fmt.Errorf("cast: %s refused %s: %s", name, action, msg)
	}
	return string(data), nil
}

// tag returns the text of the first <name> element in s, ignoring
// namespaces. SOAP responses are small and flat enough for this.
func tag(s, name string) string {
	for _, open := range []string{"<" + name + ">", ":" + name + ">"} {
		i := strings.Index(s, open)
		if i < 0 {
			continue
		}
		rest := s[i+len(open):]
		j := strings.Index(rest, "</")
		if j < 0 {
			return ""
		}
		return html.UnescapeString(strings.TrimSpace(rest[:j]))
	}
	return ""
}

// hms formats seconds as H:MM:SS for UPnP.
func hms(s float64) string {
	t := int(max(0, s))
	return fmt.Sprintf("%d:%02d:%02d", t/3600, t/60%60, t%60)
}

// parseHMS reads UPnP's H:MM:SS(.fff); "NOT_IMPLEMENTED" and the like
// give 0.
func parseHMS(s string) float64 {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	var total float64
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return 0
		}
		total = total*60 + v
	}
	return total
}
