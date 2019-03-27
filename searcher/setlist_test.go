package searcher

import (
	"reflect"
	"testing"
)

func TestParseSetlist(t *testing.T) {
	tests := []struct {
		name     string
		setlist  string
		expected *Setlist
		err      bool
	}{
		{
			name:    "valid setlist",
			setlist: "ID{1}DATE{2000-07-20}SET1{a,b,c,d}SET2{x,y,z}ENCORE{aa,bb}",
			expected: &Setlist{
				ShowId: "1",
				Date:   "2000-07-20",
				Sets: []*Set{
					&Set{Songs: []string{"a", "b", "c", "d"}},
					&Set{Songs: []string{"x", "y", "z"}},
				},
				Encore: &Set{Songs: []string{"aa", "bb"}},
			},
			err: false,
		},
		{
			name:     "missing ID",
			setlist:  "SET1{a,b,c,d}SET2{x,y,z}ENCORE{aa,bb}",
			expected: nil,
			err:      true,
		},
		{
			name:     "missing DATE",
			setlist:  "ID{1}SET1{a,b,c,d}SET2{x,y,z}ENCORE{aa,bb}",
			expected: nil,
			err:      true,
		},
		{
			name:     "missing sets",
			setlist:  "ID{1}",
			expected: nil,
			err:      true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseSetlist(tc.setlist)
			if tc.err && err == nil {
				t.Fatalf("expected error, but got nil")
			}
			if !tc.err && err != nil {
				t.Fatalf("got err=%v but expected nil", err)
			}
			if err == nil && !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("got: %v\nexpected: %v", got, tc.expected)
			}

		})

	}
}

func TestSetlistString(t *testing.T) {
	setlistString := "ID{1}DATE{2000-04-21}URL{http://google.com}SET1{a,b,c,d}SET2{x,y,z}ENCORE{aa,bb}"
	setlistStruct, err := ParseSetlist(setlistString)
	if err != nil {
		t.Fatalf("Unable to parse setlist; %v", err)
	}
	if got := setlistStruct.String(); got != setlistString {
		t.Errorf("got: %s\nexpected: %s", got, setlistString)
	}
}

