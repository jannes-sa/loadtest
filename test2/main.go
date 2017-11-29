package main

import (
	"loadtest"
)

var (
	nothing map[string]string

	// hostTXN    = `http://203.154.91.206:30004`
	hostTXN = `http://127.0.0.1:8084`
	seqURL  = hostTXN + `/sav_txn/v1/poc_sequential`

	conc    = 100
	request = 1000
)

func main() {
	seqTest()
}

func seqTest() (err error) {
	_, err = loadtest.ExecuteTest(seqURL, `seq.json`, nothing,
		conc, request, "csv/seq")
	if err != nil {
		return err
	}

	return
}
