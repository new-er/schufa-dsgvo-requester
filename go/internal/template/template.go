package template

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/new-er/schufa-dsgvo-requester/internal/config"
)

//go:embed mail.txt
var templateFS embed.FS

func Render(cfg *config.Config, configPath string) (string, error) {
	// Try to load custom template next to config file
	customPath := filepath.Join(filepath.Dir(configPath), "mail.custom.txt")
	var tmplContent []byte
	if data, err := os.ReadFile(customPath); err == nil {
		tmplContent = data
	} else {
		// fallback to embedded
		var err error
		tmplContent, err = templateFS.ReadFile("mail.txt")
		if err != nil {
			return "", err
		}
	}
	tmpl, err := template.New("mail").Parse(string(tmplContent))
	if err != nil {
		return "", err
	}
	data := struct {
		Vorname      string
		Nachname     string
		Geburtsdatum string
		Strasse      string
		PLZ          string
		Wohnort      string
		Land         string
	}{
		Vorname:      cfg.Personal.Vorname,
		Nachname:     cfg.Personal.Nachname,
		Geburtsdatum: cfg.Personal.Geburtsdatum,
		Strasse:      cfg.Personal.Strasse,
		PLZ:          cfg.Personal.PLZ,
		Wohnort:      cfg.Personal.Wohnort,
		Land:         cfg.Personal.Land,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Raw returns the raw template source (with placeholders) without rendering.
func Raw(configPath string) (string, error) {
	customPath := filepath.Join(filepath.Dir(configPath), "mail.custom.txt")
	if data, err := os.ReadFile(customPath); err == nil {
		return string(data), nil
	}
	// fallback to embedded
	data, err := templateFS.ReadFile("mail.txt")
	if err != nil {
		return "", err
	}
	return string(data), nil
}