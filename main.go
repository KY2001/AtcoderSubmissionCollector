package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Submission struct {
	ID            int     `json:"id"`
	EpochSecond   int     `json:"epoch_second"`
	ProblemID     string  `json:"problem_id"`
	ContestID     string  `json:"contest_id"`
	UserID        string  `json:"user_id"`
	Language      string  `json:"language"`
	Point         float64 `json:"point"`
	Length        int     `json:"length"`
	Result        string  `json:"result"`
	ExecutionTime int     `json:"execution_time"`
}

func main() {
	// Input username
	var userName string
	fmt.Println("-- What is your AtCoder username? --")
	fmt.Print("res: ")
	_, err := fmt.Scanln(&userName)
	if err != nil {
		log.Println("failed:", err)
		return
	}

	epochSecond := 0 // A single informationRequest can retrieve data on 500 submissions from this time.
	var dataAll []Submission
	fmt.Println("-- Please wait. I am checking the number of your submissions --")
	for i := 0; ; i++ {
		url := fmt.Sprintf("https://kenkoooo.com/atcoder/atcoder-api/v3/user/submissions?user=%s&from_second=%d", userName, epochSecond)
		data, err := informationRequest(url)
		if err != nil {
			log.Println("failed:", err)
			return
		}
		dataAll = append(dataAll, data...)

		if i == 0 && len(data) == 0 {
			log.Println("failed: " + "The submission does not exist at all.")
			return
		} else if len(data) == 0 {
			break
		}

		epochSecond = data[len(data)-1].EpochSecond + 1
		time.Sleep(time.Second)
	}

	fmt.Printf("-- It looks like you have %d submissions to AtCoder -- \n", len(dataAll))
	fmt.Println("-- Next, we will download each submission code --")
	fmt.Println("-- Enter the name of the directory(path) where the files will be saved --")
	fmt.Print("res: ")
	var dirName string
	_, err = fmt.Scanln(&dirName)
	if err != nil {
		log.Println("failed:", err)
		return
	}

	if err := os.MkdirAll(dirName, 0777); err != nil {
		log.Println("failed: ", err)
		return
	}

	fmt.Printf("-- Please wait. This is expected to take approximately %v seconds. --\n", len(dataAll))

	for i, sub := range dataAll {
		url := fmt.Sprintf("https://atcoder.jp/contests/%v/submissions/%v", sub.ContestID, sub.ID)
		code, err := submissionRequest(url)
		if err != nil {
			log.Println("failed:", err)
			return
		}

		if err := os.MkdirAll(fmt.Sprintf("./%v/%v", dirName, sub.ContestID), 0777); err != nil {
			log.Println("failed: ", err)
			return
		}

		var extension string
		if strings.Contains(sub.Language, "C++") {
			extension = "cpp"
		} else if strings.Contains(sub.Language, "C") {
			extension = "c"
		} else if strings.Contains(sub.Language, "Python") || strings.Contains(sub.Language, "PyPy") {
			extension = "py"
		} else if strings.Contains(sub.Language, "Rust") {
			extension = "rs"
		} else {
			extension = "txt"
		}
		fileName := fmt.Sprintf("%v_%v", sub.ProblemID, sub.EpochSecond)

		file, err := os.Create(fmt.Sprintf("./%v/%v/%v.%v", dirName, sub.ContestID, fileName, extension))
		if err != nil {
			fmt.Println(err)
		}
		file.WriteString(code)
		file.Close()

		fmt.Printf("%v/%v %v\n", i+1, len(dataAll), fmt.Sprintf("%v/%v.%v", sub.ContestID, fileName, extension))
		time.Sleep(time.Second)
	}

}

func informationRequest(url string) ([]Submission, error) {
	var ret []Submission

	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ret, err
	}

	res, err := client.Do(req)
	if err != nil {
		return ret, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ret, err
	}

	if err := json.Unmarshal(body, &ret); err != nil {
		log.Println("failed:", err)
		return ret, err
	}

	return ret, nil
}

func submissionRequest(url string) (string, error) {
	var ret string

	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ret, err
	}

	res, err := client.Do(req)
	if err != nil {
		return ret, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ret, err
	}
	prefix := "<pre id=\\\"submission-code\\\" class=\\\"prettyprint linenums\\\">"
	postfix := "</pre>"
	r := regexp.MustCompile(fmt.Sprintf(`(?s)%s(.*)%s`, prefix, postfix))
	ret = r.FindString(string(body))

	ret = html.UnescapeString(ret[len(prefix)-4 : len(ret)-len(postfix)])

	return ret, nil
}
