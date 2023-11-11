package doc

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

const referenceLanguage = "en"

var docFiles []Document

// File represents the single documentation file of a determined language.
type File struct {
	IsUpdated      bool
	Title          string
	referencesFile string
}

// Data retrieves data from file's actual file on disk.
func (f File) Data() (string, error) {
	data, err := os.ReadFile(f.referencesFile)
	updateIPs()
	res := strings.NewReplacer(
		"{ipmain}", ipMain,
		"{ipmirror}", ipMirror,
	).Replace(string(data))
	return res, err
}

// Document represents a documentation file, providing its old ID, its slug,
// and all its variations in the various languages.
type Document struct {
	Slug      string
	OldID     int
	Languages map[string]File
}

// File retrieves a Document's File based on the passed language, and returns
// the values for the referenceLanguage (en) if in the passed language they are
// not available
func (d Document) File(lang string) File {
	if vals, ok := d.Languages[lang]; ok {
		return vals
	}
	return d.Languages[referenceLanguage]
}

// LanguageDoc has the only purpose to be returned by GetDocs.
type LanguageDoc struct {
	Title string
	Slug  string
}

// GetDocs retrieves a list of documents in a certain language, with titles and
// slugs.
func GetDocs(lang string) []LanguageDoc {
	var docs []LanguageDoc

	for _, file := range docFiles {
		docs = append(docs, LanguageDoc{
			Slug:  file.Slug,
			Title: file.File(lang).Title,
		})
	}

	return docs
}

// SlugFromOldID gets a doc file's slug from its old ID
func SlugFromOldID(i int) string {
	for _, d := range docFiles {
		if d.OldID == i {
			return d.Slug
		}
	}

	return ""
}

// GetFile retrieves a file, given a slug and a language.
func GetFile(slug, language string) File {
	for _, f := range docFiles {
		if f.Slug != slug {
			continue
		}
		if val, ok := f.Languages[language]; ok {
			return val
		}
		return f.Languages[referenceLanguage]
	}
	return File{}
}

var (
	ipMain        = "144.217.254.156"
	ipMirror      = "144.217.254.156"
	ipLastUpdated = time.Date(2018, 5, 13, 11, 45, 0, 0, time.UTC)
	ipRegex       = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
)

func updateIPs() {
	if time.Since(ipLastUpdated) < time.Hour*24*14 {
		return
	}
	ipLastUpdated = time.Now()

	resp, err := http.Get("https://akatsuki.gg/static/current.json")
	if err != nil {
		slog.Error("error updating IPs", "error", err.Error())
		return
	}

	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		slog.Error("error updating IPs", "error", err.Error())
		return
	}

	ips := strings.SplitN(string(data), "\n", 3)
	if len(ips) < 2 || !ipRegex.MatchString(ips[0]) || !ipRegex.MatchString(ips[1]) {
		return
	}
	ipMain = ips[0]
	ipMirror = ips[1]
}

func init() {
	go updateIPs()
}
