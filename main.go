package main

import (
	"os"
	"net/http"
	"fmt"
	"strconv"
	"log"
	"encoding/json"
	"bytes"
	"time"
	"io/ioutil"
	"strings"
	"sort"
	"gopkg.in/alecthomas/kingpin.v2"
)

type IndexItem struct {
	Image string        `json:"image"`
	Score float64       `json:"score"`
}

type Index struct {
	Terms map[string][]IndexItem `json:"terms"`
}

var (
	app = kingpin.New("image-search", "A command-line application for searching for images.")

	search = app.Command("search", "Search for images by topic.")
	searchTopic = search.Arg("topic", "The topic to search for.").Required().String()

	build = app.Command("build", "Create initial search index.")
	buildApikey = build.Flag("apikey", "The Clarifai API key to use when making requests").Short('k').Required().String()
	buildPath = build.Flag("url", "The URL of a list of images to index.").Short('u').String()

)

func main() {

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case search.FullCommand():
		topic := strings.ToLower(*searchTopic)
		fmt.Println("Search for " + topic)

		index, err := LoadIndex()
		if err != nil  {
			log.Fatalln("No local index found. Run with 'build' command first.")
		}

		if terms, ok := index.Terms[topic]; ok {
			fmt.Println("Found " + strconv.Itoa(len(terms)) + " matches for " + topic)
			for _, item := range terms {
				fmt.Println(item.Image)
			}
		} else {
			fmt.Println("No images found matching that topic.")
		}

	case build.FullCommand():
		fmt.Println("Building index from scratch...")
		index := rebuildIndex(*buildPath)
		SaveIndex(index)
		fmt.Println("...done.")

	}

	//BuildIndex("https://samples.clarifai.com/metro-north.jpg", &index)
	//log.Println(index)
	//index = LoadIndex()

}


func rebuildIndex(url string) Index {

	list_url := "https://s3.amazonaws.com/clarifai-data/backend/api-take-home/images.txt"

	if len(url) > 0 {
		list_url = url
	}

	fmt.Println("Building index from " + list_url)

	urls := getList(list_url)
	fmt.Println(strconv.Itoa(len(urls)) + " urls found.")
	//urls = urls[0:10]

	//log.Println(len(urls))

	index := Index{
		Terms: make(map[string][]IndexItem),
	}

	for i, url := range urls {
		BuildIndex(url, &index)

		if i % 20 == 0 {
			fmt.Println(strconv.Itoa(i) + string(" images indexed."))
		}
	}

	index.sort()
	SaveIndex(index)

	return index
}


func getList(url string) []string {
	resp, err := http.Get(url)
	if (err != nil) {
		log.Fatalln("Could not retrieve image urls")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if (err != nil) {
		log.Fatalln("Could not retrieve image urls")
	}

	//log.Println(string(body))

	urls := strings.Split(string(body), "\n")

	return urls
}


func BuildIndex(url string, index *Index) {

	matches := makeRequest(url)

	for _, match := range matches {
		//log.Println(match)
		item := IndexItem {
			Image: url,
			Score: match.Score,
		}
		if _, ok := index.Terms[match.Key]; !ok {
			index.Terms[match.Key] = make([]IndexItem, 0, 1)
		}
		index.Terms[match.Key] = append(index.Terms[match.Key], item)
	}

}



func SaveIndex(index Index) {
	bytes, err := json.Marshal(index)
	if err != nil {
		log.Fatalln(err)
	}

	ioutil.WriteFile("./index.json", bytes, 0644)
	if err != nil {
		log.Fatalln(err)
	}

}


func LoadIndex() (Index, error) {
	var index Index

	bytes, err := ioutil.ReadFile("./index.json")
	if err != nil {
		return Index{}, err
	}

	if err := json.Unmarshal(bytes, &index); err != nil {
		return Index{}, err
	}

	return index, nil
}


func (index *Index) sort() {
	for term, items := range index.Terms {
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].Score > items[j].Score
		})

		index.Terms[term] = items
	}
}


type Match struct {
	URL string
	Key string
	Score float64
}


func makeRequests(urls []string, out chan Match) {
	for _, url := range urls {
		matches := makeRequest(url)
		for _, match := range matches {
			out <- match
		}
	}
	close(out)
}



func makeRequest(url string) []Match {

	API_ENDPOINT := "https://api.clarifai.com/v2/models/aaa03c23b3724a16a56b629203edc62c/outputs"
	API_KEY := *buildApikey

	message := map[string]interface{}{
		"inputs": []map[string]interface{}{	{
			"data": map[string]interface{}{
				"image": map[string]string{
					"url": url}}}}}

	// {
    //   "inputs": [
    //     {
    //       "data": {
    //         "image": {
    //           "url": "https://samples.clarifai.com/metro-north.jpg"
    //         }
    //       }
    //     }
    //   ]
    // }

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Println("Bytes")
		log.Fatalln(err)
	}

	//log.Println(string(bytesRepresentation))

	req, err := http.NewRequest("POST", API_ENDPOINT,
		bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Println("Req:")
		log.Fatalln(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Key " + API_KEY)

	timeout := time.Duration(50 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Do:")
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	//log.Println(resp.Status)
	//bytes2, err := json.Marshal(result)
	//log.Println(string(bytes2))
	//log.Println(result)

	matches := make([]Match, 0)


	outputs := result["outputs"].([]interface{})
	outputs_0 := outputs[0].(map[string]interface{})
	data := outputs_0["data"].(map[string]interface{})
	concepts, ok := data["concepts"].([]interface{})
	if (!ok) {
		log.Println("Found no concepts for " + url)
		return matches
	}

	for _, concept := range concepts {
		concept2 := concept.(map[string]interface{})
		score := concept2["value"].(float64)
		key := concept2["name"].(string)
		m := Match { URL: url, Key: key, Score: score }
		matches = append(matches, m)
	}

	return matches
}
