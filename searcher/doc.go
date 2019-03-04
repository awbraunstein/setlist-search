/*
Package searcher is responsible for searching through setlists.

Syntax

The search syntax uses a song as the smallest searchable unit.

Single songs:
 .              any song
 (songname)     a song by the name of songname
 [xyz]          song class, either song1, song2, or song3
 [^xyz]         negated song class, none of song1, song2, or song3

Composites:
 xy             x followed by y
 x|y            x or y (prefer x)

Repetitions:
 x*             zero or more x, prefer more
 x+             one or more x, prefer more
 x?             zero or one x, prefer one

Grouping:
 (?:x)          non-capturing group

Empty Songs:
 ^              at begining of show
 $              at end of show
 \S{SetNum}     at begining of set SetNum (SetNum = e for encore)
 \E{SetNum}     at end of set SetNum (SetNum = e for encore)


Setlists

The searcher will analyze setlists that are stored in the following format and separated by newlines.
Note that songs must not have , or {} characters in them.
 ID{showid}SET1{song1,song2,song3,song4}SET2{songa,songb,songc,songd}ENCORE{songx,songy,songz}


Examples


Matches a show that has Carini in the setlist:
 (Carini)

Matches a show with a Mike's Groove with any song in between.
 (Mike's Song).(Weekapaug Groove)

Matches a show where Mike's song is played in set1 and Weekapaug is played in set 2.
 \S{1}.*(Mike's Song).*\E{1}\S{2}.*(Weekapaug Groove).*\E{2}

Matches a show that had Tweezer Reprise that wasn't played in the encore.
 (Tweezer Reprise)\S{e}[^(Tweezer Reprise)]*\E{e}

*/
package searcher
