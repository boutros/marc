package marc

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestRecordEq(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{
			`<record> </record>`,
			`<record></record>`,
			true,
		},
		{
			`<record><leader>01142cam  2200301 a 4500</leader></record>`,
			`<record><leader>01142cam  2200301 a 4500</leader> </record>`,
			true,
		},
		{
			`<record><leader>    c                   </leader></record>`,
			`<record><leader>    d                   </leader></record>`,
			true,
		},
		{
			`<record>
				<controlfield tag="001">0010463</controlfield>
			</record>`,
			`<record>
				<controlfield tag="001">0010463</controlfield>
			</record>`,
			true,
		},
		{
			`<record>
				<controlfield tag="001">0010463</controlfield>
			</record>`,
			`<record>
				<controlfield tag="001">0010463</controlfield>
			</record>`,
			true,
		},
		{
			`<record>
				<controlfield tag="008">871001                a          0 nob r</controlfield>
				<controlfield tag="001">0010463</controlfield>
			</record>`,
			`<record>
				<controlfield tag="001">0010463</controlfield>
				<controlfield tag="008">871001                a          0 nob r</controlfield>
			</record>`,
			true,
		},
		{
			`<record>
				<controlfield tag="001">0010463</controlfield>
				<controlfield tag="008">871001                a          0 nob r</controlfield>
			</record>`,
			`<record>
					<controlfield tag="001">0010463</controlfield>
			</record>`,
			false,
		},
		{
			`<record>
				<datafield tag="082" ind1="3" ind2=" ">
					<subfield code="a">242</subfield>
				</datafield>
			</record>`,
			`<record>
				<datafield tag="082" ind1="3" ind2=" ">
					<subfield code="a">242</subfield>
				</datafield>
			</record>`,
			true,
		},
		{
			`<record>
				<datafield tag="082" ind1="3" ind2=" ">
					<subfield code="a">242</subfield>
				</datafield>
			</record>`,
			`<record>
				<datafield tag="082" ind2=" " ind1="3">
					<subfield code="a">242</subfield>
				</datafield>
			</record>`,
			true,
		},
		{
			`<record>
				<datafield tag="082" ind1="3" ind2=" ">
					<subfield code="a">242</subfield>
				</datafield>
			</record>`,
			`<record>
				<datafield tag="082" ind1="3" ind2=" ">
					<subfield code="a">243</subfield>
				</datafield>
			</record>`,
			false,
		},
		{
			`<record>
				<datafield tag="655" ind1=" " ind2=" ">
					<subfield code="a">Andaktsbøker</subfield>
					<subfield code="3">10008700</subfield>
				</datafield>
			</record>`,
			`<record>
				<datafield tag="655" ind1=" " ind2=" ">
					<subfield code="3">10008700</subfield>
					<subfield code="a">Andaktsbøker</subfield>
				</datafield>
			</record>`,
			true,
		},
		{
			`<record>
				<datafield tag="655" ind1=" " ind2=" ">
					<subfield code="a">Andaktsbøker</subfield>
					<subfield code="3">10008700</subfield>
				</datafield>
			</record>`,
			`<record>
				<datafield tag="655" ind1=" " ind2=" ">
					<subfield code="a">Andaktsbøker</subfield>
				</datafield>
			</record>`,
			false,
		},
		{
			`<record>
				<datafield tag="260" ind1=" " ind2=" ">
					<subfield code="a">Oslo</subfield>
					<subfield code="c">1968</subfield>
					<subfield code="b">Aschehoug</subfield>
				</datafield>
				<datafield tag="300" ind1=" " ind2=" ">
					<subfield code="a">97 s.</subfield>
					<subfield code="b">ill.</subfield>
				</datafield>
			</record>`,
			`<record>
				<datafield tag="300" ind1=" " ind2=" ">
					<subfield code="b">ill.</subfield>
					<subfield code="a">97 s.</subfield>
				</datafield>
				<datafield tag="260" ind1=" " ind2=" ">
					<subfield code="b">Aschehoug</subfield>
					<subfield code="a">Oslo</subfield>
					<subfield code="c">1968</subfield>
				</datafield>
			</record>`,
			true,
		},
	}

	for _, test := range tests {
		dec := NewDecoder(bytes.NewBufferString(test.a), MARCXML)
		a, err := dec.Decode()
		if err != nil {
			t.Fatal(err)
		}
		dec = NewDecoder(bytes.NewBufferString(test.b), MARCXML)
		b, err := dec.Decode()
		if err != nil {
			t.Fatal(err)
		}

		if eq := a.Eq(b); eq != test.want {
			t.Errorf("\n%v\n%v\nequal => %v; want %v", a, b, eq, test.want)
		}
	}
}

