package loadtest

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jannes-sa/easycsv"
)

type (
	// Resp ...
	Resp struct {
		Status     int
		URL        string
		XRoundTrip int
		Body       []byte
		Header     http.Header
	}

	// CSVGenerateFile ...
	CSVGenerateFile struct {
		URL        string  `csv:"url"`
		Concurrent int     `csv:"concurrent"`
		Success    int     `csv:"success"`
		Failed     int     `csv:"failed"`
		Average    float64 `csv:"average"`
		Min        int     `csv:"min"`
		Max        int     `csv:"max"`
		TimeUsed   string  `csv:"time_used"`
		TPS        float64 `csv:"tps"`
	}
)

var (
	timeNow time.Time

	tps    int
	allTPS []int
)

// ExecuteTest ...
func ExecuteTest(URL string, file string, parsingReqBody map[string]string,
	conc int, nReq int, csvFileName string) (respDatas []Resp, err error) {

	if nReq%conc > 0 {
		err = errors.New("Value Must Matched Between Each Other")
		log.Println(err)
		return
	}

	timeNow = time.Now()
	go getTPS()

	testStart := time.Now()
	jsonStr := readFile(file)
	if len(parsingReqBody) == 0 {
		ch := make(chan Resp)
		th := 0

		for x := 0; x < nReq; x++ {
			th++
			if th >= conc {
				for i := 0; i < conc; i++ {
					go post(URL, jsonStr, ch)
				}

				for i := 0; i < conc; i++ {
					respDatas = append(respDatas, <-ch)
				}
				th = 0
			}
		}
	} else {
		var newRqBody []string
		for k, v := range parsingReqBody {
			newRqBody = append(
				newRqBody,
				strings.Replace(string(jsonStr), "{{"+v+"}}", k, -1),
			)
		}

		var (
			ix1      = 0
			ix2      = 0
			th       = 0
			loopConc = 1
		)

		ch := make(chan Resp)
		for i := 0; i < len(newRqBody); i++ {
			th++
			if th >= conc {
				ix2 = conc * loopConc
				// log.Println(ix1, ix2)
				// log.Println(newRqBody[ix1:ix2])

				for y := 0; y < len(newRqBody[ix1:ix2]); y++ {
					go post(URL, []byte(newRqBody[ix1:ix2][y]), ch)
				}

				for y := 0; y < len(newRqBody[ix1:ix2]); y++ {
					respDatas = append(respDatas, <-ch)
				}

				ix1 = ix2
				loopConc++
				th = 0
			}
		}
	}

	testEnd := time.Since(testStart).String()
	avgTPS := calculateAvgTPS()

	generateCSV(respDatas, URL, conc, testEnd, csvFileName, avgTPS)

	return
}

func calculateAvgTPS() (avgTPS float64) {
	var totTPS int
	for k := range allTPS {
		totTPS += allTPS[k]
	}

	avgTPS = float64(totTPS) / float64(len(allTPS))
	return
}

func getTPS() {
	for {
		sinceSecond := time.Since(timeNow).Seconds()
		if sinceSecond >= 1 {
			allTPS = append(allTPS, tps)
			tps = 0
			timeNow = time.Now()
		}
	}
}

func generateCSV(respDatas []Resp, URL string, conc int,
	elapsed string, csvFileName string, avgTPS float64) {

	var (
		csvsData       []CSVGenerateFile
		success        = 0
		failed         = 0
		totalRoundTrip = 0
		avg            = float64(0)
		min            = respDatas[0].XRoundTrip
		max            = respDatas[0].XRoundTrip
	)
	for _, v := range respDatas {
		if v.Status == 200 {
			success++
			totalRoundTrip += v.XRoundTrip
			if min > v.XRoundTrip {
				min = v.XRoundTrip
			}
			if max < v.XRoundTrip {
				max = v.XRoundTrip
			}
		} else {
			failed++
		}
	}
	avg = float64(totalRoundTrip) / float64(success)

	csvsData = append(csvsData, CSVGenerateFile{
		URL:        URL,
		Concurrent: conc,
		Success:    success,
		Failed:     failed,
		Average:    avg,
		Min:        min,
		Max:        max,
		TimeUsed:   elapsed,
		TPS:        avgTPS,
	})

	err := easycsv.WriteCSVData(csvsData, "./"+csvFileName+".csv")
	checkErr(err)
}

func readFile(file string) []byte {
	rawData, err := ioutil.ReadFile(file)
	checkErr(err)

	return rawData
}

func rqHeader() map[string]string {
	arrHeader := make(map[string]string)
	arrHeader["x-request-id"] = "cuman diterusin"
	arrHeader["x-job-id"] = ""
	arrHeader["x-real-ip"] = "192.168.99.100"
	arrHeader["x-caller-service"] = "CUS"
	arrHeader["x-caller-domain"] = "CUS DOMAIN"
	arrHeader["x-device"] = "ANDROID DDD"
	arrHeader["x-channel"] = "MBL"
	arrHeader["user-agent"] = "POSTMAN"
	arrHeader["datetime"] = "2017-09-19T10:59:47.305411285+07:00"
	arrHeader["accept"] = "application/json"
	arrHeader["accept-language"] = "en/id"
	arrHeader["accept-encoding"] = "UTF-8"
	arrHeader["Content-Type"] = "application/json"

	return arrHeader
}

func post(URL string, jsonStr []byte, c chan Resp) {
	req, err := http.NewRequest(
		"POST",
		URL,
		bytes.NewBuffer(jsonStr),
	)
	rqHeader := rqHeader()
	for k := range rqHeader {
		req.Header.Set(k, rqHeader[k])
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	tps++

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var respData Resp
	respData.Status = resp.StatusCode

	respData.Header = resp.Header

	roundTrip, _ := strconv.Atoi(resp.Header.Get("X-Roundtrip"))
	respData.XRoundTrip = roundTrip
	respData.URL = URL

	body, _ := ioutil.ReadAll(resp.Body)
	respData.Body = body

	c <- respData

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header.Get("X-Roundtrip"))
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

// Code //
