package marc

import (
	"bufio"
	"bytes"
	"testing"
)

var lmarc1 string = `*000     c
*0010010463
*008871001                a          0 nob r
*015  $a29$bBibliofilID
*019  $bl
*0823 $a242
*090  $c242$dKar
*100 0$aKarlén, Barbro$d1954-$jsv.$310008600
*24010$aI begynnelsen skapade Gud
*24510$aI begynnelsen skapte Gud$bdikt og salmer i norsk gjendiktning   ved Inger Hagerup ; prosaen overs. av Gunnel Malmström ; illus. av Vanja Hübinette-Simonson
*260  $aOslo$bAschehoug$c1968
*300  $a97 s.$bill.
*574  $aOriginaltittel: I begynnelsen skapade Gud
*655  $aAndaktsbøker$310008700
*70010$aHagerup, Inger$d1905-1985$jn.$310008800
*850  $aDEICHM$sn
^`

func TestDecodeLineMARC(t *testing.T) {
	dec := NewDecoder(bytes.NewBufferString(lmarc1), LineMARC)

	r, err := dec.DecodeAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(r) != 1 {
		t.Logf("%v", r)
		t.Fatalf("expected 1 record; got %d", len(r))
	}
}

func BenchmarkBaseline(b *testing.B) {
	for n := 0; n < b.N; n++ {
		r := bufio.NewReader(bytes.NewBufferString(lmarc1))
		b.SetBytes(int64(len(lmarc1)))
		_, err := r.ReadBytes(Terminator)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeLineMARC(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.SetBytes(int64(len(lmarc1)))
		dec := NewDecoder(bytes.NewBufferString(lmarc1), LineMARC)
		_, err := dec.Decode()
		if err != nil {
			b.Fatal(err)
		}
	}
}