func TestParseSetlistFromPhishNet(t *testing.T) {
	setlistData := "<p><span class='set-label'>Set 1</span>: <a href='http://phish.net/song/nicu' class='setlist-song' title='NICU'>NICU</a> > <a href='http://phish.net/song/golgi-apparatus' class='setlist-song' title='Golgi Apparatus'>Golgi Apparatus</a> > <a href='http://phish.net/song/crossroads' class='setlist-song' title='Crossroads'>Crossroads</a>, <a href='http://phish.net/song/cars-trucks-buses' class='setlist-song' title='Cars Trucks Buses'>Cars Trucks Buses</a>, <a href='http://phish.net/song/train-song' class='setlist-song' title='Train Song'>Train Song</a>, <a title=\"Blistering, high octane version with nice concluding transition space and a &gt; to &quot;Fluffhead.&quot;\" href='http://phish.net/song/theme-from-the-bottom' class='setlist-song' title='Blistering, high octane version with nice concluding transition space and a &gt; to &quot;Fluffhead.&quot;'>Theme From the Bottom</a> > <a href='http://phish.net/song/fluffhead' class='setlist-song' title='Fluffhead'>Fluffhead</a>, <a href='http://phish.net/song/dirt' class='setlist-song' title='Dirt'>Dirt</a>, <a title=\"Straightforward but well played jam, followed by some downright filthy funk jamming in the &quot;Rocco&quot; section.\" href='http://phish.net/song/run-like-an-antelope' class='setlist-song' title='Straightforward but well played jam, followed by some downright filthy funk jamming in the &quot;Rocco&quot; section.'>Run Like an Antelope</a></p><p><span class='set-label'>Set 2</span>:<a title=\"Fearsome but exploratory jam. Moments of quiet settle are repeatedly upended by intense funk rocking. This legitimate monster &quot;Disease&quot; finally gives up a little belligerence only to -> into a very strong &quot;David Bowie.&quot;\" href='http://phish.net/song/down-with-disease' class='setlist-song' title='Fearsome but exploratory jam. Moments of quiet settle are repeatedly upended by intense funk rocking. This legitimate monster &quot;Disease&quot; finally gives up a little belligerence only to -> into a very strong &quot;David Bowie.&quot;'>Down with Disease</a><sup title=\"Unfinished.\">[\"2]</sup> -> <a title=\"Excellent and thrilling version with strong musicianship. Mode shift out of typical (but very well played) &quot;Bowie&quot; at 13:55 into a great groove which peaks and returns to &quot;Bowie&quot; by 17:00.\" href='http://phish.net/song/david-bowie' class='setlist-song' title='Excellent and thrilling version with strong musicianship. Mode shift out of typical (but very well played) &quot;Bowie&quot; at 13:55 into a great groove which peaks and returns to &quot;Bowie&quot; by 17:00.'>David Bowie</a><sup title=\"Antelope-esque jamming. James Bond Theme tease from Mike.\">[\"3]</sup> > <a title=\"> in from a strong &quot;Bowie.&quot; There are two &quot;I Can't Turn You Loose&quot; (Blues Brothers) jams in this solid &quot;Possum.&quot;\" href='http://phish.net/song/possum' class='setlist-song' title='> in from a strong &quot;Bowie.&quot; There are two &quot;I Can't Turn You Loose&quot; (Blues Brothers) jams in this solid &quot;Possum.&quot;'>Possum</a>, <a title=\"Simply the slowest, funkiest, and thickest &quot;Tube&quot; ever played, featuring a  seamless full-band groove and breakdown solos by Trey, Page, and Mike.  This jam is a great example of the band playing as one and is among the best versions ever.  &quot;I Feel the Earth Move&quot; tease.\" href='http://phish.net/song/tube' class='setlist-song' title='Simply the slowest, funkiest, and thickest &quot;Tube&quot; ever played, featuring a  seamless full-band groove and breakdown solos by Trey, Page, and Mike.  This jam is a great example of the band playing as one and is among the best versions ever.  &quot;I Feel the Earth Move&quot; tease.'>Tube</a>, <a href='http://phish.net/song/you-enjoy-myself' class='setlist-song' title='You Enjoy Myself'>You Enjoy Myself</a></p><p><span class='set-label'>Encore</span>:<a href='http://phish.net/song/good-times-bad-times' class='setlist-song' title='Good Times Bad Times'>Good Times Bad Times</a><p class='setlist-footer'>[2] Unfinished.<br>[3] Antelope-esque jamming. James Bond Theme tease from Mike.<br></p>"

	want := &Setlist{
		ShowId: "1",
		Date:   "2000-04-20",
		Url:    "http://phish.net/setlists/phish-december-29-1997-madison-square-garden-new-york-ny-usa.html",
		Sets: []*Set{
			&Set{Songs: []string{"nicu", "golgi-apparatus", "crossroads", "cars-trucks-buses", "train-song", "theme-from-the-bottom", "fluffhead", "dirt", "run-like-an-antelope"}},
			&Set{Songs: []string{"down-with-disease", "david-bowie", "possum", "tube", "you-enjoy-myself"}},
		},
		Encore: &Set{Songs: []string{"good-times-bad-times"}},
	}

	wantSongSet := map[string]string{"Cars Trucks Buses": "cars-trucks-buses", "Crossroads": "crossroads", "David Bowie": "david-bowie", "Dirt": "dirt", "Down with Disease": "down-with-disease", "Fluffhead": "fluffhead", "Golgi Apparatus": "golgi-apparatus", "Good Times Bad Times": "good-times-bad-times", "NICU": "nicu", "Possum": "possum", "Run Like an Antelope": "run-like-an-antelope", "Theme From the Bottom": "theme-from-the-bottom", "Train Song": "train-song", "Tube": "tube", "You Enjoy Myself": "you-enjoy-myself"}

	got, gotss, err := ParseSetlistFromPhishNet("1", "2000-04-20", "http://phish.net/setlists/phish-december-29-1997-madison-square-garden-new-york-ny-usa.html", setlistData)
	if err != nil {
		t.Fatalf("Unable to parse setlist; %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v\nexpected: %v", got, want)
	}
	if !reflect.DeepEqual(gotss, wantSongSet) {
		t.Errorf("got: %#v\nexpected: %v", gotss, wantSongSet)
	}

}
