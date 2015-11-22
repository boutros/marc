package marc

import (
	"bufio"
	"bytes"
	"testing"
)

var sampleMARC = `01142cam  2200301 a 4500001001300000003000400013005001700017008004100034010001700075020002500092040001800117042000900135050002600144082001600170100003200186245008600218250001200304260005200316300004900368500004000417520022800457650003300685650003300718650002400751650002100775650002300796700002100819   92005291 DLC19930521155141.9920219s1993    caua   j      000 0 eng    a   92005291   a0152038655 :c$15.95  aDLCcDLCdDLC  alcac00aPS3537.A618bA88 199300a811/.522201 aSandburg, Carl,d1878-1967.10aArithmetic /cCarl Sandburg ; illustrated as an anamorphic adventure by Ted Rand.  a1st ed.  aSan Diego :bHarcourt Brace Jovanovich,cc1993.  a1 v. (unpaged) :bill. (some col.) ;c26 cm.  aOne Mylar sheet included in pocket.  aA poem about numbers and their characteristics. Features anamorphic, or distorted, drawings which can be restored to normal by viewing from a particular angle or by viewing the image's reflection in the provided Mylar cone. 0aArithmeticxJuvenile poetry. 0aChildren's poetry, American. 1aArithmeticxPoetry. 1aAmerican poetry. 1aVisual perception.1 aRand, Ted,eill.`

var sampleLineMARC = `*000     c
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

var sampleMARCXML = `
<?xml version="1.0" encoding="UTF-8"?>
<collection xmlns="http://www.loc.gov/MARC21/slim">
  <record>
    <leader>01142cam  2200301 a 4500</leader>
    <controlfield tag="001">   92005291 </controlfield>
    <controlfield tag="003">DLC</controlfield>
    <controlfield tag="005">19930521155141.9</controlfield>
    <controlfield tag="008">920219s1993    caua   j      000 0 eng  </controlfield>
    <datafield tag="010" ind1=" " ind2=" ">
      <subfield code="a">   92005291 </subfield>
    </datafield>
    <datafield tag="020" ind1=" " ind2=" ">
      <subfield code="a">0152038655 :</subfield>
      <subfield code="c">$15.95</subfield>
    </datafield>
    <datafield tag="040" ind1=" " ind2=" ">
      <subfield code="a">DLC</subfield>
      <subfield code="c">DLC</subfield>
      <subfield code="d">DLC</subfield>
    </datafield>
    <datafield tag="042" ind1=" " ind2=" ">
      <subfield code="a">lcac</subfield>
    </datafield>
    <datafield tag="050" ind1="0" ind2="0">
      <subfield code="a">PS3537.A618</subfield>
      <subfield code="b">A88 1993</subfield>
    </datafield>
    <datafield tag="082" ind1="0" ind2="0">
      <subfield code="a">811/.52</subfield>
      <subfield code="2">20</subfield>
    </datafield>
    <datafield tag="100" ind1="1" ind2=" ">
      <subfield code="a">Sandburg, Carl,</subfield>
      <subfield code="d">1878-1967.</subfield>
    </datafield>
    <datafield tag="245" ind1="1" ind2="0">
      <subfield code="a">Arithmetic /</subfield>
      <subfield code="c">Carl Sandburg ; illustrated as an anamorphic adventure by Ted Rand.</subfield>
    </datafield>
    <datafield tag="250" ind1=" " ind2=" ">
      <subfield code="a">1st ed.</subfield>
    </datafield>
    <datafield tag="260" ind1=" " ind2=" ">
      <subfield code="a">San Diego :</subfield>
      <subfield code="b">Harcourt Brace Jovanovich,</subfield>
      <subfield code="c">c1993.</subfield>
    </datafield>
    <datafield tag="300" ind1=" " ind2=" ">
      <subfield code="a">1 v. (unpaged) :</subfield>
      <subfield code="b">ill. (some col.) ;</subfield>
      <subfield code="c">26 cm.</subfield>
    </datafield>
    <datafield tag="500" ind1=" " ind2=" ">
      <subfield code="a">One Mylar sheet included in pocket.</subfield>
    </datafield>
    <datafield tag="520" ind1=" " ind2=" ">
      <subfield code="a">A poem about numbers and their characteristics. Features anamorphic, or distorted, drawings which can be restored to normal by viewing from a particular angle or by viewing the image's reflection in the provided Mylar cone.</subfield>
    </datafield>
    <datafield tag="650" ind1=" " ind2="0">
      <subfield code="a">Arithmetic</subfield>
      <subfield code="x">Juvenile poetry.</subfield>
    </datafield>
    <datafield tag="650" ind1=" " ind2="0">
      <subfield code="a">Children's poetry, American.</subfield>
    </datafield>
    <datafield tag="650" ind1=" " ind2="1">
      <subfield code="a">Arithmetic</subfield>
      <subfield code="x">Poetry.</subfield>
    </datafield>
    <datafield tag="650" ind1=" " ind2="1">
      <subfield code="a">American poetry.</subfield>
    </datafield>
    <datafield tag="650" ind1=" " ind2="1">
      <subfield code="a">Visual perception.</subfield>
    </datafield>
    <datafield tag="700" ind1="1" ind2=" ">
      <subfield code="a">Rand, Ted,</subfield>
      <subfield code="e">ill.</subfield>
    </datafield>
  </record>
</collection>`

func TestDecodeMARC(t *testing.T)     { testDecodeRecord(t, sampleMARC, MARC) }
func TestDecodeLineMARC(t *testing.T) { testDecodeRecord(t, sampleLineMARC, LineMARC) }
func TestDecodeMARCXML(t *testing.T)  { testDecodeRecord(t, sampleMARCXML, MARCXML) }

func testDecodeRecord(t *testing.T, input string, f Format) {
	dec := NewDecoder(bytes.NewBufferString(input), f)

	r, err := dec.DecodeAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(r) != 1 {
		t.Fatalf("expected 1 record; got %d", len(r))
	}
}

func BenchmarkDecodeBaseline(b *testing.B) {
	for n := 0; n < b.N; n++ {
		r := bufio.NewReader(bytes.NewBufferString(sampleLineMARC))
		b.SetBytes(int64(len(sampleLineMARC)))
		_, err := r.ReadBytes(0x5E)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeMARC(b *testing.B)     { benchmarkDecode(b, sampleMARC, MARC) }
func BenchmarkDecodeLineMARC(b *testing.B) { benchmarkDecode(b, sampleLineMARC, LineMARC) }
func BenchmarkDecodeMARCXML(b *testing.B)  { benchmarkDecode(b, sampleMARCXML, MARCXML) }

func benchmarkDecode(b *testing.B, sample string, f Format) {
	for n := 0; n < b.N; n++ {
		b.SetBytes(int64(len(sample)))
		dec := NewDecoder(bytes.NewBufferString(sample), f)
		_, err := dec.Decode()
		if err != nil {
			b.Fatal(err)
		}
	}
}
