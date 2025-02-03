package manganato

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/belphemur/mangal/provider/generic"
	"github.com/belphemur/mangal/util"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var timeAgoRegex = regexp.MustCompile(`(?m)((?P<hours>\d{1,2}) hour ago)|((?P<mins>\d{1,2}) mins ago)`)

var Config = &generic.Configuration{
	Name:            "Manganato",
	Delay:           50 * time.Millisecond,
	Parallelism:     50,
	ReverseChapters: true,
	BaseURL:         "https://manganato.com",
	GenerateSearchURL: func(baseUrl string, query string) string {
		query = strings.ReplaceAll(query, " ", "_")
		query = strings.TrimSpace(query)
		query = strings.ToLower(query)
		query = url.QueryEscape(query)
		template := "https://manganato.com/search/story/%s"
		return fmt.Sprintf(template, query)
	},
	MangaExtractor: &generic.MangaExtractor{
		Selector: "div.search-story-item",
		Name: func(selection *goquery.Selection) string {
			return strings.TrimSpace(selection.Find("a.item-title").Text())
		},
		URL: func(selection *goquery.Selection) string {
			return selection.Find("a.item-title").AttrOr("href", "")
		},
		Cover: func(selection *goquery.Selection) string {
			return selection.Find("img").AttrOr("src", "")
		},
	},
	ChapterExtractor: &generic.ChapterExtractor{
		Selector: "li.a-h",
		Name: func(selection *goquery.Selection) string {
			name := selection.Find("a").Text()
			if strings.HasPrefix(name, "Vol.") {
				splitted := strings.Split(name, " ")
				name = strings.Join(splitted[1:], " ")
			}
			return name
		},
		URL: func(selection *goquery.Selection) string {
			return selection.Find("a").AttrOr("href", "")
		},
		Volume: func(selection *goquery.Selection) string {
			name := selection.Find("a").Text()
			if strings.HasPrefix(name, "Vol.") {
				splitted := strings.Split(name, " ")
				return splitted[0]
			}
			return ""
		},
		Date: func(selection *goquery.Selection) *time.Time {
			layout := "Jan 02,06"
			publishedDate := strings.TrimSpace(selection.Find(".chapter-time").Text())
			date, err := time.Parse(layout, publishedDate)
			if err != nil {
				dateMatch := util.ReGroups(timeAgoRegex, publishedDate)
				if len(dateMatch) == 0 {
					return nil
				}
				var timeUnit time.Duration
				var timeToParse string

				if dateMatch["hours"] != "" {
					timeUnit = time.Hour
					timeToParse = dateMatch["hours"]
				} else if dateMatch["mins"] != "" {
					timeUnit = time.Minute
					timeToParse = dateMatch["mins"]
				} else {
					return nil
				}

				timeAgo, err := strconv.ParseInt(timeToParse, 10, 8)
				if err != nil {
					return nil
				}
				date = time.Now().Add(time.Duration(-timeAgo) * timeUnit)
			}
			return &date
		},
	},
	PageExtractor: &generic.PageExtractor{
		Selector: ".container-chapter-reader img",
		URL: func(selection *goquery.Selection) string {
			return selection.AttrOr("src", "")
		},
	},
}
