package index

import (
	"io/ioutil"
	"os"
	"testing"
)

// TODO: replace this with a test that doesn't use a full index.
func TestOpenIndex(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "searcher-read-test")
	if err != nil {
		t.Fatal(err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	indexStr := `setsearcher index 1
[SONGS]
Chalk Dust Torture|chalk-dust-torture
[END]
[SETLISTS]
ID{1249948108}DATE{2000-09-17}SET1{guyute,back-on-the-train,bathtub-gin,limb-by-limb,the-moma-dance,lawn-boy,fluffhead,the-curtain-with,chalk-dust-torture}SET2{rock-and-roll,theme-from-the-bottom,dog-log,the-mango-song,free}ENCORE{contact,rocky-top}
ID{1249948445}DATE{1985-03-04}SET1{anarchy,camel-walk,fire-up-the-ganja,skippy-the-wondermouse,in-the-midnight-hour}
ID{1250019273}DATE{1998-07-02}SET1{birds-of-a-feather,cars-trucks-buses,theme-from-the-bottom,brian-and-robert,meat,fikus,shafty,fluffhead,ginseng-sullivan,punch-you-in-the-eye,character-zero}SET2{ghost,runaway-jim,prince-caspian,you-enjoy-myself}ENCORE{simple}
ID{1250024745}DATE{1998-07-10}SET1{down-with-disease,dogs-stole-things,divided-sky,mikes-song}SET2{halleys-comet,roggae,sparkle,mikes-song,simple,weekapaug-groove,sample-in-a-jar,good-times-bad-times}ENCORE{brian-and-robert,taste}
ID{1250387629}DATE{1990-01-20}SET1{carolina,the-sloth,bathtub-gin,you-enjoy-myself,the-squirming-coil,caravan,the-lizards,run-like-an-antelope}SET2{the-oh-kee-pa-ceremony,suzy-greenberg,bouncing-around-the-room,reba,tela,la-grange,lawn-boy,esther,mikes-song,i-am-hydrogen,weekapaug-groove}ENCORE{harry-hood}
ID{1250454896}DATE{1994-04-04}SET1{divided-sky,sample-in-a-jar,scent-of-a-mule,maze,fee,reba,horn,its-ice,possum}SET2{down-with-disease,if-i-could,buried-alive,the-landlady,julius,magilla,split-open-and-melt,wolfmans-brother,i-wanna-be-like-you,the-oh-kee-pa-ceremony,suzy-greenberg}ENCORE{harry-hood,cavern}
ID{1250458591}DATE{1994-04-05}SET1{runaway-jim,foam,fluffhead,glide,julius,bouncing-around-the-room,rift,acdc-bag}SET2{peaches-en-regalia,ya-mar,tweezer,if-i-could,you-enjoy-myself,i-wanna-be-like-you,hold-your-head-up,chalk-dust-torture,amazing-grace}ENCORE{nellie-kane,golgi-apparatus}
ID{1250458932}DATE{1994-04-06}SET1{llama,guelah-papyrus,poor-heart,stash,the-lizards,sample-in-a-jar,scent-of-a-mule,fee,run-like-an-antelope}SET2{the-curtain,down-with-disease,wolfmans-brother,sparkle,mikes-song,lifeboy,weekapaug-groove,the-squirming-coil,cavern}ENCORE{ginseng-sullivan,nellie-kane,sweet-adeline}
[END]`

	if _, err := tmpFile.WriteString(indexStr); err != nil {
		t.Fatalf("unable to write tempfile; %v", err)
	}

	tmpFile.Close()

	i, err := Open(tmpName)
	if err != nil {
		t.Fatalf("unable to read index; %v", err)
	}

	if len(i.setlists) != 8 {
		t.Errorf("Wrong number of shows. Expected 8, but got %d", len(i.setlists))
	}

	if len(i.reverseIndex["wolfmans-brother"]) != 2 {
		t.Errorf("Expected two shows with the song wolfmans-brother")
	}

	writeFile, err := ioutil.TempFile("", "searcher-read-test")
	if err != nil {
		t.Fatal(err)
	}
	writeFileName := writeFile.Name()
	writeFile.Close()
	defer os.Remove(writeFileName)
	if err := i.Write(writeFileName); err != nil {
		t.Fatalf("unable to write index; %v", err)
	}
	wf, err := os.Open(writeFileName)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ioutil.ReadAll(wf)
	if err != nil {
		t.Fatalf("unable to read file; %v", err)
	}
	if string(got) != indexStr {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", indexStr, got)
	}
}