func TestDump(t *testing.T) {
	var b bytes.Buffer

	dec := NewDecoder(bytes.NewBufferString(sampleLineMARC), LineMARC)
	r, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}

	want := `     c   a22        4500
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

func TestManipulateRecord(t *testing.T) {
	input := `*000     n
*0011830213
*008160205                j          10mul 2
*015  $a0451876$bBIBBI
*015  $a10417351$bBibliofilID
*019  $abu,u,mu$bl$dR$etd
*020  $a978-82-999778-1-4$bh.$cNkr 300.00
*020  $a978-82-999778-1-5$bh.$cNkr 3000.00
*041  $anobsme
*090  $au$bu$cSAM$dGut
*100 0$aGuttorm, Anna Anita$32050849700
*24510$aSkuvlla-Biehtára Ánne ja Bohkosahtti Dáža Sirkus$cAnna Anita Guttorm & Elin Margrethe Wersland
*24600$aSkole-Petter Anna og det lattervekkende norske sirkus
*260  $a[Kårvik]$bE.M. Wersland$c2015
*300  $a128 s.$bill.
*520  $aSenter for halsbrekkende samiske leker har gjort samene søkkrike. Skole-Petter Anna og Rosa-Rosa har hørt at telemarkingene er i ferd med å miste dialekta. De bestemmer seg for å la nordmenn opptre som sirkusartister i Karasjok mot penger, som en slags hjelp til selvhjelp. De kontakter Senter for søringer, der Bård Tufteluft og Harald Eira jobber. Her møter man to kvænske søstre, Tenn-Sticki og Dyna-Mitti Paukk-Aua, samt ulike Oslo-samer, Snåsa-samer, samer fra Manndalen og til slutt Karasjok. Det lattervekkende norske sirkus har et artistgalleri med flere norske sirkusartister, f.eks. Halvard Kvit-Jord som er spesialist på fryktelige fistelstev. I tillegg har Senter for halsbrekkende samiske leker fått en ny gren, det vil si NM i lavvokasting. Dette er en humoristisk og samfunnskritisk historie. Tekst på nordsamisk og bokmål. Fortelling for mellom-/ungdomstrinnet.
*546  $aParallelltekst på bokmål og nordsamisk
*590  $a16-uke5j
*599  $a16x0205gb
*692 2$aKarasjok$xFortellinger$32050849800
*693 2$aHumoristisk$328982800
*699 2$aHumoristisk$2BS
*699 2$aKarasjok$1948.4621$2BS
*699 2$aSirkus$1791.3$2BS
*700 0$aWersland, Elin Margrethe$jn.$emedforf.$32039139900
*850  $aDEICHM$sn
^`

	want := `*008                                 1
*020  $a978-82-999778-1-4
*020  $a978-82-999778-1-5
^`

	dec := NewDecoder(bytes.NewBufferString(input), LineMARC)
	inputRec, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}

	dec = NewDecoder(bytes.NewBufferString(want), LineMARC)
	wantRec, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}

	got := NewRecord()
	f008, ok := inputRec.GetCField("008")
	if !ok {
		t.Fatal("missing control field 008")
	}
	got.SetCField(CField{
		Tag:   "008",
		Value: fmt.Sprintf("%34s", f008.Value[33:34]),
	})
	f20 := inputRec.GetDFields("020")
	for _, f := range f20 {
		got.AddDField(
			NewDField("020").AddSubField("a", f.SubField("a")),
		)
	}

	wantRec.Leader = ""
	if !wantRec.Eq(got) {
		t.Errorf("got:\n%v\nwant:\n%v", got, wantRec)
	}

}
