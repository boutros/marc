package marc

import (
	"bytes"
	"strings"
	"testing"
)

func TestDump(t *testing.T) {
	var b bytes.Buffer

	dec := NewDecoder(bytes.NewBufferString(lmarc1), LineMARC)
	r, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}

	want := `leader
000      c
001 0010463
008 871001                a          0 nob r
015 __ |a 29 |b BibliofilID
019 __ |b l
082 3_ |a 242
090 __ |c 242 |d Kar
100 _0 |a Karlén, Barbro |d 1954- |j sv. |3 10008600
240 10 |a I begynnelsen skapade Gud
245 10 |a I begynnelsen skapte Gud |b dikt og salmer i norsk gjendiktning ved
        Inger Hagerup ; prosaen overs. av Gunnel Malmström ; illus. av Vanja
        Hübinette-Simonson
260 __ |a Oslo |b Aschehoug |c 1968
300 __ |a 97 s. |b ill.
574 __ |a Originaltittel: I begynnelsen skapade Gud
655 __ |a Andaktsbøker |3 10008700
700 10 |a Hagerup, Inger |d 1905-1985 |j n. |3 10008800
850 __ |a DEICHM |s n

`
	r.DumpTo(&b, false)
	if strings.Replace(b.String(), " ", "", -1) != strings.Replace(want, " ", "", -1) {
		t.Errorf("Record.DumpTo =>\n%q\nwant:\n%q", b.String(), want)
	}
}
