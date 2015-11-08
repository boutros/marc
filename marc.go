package marc

//import "encoding/xml"

type Record struct {
	leader     [24]byte
	ctrlFields []cField
	dataFields []dField
}

type cField struct {
	Tag   string // 3 chars
	Field string // if Tag == "000"; 40 chars
}

type dField struct {
	Tag       string // 3 chars
	Ind1      string // 1 char
	Ind2      string // 1 char
	SubFields []subField
}

type subField struct {
	Code  string // 1 char
	Value string
}
