package rutracker

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

type Client struct {
	http   *http.Client
	parser *Parser
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
