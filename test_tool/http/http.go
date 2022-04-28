package http2

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Get (url string) []byte {
	response, error := http.Get(url)
	if error != nil {
		fmt.Println(error)
		return nil
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	return body
}