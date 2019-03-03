package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html/charset"
)

type Client struct {
	http   *http.Client
	parser *Parser
}

type cookieJar struct {
	mu  sync.RWMutex
	jar map[string][]*http.Cookie
}

func newCookieJar() *cookieJar {
	return &cookieJar{
		jar: map[string][]*http.Cookie{},
	}
}

func (j cookieJar) Cookies(u *url.URL) []*http.Cookie {
	j.mu.RLock()
	cookies := j.jar[u.Hostname()]
	j.mu.RUnlock()
	return cookies
}

func (j cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.mu.Lock()
	j.jar[u.Hostname()] = cookies
	j.mu.Unlock()
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: newCookieJar(),
		},
		parser: &Parser{},
	}
}

func (c *Client) Login(user, password string) error {
	body := strings.NewReader(
		"redirect=search.php&login_username=" + user + "&login_password=" + password + "&login=%C2%F5%EE%E4",
	)
	req, err := http.NewRequest("POST", "http://rutracker.org/forum/login.php", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "http://rutracker.org/forum/login.php?redirect=search.php")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

type TorrentFile struct {
	ID         int
	Name       string
	Seeds      int
	Size       int
	Date       time.Time
	Tags       []string
	ForumTopic string
}

func (c *Client) List(q string) ([]TorrentFile, error) {
	form := url.Values{
		"f[]": {"-1"},
		"o":   {"10"},
		"s":   {"2"},
		"pn":  {""},
		"nm":  {q},
	}

	resp, err := c.http.PostForm("https://rutracker.org/forum/tracker.php?nm="+q, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r, err := newUTF8ResponseReader(resp)
	if err != nil {
		return nil, err
	}
	return c.parser.ParseList(r)
}

func newUTF8ResponseReader(resp *http.Response) (io.Reader, error) {
	return charset.NewReader(
		resp.Body,
		resp.Header.Get("Content-Type"),
	)
}

type Links struct {
	Download string
	Magnet   string
}

func (c *Client) GetLinks(f TorrentFile) (links Links, err error) {
	resp, err := c.http.Get("https://rutracker.org/forum/viewtopic.php?t=" + strconv.Itoa(f.ID))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	r, err := newUTF8ResponseReader(resp)
	if err != nil {
		return
	}

	magnet, err := c.parser.ParseMagnetLink(r)
	if err != nil {
		return
	}

	links.Magnet = magnet + "&dn=" + url.QueryEscape(f.Name)
	links.Download = "https://rutracker.org/forum/dl.php?t=" + strconv.Itoa(f.ID)

	return
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	c := NewClient()

	err := c.Login("", "")
	checkError(err)

	files, err := c.List("napoleon newborn")
	checkError(err)
	log.Printf("%#v", files[0])

	links, err := c.GetLinks(files[0])
	checkError(err)
	log.Print(links)
}
