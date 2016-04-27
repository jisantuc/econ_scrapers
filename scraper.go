package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type UrlDirector struct {
	Url     string
	Journal string
}

type Record struct {
	Url      string
	Journal  string
	Abstract string
	JelCodes []string
	Citation string
}

func UrlDirectorFromSel(s *goquery.Selection, journal string) UrlDirector {
	link, _ := s.Find("a").Attr("href")
	ud := UrlDirector{link, journal}
	return ud
}

func write_urls(outf string, uds []UrlDirector) {
	f, ferr := os.Create(outf)
	defer f.Close()
	if ferr != nil {
		log.Fatal(ferr)
	}

	jsonified, jsonerr := json.Marshal(uds)
	if jsonerr != nil {
		log.Fatal(jsonerr)
	}

	f.Write(jsonified)
}

func getCitation(sel *goquery.Selection) string {
	strSlice := make([]string, 2)
	strSlice[0] = sel.Find("h1").Text()             // title
	strSlice[1] = sel.Find("p:nth-child(5)").Text() // publication data
	return strings.Join(strSlice, ". ")
}

func ProcAERUrl(url string) Record {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(url)
		log.Fatal(err)
	}

	var rec Record

	jelFromHref := func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if strings.Index("jel", href) != -1 {
			length := len([]rune(href))
			code := href[length-3:]
			rec.JelCodes = append(rec.JelCodes, code)
		}
	}

	bodyText := doc.Find("div.bodytext")
	abstractRaw := bodyText.Find("p:nth-child(6)").Text()
	// apparently some AER papers **don't have abstracts**?? what the hell
	if strings.Index(abstractRaw, "Abstract") == -1 {
		return rec
	}
	abstractRawLength := len([]rune(abstractRaw))
	var firstJel int
	jelInd := strings.Index(abstractRaw, " (JEL")
	if jelInd >= 0 {
		firstJel = jelInd + len([]rune(" (JEL "))
		rec.Abstract = abstractRaw[:jelInd]
		codes := abstractRaw[firstJel : abstractRawLength-1]
		if strings.Index(", ", codes) >= 0 {
			codesSl := strings.Split(codes, ", ")
			for _, code := range codesSl {
				rec.JelCodes = append(rec.JelCodes, code)
			}
		} else {
			rec.JelCodes = append(rec.JelCodes, codes)
		}
	} else {
		rec.Abstract = abstractRaw
		codeBlock := bodyText.Find("p:nth-child(7) a")
		codeBlock.Each(jelFromHref)
	}

	rec.Citation = getCitation(bodyText)
	rec.Url = url
	rec.Journal = "AER"

	return rec
}

func ProcQJEUrl(url string) Record {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	var rec Record

	bodyText := doc.Find("div.bodytext")
	abstractRaw := bodyText.Find("p:nth-child(6)").Text()
	abstractRawLength := len([]rune(abstractRaw))
	jelInd := strings.Index(abstractRaw, "JEL Codes")
	var firstJel int
	if jelInd >= 0 {
		firstJel = jelInd + len([]rune("JEL Codes: "))
	} else if strings.Index(abstractRaw, "JEL Code") >= 0 {
		jelInd = strings.Index(abstractRaw, "JEL Code")
		firstJel = jelInd + len([]rune("JEL Code: "))
	} else if strings.Index(abstractRaw, "JELCodes") >= 0 {
		jelInd = strings.Index(abstractRaw, "JELCodes")
		firstJel = jelInd + len([]rune("JELCodes: "))
	}

	// if-else logic because Split inelegantly handles cases
	// of missing sep string
	if strings.Index(abstractRaw, ", ") >= 0 {
		codes := strings.Split(
			abstractRaw[firstJel:abstractRawLength-1], ", ")
		for _, code := range codes {
			rec.JelCodes = append(rec.JelCodes, code)
		}
	} else {
		rec.JelCodes = append(rec.JelCodes,
			abstractRaw[firstJel:abstractRawLength-1])
	}
	rec.Abstract = abstractRaw[:jelInd]
	rec.Citation = getCitation(bodyText)
	rec.Url = url
	rec.Journal = "QJE"

	return rec
}

func ProcJPEUrl(url string) Record {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	var rec Record

	bodyText := doc.Find("div.bodytext")
	abstractRaw := bodyText.Find("p:nth-child(6)").Text()
	rec.Abstract = abstractRaw
	rec.Citation = getCitation(bodyText)
	rec.Url = url
	rec.Journal = "JPE"

	return rec
}

func ProcEMAUrl(url string) Record {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	var rec Record

	bodyText := doc.Find("div.bodytext")
	abstractRaw := bodyText.Find("p:nth-child(6)").Text()
	rec.Abstract = abstractRaw
	rec.Citation = getCitation(bodyText)
	rec.Url = url
	rec.Journal = "EMA"

	return rec
}

