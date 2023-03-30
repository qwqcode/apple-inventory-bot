package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
)

var startTime time.Time
var msg string
var number int

type Model struct {
	Model  string  `json:"model"`
	Title  string  `json:"title"`
	Mem    string  `json:"mem"`
	Disk   string  `json:"disk"`
	Price  float64 `json:"price"`
	Status string  `json:"status"`
	Url    string  `json:"url"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	startTime = time.Now()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	http.HandleFunc("/", home)

	log.Println("Server running on " + addr)
	go func() { log.Fatal(http.ListenAndServe(addr, nil)) }()

	for {
		req()
		time.Sleep(1 * time.Minute)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	upTime := time.Since(startTime)
	w.Write([]byte("Running " + upTime.String() + "\n"))
	w.Write([]byte("Last Checked " + strconv.Itoa(number) + "\n"))
	w.Write([]byte("Last Message " + msg + "\n"))
}

func notify(msg string) bool {
	baseURL := os.Getenv("LARK")
	msgJson := []byte(`{"msg_type":"text", "content": {"text":` + strconv.Quote(msg) + `} }`)
	fmt.Println(string(msgJson))
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, _ := http.NewRequest("POST", baseURL, bytes.NewBuffer(msgJson))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Println(string(body))
	return resp.StatusCode == 200
}

func req() {
	log.Println("Start to fetch page...")
	url := strings.TrimSpace(os.Getenv("LIST_URL"))
	if url == "" {
		url = "https://www.apple.com.cn/shop/refurbished/mac/2021-macbook-pro"
	}
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Println(err)
		return
	}
	cookies := os.Getenv("COOKIES")
	if cookies == "" {
		cookies = ""
	}
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"89\", \"Chromium\";v=\"89\", \";Not A Brand\";v=\"99\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,zh-TW;q=0.6")
	req.Header.Add("Cookie", cookies)

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	// body, err := ioutil.ReadAll(res.Body)
	// fmt.Println(string(body))
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	models := []Model{}

	available := false

	log.Println("Page loaded")
	number = 0
	doc.Find("div[role=main]>script").Each(func(i int, s *goquery.Selection) {
		data := strings.ReplaceAll(s.Text(), "window.REFURB_GRID_BOOTSTRAP =", "")
		products := gjson.Get(data, "tiles")
		products.ForEach(func(key, value gjson.Result) bool {
			number = number + 1
			m := Model{
				Model:  value.Get("partNumber").String(),
				Title:  value.Get("title").String(),
				Mem:    value.Get("filters.dimensions.tsMemorySize").String(),
				Disk:   value.Get("filters.dimensions.dimensionCapacity").String(),
				Price:  value.Get("price.currentPrice.raw_amount").Float(),
				Status: value.Get("omnitureModel.customerCommitString").String(),
				Url:    "https://www.apple.com.cn" + value.Get("productDetailsUrl").String(),
			}

			checkMatch := func() bool {
				keywords := os.Getenv("KEYWORDS")
				if keywords != "" {
					for _, k := range strings.Split(keywords, ",") {
						k = strings.TrimSpace(k)
						if k != "" && !strings.Contains(m.Title, k) {
							return false
						}
					}
				}

				mem := os.Getenv("MEM")
				if mem != "" && m.Mem != mem {
					return false
				}

				disk := os.Getenv("DISK")
				if disk != "" && m.Disk != disk {
					return false
				}

				return true
			}

			if checkMatch() {
				available = true

				log.Printf("%.2f %s %s %s\n", m.Price, m.Title, m.Mem, m.Disk)
				models = append(models, m)
			}

			return true // keep iterating
		})

	})

	if available && len(models) > 0 {
		info, err := json.MarshalIndent(models, "", "\t")
		msg = string(info)

		if err != nil {
			log.Println(err)
		}
		if notify(msg) {
			os.Exit(0)
		}
	}
}
