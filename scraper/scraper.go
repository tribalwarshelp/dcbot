package scraper

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"twdcbot/tribalwars"
	"twdcbot/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

const (
	pathEnnoblementsLive = "/%s/index.php?page=ennoblements&live=live"
)

type Conquest struct {
	Village           string
	VillageID         int
	NewOwnerID        int
	NewOwnerName      string
	NewOwnerTribeID   int
	NewOwnerTribeName string
	OldOwnerID        int
	OldOwnerName      string
	OldOwnerTribeID   int
	OldOwnerTribeName string
	ConqueredAt       time.Time
}

type Conquests []*Conquest

func (c Conquests) LostVillages(tribeID int) Conquests {
	filtered := Conquests{}
	for _, conquer := range c {
		if conquer.OldOwnerTribeID == tribeID && conquer.OldOwnerTribeID != conquer.NewOwnerTribeID {
			filtered = append(filtered, conquer)
		}
	}
	return filtered
}

func (c Conquests) ConqueredVillages(tribeID int) Conquests {
	filtered := Conquests{}
	for _, conquer := range c {
		if conquer.NewOwnerTribeID == tribeID && conquer.NewOwnerTribeID != conquer.OldOwnerTribeID {
			filtered = append(filtered, conquer)
		}
	}
	return filtered
}

type Scraper struct {
	worlds    []string
	since     time.Time
	collector *colly.Collector
	mutex     sync.Mutex
	result    map[string]Conquests
}

func New(worlds []string, since time.Time) *Scraper {
	s := &Scraper{
		since:  since,
		worlds: worlds,
		collector: colly.NewCollector(
			colly.Async(true),
		),
	}
	s.collector.Limit(&colly.LimitRule{
		RandomDelay: time.Second,
		DomainGlob:  "*",
		Parallelism: 5,
	})
	return s
}

func (s *Scraper) getIDFromNodeHref(node *goquery.Selection) int {
	if node != nil {
		nodeHref, ok := node.Attr("href")
		if ok {
			u, err := url.Parse(nodeHref)
			if err == nil {
				if idStr := u.Query().Get("id"); idStr != "" {
					id, err := strconv.Atoi(idStr)
					if err == nil {
						return id
					}
				}
			}
		}
	}

	return 0
}

func (s *Scraper) handleHTML(row *colly.HTMLElement) {
	world := strings.Split(row.Request.URL.Path, "/")[1]
	var err error
	c := &Conquest{}

	conqueredAtString := strings.TrimSpace(row.DOM.Find("td:last-child").Text())
	location := utils.GetLocation(tribalwars.LanguageCodeFromWorldName(world))
	c.ConqueredAt, err = time.ParseInLocation("2006-01-02 - 15:04:05",
		conqueredAtString,
		location)
	if err != nil || c.ConqueredAt.Before(s.since.In(location)) {
		return
	}

	villageAnchor := row.DOM.Find("a:first-child").First()
	c.VillageID = s.getIDFromNodeHref(villageAnchor)
	c.Village = strings.TrimSpace(villageAnchor.Text())

	oldOwnerNode := row.DOM.Find("td:nth-child(3) a:first-child")
	if len(oldOwnerNode.Nodes) == 0 {
		c.OldOwnerName = "-"
		c.OldOwnerTribeName = "-"
	} else {
		c.OldOwnerID = s.getIDFromNodeHref(oldOwnerNode)
		c.OldOwnerName = strings.TrimSpace(oldOwnerNode.Text())
		oldOwnerTribeNode := row.DOM.Find("td:nth-child(3) .tribelink")
		if len(oldOwnerTribeNode.Nodes) != 0 {
			c.OldOwnerTribeName = strings.TrimSpace(oldOwnerTribeNode.Text())
			c.OldOwnerTribeID = s.getIDFromNodeHref(oldOwnerTribeNode)
		} else {
			c.OldOwnerTribeName = "-"
		}
	}

	newOwnerNode := row.DOM.Find("td:nth-child(4) a:first-child")
	c.NewOwnerID = s.getIDFromNodeHref(newOwnerNode)
	c.NewOwnerName = strings.TrimSpace(newOwnerNode.Text())
	newOwnerTribeNode := row.DOM.Find("td:nth-child(4) .tribelink")
	if len(newOwnerTribeNode.Nodes) != 0 {
		c.NewOwnerTribeID = s.getIDFromNodeHref(newOwnerTribeNode)
		c.NewOwnerTribeName = strings.TrimSpace(newOwnerTribeNode.Text())
	} else {
		c.NewOwnerTribeName = "-"
	}

	s.mutex.Lock()
	s.result[world] = append(s.result[world], c)
	s.mutex.Unlock()
}

func (s *Scraper) Scrap() map[string]Conquests {
	s.result = make(map[string]Conquests)
	s.collector.OnHTML(".r1", s.handleHTML)
	s.collector.OnHTML(".r2", s.handleHTML)

	for _, world := range s.worlds {
		url := TwstatsURLs[tribalwars.LanguageCodeFromWorldName(world)]
		if url != "" {
			s.collector.Visit(fmt.Sprintf(url+pathEnnoblementsLive, world))
		}
	}
	s.collector.Wait()
	return s.result
}