func ProcRESUrl(url string) Record {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	var rec Record

	bodyText := doc.Find("div.bodytext")
	abstractRaw := bodyText.Find("p:nth-child(6)").Text()
	rec.Abstract = abstractRaw
	rec.Citation = getCitation(bodyText)
	rec.Url = url
	rec.Journal = "RES"

	// special handling necessary because JEL codes are hidden behind the
	// link to the text of the paper
	paperLink := bodyText.Find("p:nth-child(8) a").Text()
	linkDoc, err := goquery.NewDocument(paperLink)
	if err != nil {
		log.Fatal(err)
	}

	elements := linkDoc.Find("ul.jel li span a")

	f := func(i int, s *goquery.Selection) {
		rec.JelCodes = append(rec.JelCodes, s.Text())
	}

	elements = elements.Each(f)

	return rec

}

func getFunc(journal string) func(string) Record {
	if journal == "AER" {
		return ProcAERUrl
	} else if journal == "EMA" {
		return ProcEMAUrl
	} else if journal == "JPE" {
		return ProcJPEUrl
	} else if journal == "QJE" {
		return ProcQJEUrl
	} else {
		return ProcRESUrl
	}
}

func ProcURL(url string, journal string) Record {
	f := getFunc(journal)
	return f(url)
}

func scrape_links(url, journal string) []UrlDirector {
	doc, err := goquery.NewDocument(url)

	if err != nil {
		log.Fatal(err)
	}

	outSlice := make([]UrlDirector, 0)

	selString := "div.bodytext dl dt"
	dts := doc.Find(selString)

	dts.Each(func(i int, s *goquery.Selection) {
		outSlice = append(outSlice, UrlDirectorFromSel(s, journal))
	})

	return outSlice
}

func write_out(f *os.File, rec Record) {

	js, jserr := json.Marshal(rec)
	if jserr != nil {
		log.Fatal(jserr)
	}
	f.Write(js)
	f.Write([]byte("\n"))

}

func ScrapeQJE(f *os.File, wg *sync.WaitGroup) {
	journal := "QJE"
	url := "http://econpapers.repec.org/article/oupqjecon/"
	links := scrape_links(url, journal)
	var fullUrl string
	for _, lnk := range links {
		fullUrl = url + lnk.Url
		rec := ProcURL(fullUrl, lnk.Journal)
		write_out(f, rec)
	}
	wg.Done()
}

func ScrapeJPE(f *os.File, wg *sync.WaitGroup) {
	journal := "JPE"
	url := "http://econpapers.repec.org/article/ucpjpolec/"
	links := scrape_links(url, journal)
	var fullUrl string
	for _, lnk := range links {
		fullUrl = url + lnk.Url
		rec := ProcURL(fullUrl, lnk.Journal)
		write_out(f, rec)
	}
	wg.Done()
}

func ScrapeEMA(f *os.File, wg *sync.WaitGroup) {
	journal := "EMA"
	url := "http://econpapers.repec.org/article/wlyemetrp/"
	links := scrape_links(url, journal)
	var fullUrl string
	for _, lnk := range links {
		fullUrl = url + lnk.Url
		rec := ProcURL(fullUrl, lnk.Journal)
		write_out(f, rec)
	}
	wg.Done()
}

func ScrapeRES(f *os.File, wg *sync.WaitGroup) {
	journal := "RES"
	url := "http://econpapers.repec.org/article/ouprestud/"
	links := scrape_links(url, journal)
	var fullUrl string
	for _, lnk := range links {
		fullUrl = url + lnk.Url
		rec := ProcURL(fullUrl, lnk.Journal)
		write_out(f, rec)
	}
	wg.Done()
}

func ScrapeAER(f *os.File, wg *sync.WaitGroup) {
	journal := "AER"
	url := "http://econpapers.repec.org/article/aeaaecrev/"
	links := scrape_links(url, journal)
	var fullUrl string
	for _, lnk := range links {
		fullUrl = url + lnk.Url
		rec := ProcURL(fullUrl, lnk.Journal)
		write_out(f, rec)
	}
	wg.Done()
}

func ScrapeAll(f *os.File) {
	fmt.Println("Starting scrapes")
	var wg sync.WaitGroup
	// n goroutines to wait for, should be same as n lines before
	// call to wg.Wait
	wg.Add(5)
	go ScrapeQJE(f, &wg)
	go ScrapeJPE(f, &wg)
	go ScrapeEMA(f, &wg)
	go ScrapeRES(f, &wg)
	go ScrapeAER(f, &wg)
	wg.Wait()
}

func main() {
	out, ferr := os.OpenFile("records.json",
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer out.Close()
	if ferr != nil {
		log.Fatal(ferr)
	}
	ScrapeAll(out)
}
