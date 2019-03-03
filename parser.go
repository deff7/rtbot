package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

type Parser struct{}

func (p *Parser) ParseList(r io.Reader) ([]TorrentFile, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var (
		files = []TorrentFile{}
	)

	doc.Find("tr.tCenter.hl-tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		var file TorrentFile

		// Parse ID
		titleLink := s.Find(".t-title > a")
		rawID, ok := titleLink.Attr("data-topic_id")
		if !ok {
			err = fmt.Errorf("can't fetch ID at row %d", i)
			return false
		}
		file.ID, err = strconv.Atoi(rawID)
		if err != nil {
			err = errors.Wrap(err, "parse ID")
			return false
		}
		// Parse Name
		file.Name = strings.TrimSpace(titleLink.Text())

		// Parse Seeds
		rawSeeds := s.Find("b.seedmed").Text()
		if rawSeeds != "" {
			file.Seeds, err = strconv.Atoi(rawSeeds)
			if err != nil {
				err = errors.Wrap(err, "parse seeds")
				return false
			}
		}

		// Parse Date
		rawDate := s.Find("td > u").Last().Text()
		ts, err := strconv.Atoi(rawDate)
		if err != nil {
			err = errors.Wrap(err, "parse date")
			return false
		}
		file.Date = time.Unix(int64(ts), 0)

		// Parse Size
		rawSize := s.Find("td.tor-size > u").Text()
		file.Size, err = strconv.Atoi(rawSize)
		if err != nil {
			err = errors.Wrap(err, "parse size")
			return false
		}

		// Parse Tags
		file.Tags = s.Find("span.tg").Map(func(_ int, s *goquery.Selection) string {
			return strings.TrimSpace(s.Text())
		})

		// Parse Forum topic
		file.ForumTopic = strings.TrimSpace(s.Find(".f-name > a").Text())

		files = append(files, file)
		return true
	})

	return files, err
}

func (p *Parser) ParseMagnetLink(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", err
	}

	link, ok := doc.Find("a.magnet-link").Attr("href")
	log.Println(link)
	if !ok {
		return "", errors.New("magnet link not found")
	}

	return link, nil
}
