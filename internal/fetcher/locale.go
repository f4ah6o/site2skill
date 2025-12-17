// Package fetcher provides website crawling and downloading functionality.
package fetcher

import (
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// LocaleConfig specifies locale priority and parameter configuration for content negotiation.
// It controls how the fetcher selects localized versions of pages when multiple languages are available.
type LocaleConfig struct {
	// Priority is an ordered list of preferred locales (e.g., ["en", "ja"]).
	// The fetcher will attempt locales in this order when multiple versions exist.
	Priority []string
	// ParamName is the query parameter name used for locale selection (e.g., "hl" for ?hl=ja).
	// If empty, path-based locale detection is used instead (e.g., /ja/docs).
	ParamName string
}

// DefaultLocalePriority is the default locale preference order used when none is specified.
var DefaultLocalePriority = []string{"en", "ja"}

// KnownLocales is a map of recognized locale codes for path-based locale detection.
// It supports both language codes (e.g., "en", "ja") and region-specific codes (e.g., "en-us", "zh-tw").
var KnownLocales = map[string]bool{
	"en": true, "en-us": true, "en-gb": true,
	"ja": true, "ja-jp": true,
	"zh": true, "zh-cn": true, "zh-tw": true, "zh-hk": true,
	"ko": true, "ko-kr": true,
	"de": true, "de-de": true,
	"fr": true, "fr-fr": true,
	"es": true, "es-es": true,
	"it": true, "it-it": true,
	"pt": true, "pt-br": true,
	"ru": true, "ru-ru": true,
	"ar": true, "nl": true, "pl": true, "tr": true,
	"vi": true, "th": true, "id": true, "ms": true,
}

// localePathPattern matches path-based locale patterns in URLs.
// It looks for locale codes enclosed in slashes (e.g., /ja/ or /en-us/) anywhere in the path.
var localePathPattern = regexp.MustCompile(`/([a-z]{2}(?:-[a-zA-Z]{2,4})?)/`)

// ExtractLocale extracts the locale and canonical path from a URL based on the provided configuration.
//
// It supports two locale detection modes:
//   - Query parameter mode: Uses the configured ParamName (e.g., ?hl=ja)
//   - Path mode: Detects locale from the first valid locale segment in the path
//     (e.g., /site/ja/docs -> locale="ja", canonical="/site/docs")
//
// Returns the detected locale and the canonical path without locale information.
// If no locale is detected, locale is empty string and canonical is the full path.
func ExtractLocale(u *url.URL, cfg *LocaleConfig) (locale, canonical string) {
	if u == nil {
		return "", ""
	}

	// クエリパラメータ形式
	if cfg != nil && cfg.ParamName != "" {
		locale = u.Query().Get(cfg.ParamName)
		// canonical はクエリパラメータを除いたパス
		canonical = u.Path
		return locale, canonical
	}

	// パス形式の自動検出
	path := u.Path

	// パス内のすべてのマッチ候補を探す
	matches := localePathPattern.FindAllStringSubmatchIndex(path, -1)

	for _, match := range matches {
		// match[2], match[3] がキャプチャグループ1 (ロケール部分) のインデックス
		localePart := path[match[2]:match[3]]
		potentialLocale := strings.ToLower(localePart) // インデックス部分のスライス

		if KnownLocales[potentialLocale] {
			locale = potentialLocale

			// canonical path の再構築
			// /prefix/ja/docs -> /prefix/docs
			// マッチ全体 (/ja/) を / に置換するのは不正確 (複数のスラッシュが絡むため)
			// 単純に /locale/ を / に置換する

			// マッチ箇所を削除してスラッシュを一つ残す
			// match[0], match[1] がマッチ全体 (/ja/)
			canonical = path[:match[0]] + "/" + path[match[1]:]

			// 連続するスラッシュの整理（/prefix//docs -> /prefix/docs）
			canonical = strings.ReplaceAll(canonical, "//", "/")

			if canonical == "" {
				canonical = "/"
			}
			return locale, canonical
		}
	}

	// ロケールが見つからない場合はパス全体がcanonical
	return "", path
}

// BuildLocaleURL constructs a URL with the specified locale using the configured locale strategy.
//
// It combines a base URL, locale code, and canonical path. The URL is constructed using either
// query parameter mode (appends locale as a query parameter) or path mode (prepends locale to path).
//
// If no locale is provided, returns baseURL + canonical without modification.
func BuildLocaleURL(baseURL, locale, canonical string, cfg *LocaleConfig) string {
	if locale == "" {
		return baseURL + canonical
	}

	// Query parameter mode
	if cfg != nil && cfg.ParamName != "" {
		u, err := url.Parse(baseURL + canonical)
		if err != nil {
			return baseURL + canonical
		}
		q := u.Query()
		q.Set(cfg.ParamName, locale)
		u.RawQuery = q.Encode()
		return u.String()
	}

	// Path mode
	return baseURL + "/" + locale + canonical
}

// ExtractHreflang extracts hreflang alternate links from HTML, returning a map of locales to URLs.
//
// hreflang links indicate alternate versions of a page in different languages/regions.
// This function searches the HTML for <link rel="alternate" hreflang="..."> elements
// and returns a map like {"en": "https://...", "ja": "https://..."}.
//
// Locales are normalized to lowercase. Returns an empty map if no hreflang links are found.
func ExtractHreflang(doc *html.Node) map[string]string {
	result := make(map[string]string)
	if doc == nil {
		return result
	}

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" {
			var rel, hreflang, href string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "rel":
					rel = attr.Val
				case "hreflang":
					hreflang = attr.Val
				case "href":
					href = attr.Val
				}
			}
			if rel == "alternate" && hreflang != "" && href != "" {
				// ロケールを正規化（小文字、ja-JP -> ja-jp）
				normalizedLocale := strings.ToLower(hreflang)
				result[normalizedLocale] = href
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)
	return result
}

// SelectPreferredLocaleURL selects the URL of the preferred locale from a hreflang map.
//
// It uses the provided priority list to select the best match from available locales.
// If an exact match isn't found, it attempts expanded forms (e.g., "ja" -> "ja-jp").
// If no matches in the priority list, returns the first available locale.
//
// Returns both the matched locale code and its URL. Returns empty strings if the map is empty.
func SelectPreferredLocaleURL(hreflangMap map[string]string, priority []string) (locale, url string) {
	if len(hreflangMap) == 0 {
		return "", ""
	}

	// Select according to priority order
	for _, loc := range priority {
		if u, ok := hreflangMap[loc]; ok {
			return loc, u
		}
		// Try expanded forms like ja -> ja-jp
		if u, ok := hreflangMap[loc+"-"+loc]; ok {
			return loc + "-" + loc, u
		}
	}

	// If no priority match, return the first available
	for loc, u := range hreflangMap {
		return loc, u
	}

	return "", ""
}

// NormalizeLocale normalizes locale codes to a canonical form.
//
// It converts locale codes to lowercase and resolves common aliases:
//   - ja-jp -> ja
//   - en-us, en-gb -> en
//   - zh-hans, zh-cn -> zh-cn
//   - zh-hant, zh-tw -> zh-tw
//
// This ensures consistent locale matching across different locale formats.
func NormalizeLocale(locale string) string {
	locale = strings.ToLower(locale)
	// 主要なエイリアス
	switch locale {
	case "ja-jp":
		return "ja"
	case "en-us", "en-gb":
		return "en"
	case "zh-hans", "zh-cn":
		return "zh-cn"
	case "zh-hant", "zh-tw":
		return "zh-tw"
	}
	return locale
}
