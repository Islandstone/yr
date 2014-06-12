package yr

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var cache map[string]CacheEntry = nil

type WeatherData struct {
	Name   string `xml:"location>name"`
	Time   []Time `xml:"forecast>tabular>time"`
	Links  []Link `xml:"links>link"`
	Credit Credit `xml:"credit>link"`
}

type Credit struct {
	Text string `xml:"text,attr"`
	URL  string `xml:"url,attr"`
}

type Link struct {
	Id  string `xml:"id,attr"`
	URL string `xml:"url,attr"`
}

type TempData struct {
	// XMLName xml.Name `xml:"temperature"`
	// Unit string `xml:"unit,attr"`
	Value int `xml:"value,attr"`
}

type Tabular struct {
	Inner string `xml:,innerxml`
	// Intervals []Time `xml:"time"`
}

type Symbol struct {
	Number    int    `xml:"numberEx,attr"`
	Variation string `xml:"var,attr"`
	Name      string `xml:"name,attr"`
}

type Time struct {
	// Inner string `xml:",innerxml"`
	From string `xml:"from,attr"`
	To   string `xml:"to,attr"`
	// Temp string `xml:>temperature>value,attr`
	Symbol      Symbol   `xml:"symbol"`
	Temperature TempData `xml:"temperature"`
}

func (d WeatherData) String() string {
	return fmt.Sprintf("Current weather for %s: %s, %d degrees C",
		d.Name,
		d.Current().Symbol.Name,
		d.Current().Temperature.Value)
}

func (w WeatherData) Current() Time {
	return w.Time[0]
}

func (t Time) String() string {
	return fmt.Sprintf("%s to %s: %d deg.", t.From, t.To, t.Temperature.Value)
}

func LoadFromFile(filename string) (data *WeatherData, err error) {
	var buf []byte

	if buf, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	data = &WeatherData{}
	if err = xml.Unmarshal(buf, data); err != nil {
		data = nil
		return
	}

	return
}

type CacheEntry struct {
	expires time.Time
	data    WeatherData
}

func (d WeatherData) GetCredits() Credit {
	return d.Credit
}

func InvalidateCache() {
	for k, _ := range cache {
		delete(cache, k)
	}
}

func LoadFromURL(URL string) (weatherData *WeatherData, err error) {
	if cache == nil {
		cache = make(map[string]CacheEntry)
	}

	var data []byte = nil

	if entry, ok := cache[URL]; ok {
		if entry.expires.After(time.Now()) {
			return &entry.data, nil
		}
	}

	resp, err := http.Get(URL)

	if err != nil {
		return nil, err
	}

	defer func() {
		err = resp.Body.Close() // errcheck
	}()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	wData := WeatherData{}
	if err = xml.Unmarshal(data, &wData); err != nil {
		weatherData = nil
	}

	cache[URL] = CacheEntry{time.Now().Add(10 * time.Minute), wData}

	return &wData, nil
}
