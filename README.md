# marc

This package provides encoders and decoders for MARC bibliographic records. It can handle standard binary MARC (MARC21 ISO2709), MARCXML (MarcXchange ISO25577) and Line-MARC (not a standard, but commonly used in Norway in the semi-standard NORMARC).

## Usage 

The package implements streaming decoding and encoding of MARC records, enabeling you to parse huge datasets with minimum memory footprint. Simply create a decoder over a ìo.Reader`, and call `Decode()` until the end of stream:

```
marcFile, err := os.Open("mydb.mrc")
if err != nil {
	log.Fatal(err)
}

dec := marc.NewDecoder(marcFile, marc.LineMARC)
for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
	doSomethingWith(rec)
}
```

See the [marc2marc](cmd/marc2marc) utility for a more complete example.

## Command line utilities

The repo includes 3 utilities which can be seen as example of how to use the package, or maybe usefull in their own right:

* [marccheck](cmd/marccheck) - Parse MARC database to check for errors.
* [marcdump](cmd/marcdump) - Pretty print MARC database to terminal.
* [marc2marc](cmd/marc2marc) - Convert between different MARC serializations.

## Performance

I have not looked into optimizing this yet, but it should be reasonably performant (with the exception of XML parsing, which is inherently slower). Here are the numbers from my laptop: (The baseline benchmarks measures reading and writing arbritary bytes):

```
BenchmarkDecodeBaseline-4	 1000000	      2109 ns/op	 273.00 MB/s	    5360 B/op	       4 allocs/op
BenchmarkDecodeMARC-4    	   30000	     44486 ns/op	  25.67 MB/s	   17712 B/op	     245 allocs/op
BenchmarkDecodeLineMARC-4	   50000	     27613 ns/op	  20.86 MB/s	   11056 B/op	     137 allocs/op
BenchmarkDecodeMARCXML-4 	    3000	    743082 ns/op	   4.63 MB/s	   53232 B/op	    1104 allocs/op
BenchmarkEncodeBaseline-4	 1000000	      2964 ns/op	 385.24 MB/s	    2394 B/op	       0 allocs/op
BenchmarkEncodeMARC-4    	   30000	     66836 ns/op	  17.09 MB/s	    9763 B/op	     106 allocs/op
BenchmarkEncodeLineMARC-4	  200000	      9912 ns/op	  58.11 MB/s	    4192 B/op	       3 allocs/op
BenchmarkEncodeMARCXML-4 	   10000	    138788 ns/op	  24.81 MB/s	   25776 B/op	     179 allocs/op
```