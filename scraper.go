package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

//TODO: write scraper functions for each journal page
//JPE: http://econpapers.repec.org/article/ucpjpolec/doi_3a10.1086_2f684718.htm
//EMA: http://econpapers.repec.org/article/wlyemetrp/v_3a83_3ay_3a2015_3ai_3a5_3ap_3a1685-1725.htm
//AER: http://econpapers.repec.org/article/aeaaecrev/v_3a106_3ay_3a2016_3ai_3a4_3ap_3a855-902.htm
//ReStud1: http://econpapers.repec.org/article/ouprestud/v_3a82_3ay_3a2015_3ai_3a3_3ap_3a825-867..htm
//ReStud2: http://restud.oxfordjournals.org/content/82/3/825

func getCitation(sel *goquery.Selection) string {
	strSlice := make([]string, 2)
	strSlice[0] = sel.Find("h1").Text()             // title
	strSlice[1] = sel.Find("p:nth-child(5)").Text() // publication data
	return strings.Join(strSlice, ". ")
}

func ProcAERUrl(url string) Record {
	return Record{}
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
	return Record{}
}

func ProcEMAUrl(url string) Record {
	return Record{}
}

func ProcRESUrl(url string) Record {
	return Record{}
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

func ScrapeQJE(ch chan Record) {
	journal := "QJE"
	url := "http://econpapers.repec.org/article/oupqjecon/"
	links := scrape_links(url, journal)
	var fullUrl string
	for _, lnk := range links {
		fullUrl = url + lnk.Url
		ch <- ProcURL(fullUrl, lnk.Journal)
	}
	close(ch)
}

// TODO write scrape functions for other journals

func main() {
	ch := make(chan Record)
	go ScrapeQJE(ch)
	for rec := range ch {
		fmt.Println(rec.Journal)
	}
}
