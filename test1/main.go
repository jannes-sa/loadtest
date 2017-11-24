package main

import (
	"loadtest"
)

var (
	nothing map[string]string

	// hostTXN    = `http://203.154.91.206:30004`
	hostTXN    = `http://127.0.0.1:8084`
	prePostURL = hostTXN + `/sav_txn/v1/incoming/validate/OTHBANK`
	postURL    = hostTXN + `/sav_txn/v1/incoming/submit/OTHBANK`

	conc    = 50
	request = 1000
)

func main() {
	txnTest()
}

func txnTest() {
	respPrePost := loadtest.ExecuteTest(prePostURL, `pre_post.json`, nothing,
		conc, request, "csv/prepost")

	// time.Sleep(10 * time.Second)

	xJobID := make(map[string]string)
	for k := range respPrePost {
		xJobID[respPrePost[k].Header.Get("X-job-id")] = `pre-job-id`
	}

	loadtest.ExecuteTest(postURL, `post.json`, xJobID, conc,
		0, "csv/post")
}
