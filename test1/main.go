package main

import (
	"encoding/json"
	"loadtest"
	"log"
	"time"
)

var (
	nothing map[string]string

	hostAccount = `http://127.0.0.1:8082`
	// hostAccount      = `http://203.154.91.206:30002`
	createAccountURL = hostAccount + `/sav_account/v1/accounts`

	// hostTXN = `http://203.154.91.206:30004`
	hostTXN    = `http://127.0.0.1:8084`
	prePostURL = hostTXN + `/sav_txn/v1/incoming/validate/OTHBANK`
	postURL    = hostTXN + `/sav_txn/v1/incoming/submit/OTHBANK`

	conc    = 100
	request = 1000
)

func main() {
	txnTest()
}

func txnTest() (err error) {
	respAccount, err := loadtest.ExecuteTest(createAccountURL,
		`create_account.json`, nothing, conc, request, "csv/create_account")

	accountNumberData := make(map[string]string)
	for k := range respAccount {
		var t map[string]interface{}
		json.Unmarshal(respAccount[k].Body, &t)
		acc := t["rsBody"].(map[string]interface{})["account_number"].(string)
		accountNumberData[acc] = `beneficiary_account_number`
	}
	log.Println(accountNumberData)

	time.Sleep(10 * time.Second)

	respPrePost, err := loadtest.ExecuteTest(prePostURL, `pre_post.json`, accountNumberData,
		conc, request, "csv/prepost")
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	xJobID := make(map[string]string)
	for k := range respPrePost {
		xJobID[respPrePost[k].Header.Get("X-job-id")] = `pre-job-id`
	}

	loadtest.ExecuteTest(postURL, `post.json`, xJobID, conc,
		0, "csv/post")

	return
}
