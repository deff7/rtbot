package main

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestParseList(t *testing.T) {
	f, err := os.Open("./test/html/list.html")
	if err != nil {
		t.Fatal(err)
	}

	p := &Parser{}
	got, err := p.ParseList(f)
	if err != nil {
		t.Fatal(err)
	}

	want := []TorrentFile{
		{
			ID:         5233744,
			Name:       "Napoleon - Newborn Mind - 2016, MP3, 320 kbps",
			Seeds:      6,
			Date:       time.Unix(1464537290, 0),
			Size:       90945651,
			Tags:       []string{"Melodic Metalcore", "Hardcore"},
			ForumTopic: "Hardcore (lossy)",
		},
		{
			ID:         4145208,
			Name:       "Napoleon - What We See [EP] (2012) , MP3, 320 kbps",
			Seeds:      2,
			Date:       time.Unix(1344159192, 0),
			Size:       59357392,
			Tags:       []string{"Melodic Metalcore", "Hardcore"},
			ForumTopic: "Hardcore (lossy)",
		},
	}

	if len(want) != len(got) {
		t.Fatalf("expect\n%v\nto equal\n%v", got, want)
	}

	for i := range want {
		if !reflect.DeepEqual(want[i], got[i]) {
			t.Errorf("expect\n%v\nto equal\n%v", got[i], want[i])
		}
	}
}

func TestParseMagnetLink(t *testing.T) {
	f, err := os.Open("./test/html/details.html")
	if err != nil {
		t.Fatal(err)
	}

	p := &Parser{}

	got, err := p.ParseMagnetLink(f)
	if err != nil {
		t.Fatal(errors.Wrap(err, "details"))
	}

	want := `magnet:?xt=urn:btih:B478519A944D7C618EDC4D8A51E89FD1CEE24CFA&tr=http%3A%2F%2Fbt4.t-ru.org%2Fann%3Fmagnet&dn=(Melodic%20Metalcore%20%2F%20Hardcore)%20Napoleon%20-%20Newborn%20Mind%20-%202016%2C%20MP3%2C%20320%20kbps`

	if want != got {
		t.Errorf("expect %q to equal %q", got, want)
	}
}
