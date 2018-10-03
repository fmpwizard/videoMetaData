package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

var filename string

func main() {

	flag.StringVar(&filename, "f", "", "filename or path/to/file to open")

	flag.Parse()
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to open %s, err: %s\n", filename, err)
	}
	walkFile(b)

}

func walkFile(b []byte) {

	// start reading at position 0x20, which is the 8 bytes (2 sets of 4 bytes) offset
	// from http://xhelmboyx.tripod.com/formats/mp4-layout.txt
	// -> 8 bytes wider mdat box offset = 64-bit unsigned offset
	//   - only if mdat standard offset set to 1
	initialOffset := int64(0x20)
	// read 8 bytes
	var start, end int64 = startEnd(initialOffset)
	n := getPortion(b, start, end)
	log.Printf("at position: '%#02X', got: '% 02X'\n", start, n)
	// ret := fmt.Sprintf("%X", n)
	newStart := byteToI(n)
	log.Printf("================== %X\n", newStart)
	s, e := startEnd(newStart + initialOffset)
	a := getPortion(b, s, e) //exploring
	//a := getPortion(b, newStart+0X19, newStart+0X19+12) //exploring
	log.Printf("Rount 2: got: '% 02X'\n", a)
	// get box version to see if we use 32 or 64 bit for info (4 bytes or 8 bytes words)
	v := getPortion(b, s+8, s+8+4)
	wordLength := int64(4)
	if byteToI(v) == 1 {
		wordLength = 8
	}
	// get creation time

	// previous start + 8 bytes of data that represent lmvhd + 4 for version
	// then read only 4 bytes for creation time
	creation := printCreationTime(b, s+8+wordLength, s+8+wordLength+wordLength)

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatalln("failed to set New York as timezone", err)
	}
	log.Println("creation time: ", creation.In(loc))

	//log.Println("==================")
	//getPortion(b, newStart+0X19+12, newStart+0X19+12+4)
	//log.Println("==================")
	//getPortion(b, newStart+0X19+12+4, newStart+0X19+12+4+8)

}

func getPortion(b []byte, start, end int64) []byte {

	//log.Printf("start is %d\n", start)
	//log.Printf("end   is %d\n", end)
	log.Printf("start is % #X\n", start)
	log.Printf("end   is % #X\n", end)
	log.Printf("ascii:   '%s'\n", b[start:end])
	log.Printf("hex:     % 02X\n", b[start:end])
	log.Printf("offset:  % 02x", start)
	return b[start:end]
}

func startEnd(x int64) (int64, int64) {
	return x, x + 8
}

func printCreationTime(b []byte, start, end int64) time.Time {
	log.Printf("===> Time hex:     % 02X\n", b[start:end])
	// midnight,	Jan.	1,	1904,	in	UTC	time
	startingVideoEpoc := time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)
	return startingVideoEpoc.Add(time.Duration(byteToI(b[start:end])) * time.Second)
}

func byteToI(b []byte) int64 {
	newStart, err := strconv.ParseInt(fmt.Sprintf("%X", b), 16, 64)
	if err != nil {
		log.Fatalf("failed to convert %+v to int, err: %s\n", b, err)
	}
	return newStart
}
