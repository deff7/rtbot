package rutracker

import "time"

type TorrentFile struct {
	ID         int
	Name       string
	Seeds      int
	Size       int
	Date       time.Time
	Tags       []string
	ForumTopic string
}

type Links struct {
	Download string
	Magnet   string
}
