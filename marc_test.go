package marc

import (
	"bytes"
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
			false,
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

	dec := NewDecoder(bytes.NewBufferString(lmarc1), LineMARC)
	r, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}

	want := `000      c
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
