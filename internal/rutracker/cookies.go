package rutracker

import (
	"net/http"
	"net/url"
	"sync"
)

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
	cookiesMap := map[string]*http.Cookie{}
	for _, c := range j.jar[u.Hostname()] {
		cookiesMap[c.Name] = c
	}
	for _, c := range cookies {
		cookiesMap[c.Name] = c
	}
	cookies = make([]*http.Cookie, 0, len(cookiesMap))
	for _, c := range cookiesMap {
		cookies = append(cookies, c)
	}
	j.jar[u.Hostname()] = cookies
	j.mu.Unlock()
}
