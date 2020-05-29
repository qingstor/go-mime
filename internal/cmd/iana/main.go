package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/panjf2000/ants/v2"
)

//go:generate go-bindata -nometadata -ignore "\\.go$" -prefix tmpl ./tmpl

// Source related URLs, only need update while source missing.
const (
	MailcapMIMETypes         = "https://pagure.io/mailcap/raw/master/f/mime.types"
	MediaTypesRegistry       = "https://www.iana.org/assignments/media-types/media-types.xhtml"
	MediaTypesTemplatePrefix = "https://www.iana.org/assignments/media-types/"
)

var (
	// FileExtRegex is the file extension name regex.
	FileExtRegex = regexp.MustCompile("\\.[\\w]+")

	typesT = newT("types")
)

// Item represent a MIME Item
type Item struct {
	Type    string
	SubType string

	Extensions []string

	InternalName string // internal name for this media type
	TemplateURL  string
}

func (i *Item) String() string {
	return fmt.Sprintf("type: %s, subtype: %s, extensions: %v", i.Type, i.SubType, i.Extensions)
}

// DetectSuffix will try to detect a item's suffix.
func (i *Item) DetectSuffix() {
	if i.TemplateURL == "" {
		return
	}

	resp, err := http.Get(i.TemplateURL)
	if err != nil {
		log.Fatalf("Get media type %s: %v", i.InternalName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// Just ignore those types.
		log.Printf("Get media type %s: %d %s", i.InternalName, resp.StatusCode, resp.Status)
		return
	}

	s := bufio.NewScanner(resp.Body)
	for s.Scan() {
		v := s.Text()
		if !strings.Contains(v, "Extension") && !strings.Contains(v, "extension") {
			continue
		}

		x := FileExtRegex.FindAllString(v, -1)
		for _, v := range x {
			i.Extensions = append(i.Extensions, strings.TrimPrefix(v, "."))
		}
	}
}

func parseMediaTypesRegistry() []*Item {
	// Get content from media types registry
	resp, err := http.Get(MediaTypesRegistry)
	if err != nil {
		log.Fatalf("Get all media types: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("Get all media types: %d %s", resp.StatusCode, resp.Status)
	}

	// Load document.
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Create document: %v", err)
	}

	var items []*Item
	idx := 0

	// Parse all media types.
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		id, ok := s.Attr("id")
		if !ok || !strings.HasPrefix(id, "table-") {
			return
		}

		s.Find("td").Each(func(i int, s *goquery.Selection) {
			value := strings.Trim(s.Text(), "\n ")
			// Remove all \n in value.
			value = strings.ReplaceAll(value, "\n", "")
			switch i % 3 {
			case 0:
				items = append(items, &Item{})
				items[idx].InternalName = value
			case 1:
				if value == "" {
					println(fmt.Sprintf("type %s doesn't have template", items[idx].InternalName))
					return
				}

				types := strings.SplitN(value, "/", 2)
				items[idx].Type, items[idx].SubType = types[0], types[1]

				url, ok := s.Find("a").Attr("href")
				if ok {
					items[idx].TemplateURL = MediaTypesTemplatePrefix + url
				}
			case 2:
				idx++
			}
		})
	})

	return items
}

func parseMailcapMIMETypes() map[string]string {
	resp, err := http.Get(MailcapMIMETypes)
	if err != nil {
		log.Fatalf("Get all media types: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("Get all media types: %d %s", resp.StatusCode, resp.Status)
	}

	m := make(map[string]string)

	s := bufio.NewScanner(resp.Body)
	for s.Scan() {
		v := s.Text()
		// Ignore all comment lines
		if strings.HasPrefix(v, "#") {
			continue
		}

		x := strings.Split(v, "\t")
		mediaType := x[0]
		for _, exts := range x[1:] {
			exts = strings.Trim(exts, "")
			if exts == "" {
				continue
			}
			ext := strings.Split(exts, " ")
			for _, e := range ext {
				if e == "" {
					continue
				}
				m[e] = mediaType
			}
		}
	}
	return m
}

func main() {
	mailcapMIMEMap := parseMailcapMIMETypes()

	items := parseMediaTypesRegistry()

	// Try to detect suffix
	wg := sync.WaitGroup{}
	pool, _ := ants.NewPool(100)
	for _, v := range items {
		v := v

		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()

			v.DetectSuffix()
		})
	}
	wg.Wait()

	// Calculate correct suffix
	m := make(map[string]string)
	// Sync from mailcap first.
	for k, v := range mailcapMIMEMap {
		m[k] = v
	}
	for _, v := range items {
		v := v

		for _, ext := range v.Extensions {
			if ext == "" {
				continue
			}
			if _, ok := mailcapMIMEMap[ext]; ok {
				// We can ignore conflict with mailcap safely.
				continue
			}
			if it, ok := m[ext]; ok {
				fmt.Printf("file extension %s conflict in: %s and %s\n", ext, v, it)
				continue
			}
			m[ext] = fmt.Sprintf("%s/%s", v.Type, v.SubType)
		}
	}

	// Generate code
	generateT(typesT, "generated.go", m)
}
