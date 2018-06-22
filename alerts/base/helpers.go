package base

/*
func MakeELKRequest(address string, method string, body []byte, ll int) (int, []byte, error) {
	//assume that we have the normal auth
	req, err := http.NewRequest(method, address, bytes.NewReader(body))
	if err != nil {
		log.Printf("[Heartbeat-lost] There was a problem forming the request: %v", address)
		return 0, []byte{}, err
	}

	req.SetBasicAuth(os.Getenv("ELK_SA_USERNAME"), os.Getenv("ELK_SA_PASSWORD"))
	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("\tThere was a problem making the request: %v", err.Error())
		return 0, []byte{}, err
	}

	//get the body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("\tCould not read the response body: %v", err.Error())
		return 0, []byte{}, err
	}

	if resp.StatusCode/100 != 2 {
		log.Printf("\tnon 200 response code sent. Code: %v, body: %s ", resp.StatusCode, b)
	}
	return resp.StatusCode, b, nil
}
*/
