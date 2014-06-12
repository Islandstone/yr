package go-yr

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"
)

type TabularData struct {
	XMLName   xml.Name `xml:"weatherdata>forecast>tabular"`
	intervals []Interval
}

type Interval struct {
	XMLName xml.Name `xml:"weatherdata>forecast>tabular>time"`
	from    string   `xml:from,attr`
	to      string   `xml:to,attr`
}

func TestYr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var buf []byte
		var err error

		if buf, err = ioutil.ReadFile("varsel.xml"); err != nil {
			t.Error(err)
		}

		w.Write(buf)
	}))
	defer ts.Close()

	data, err := LoadFromURL(ts.URL)

	assert.Equal(t, err, nil)

	CheckResponseFormat(t, data)
}

func TestCache(t *testing.T) {
	hitCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var buf []byte
		var err error

		if buf, err = ioutil.ReadFile("varsel.xml"); err != nil {
			t.Error(err)
		}

		w.Write(buf)
		hitCount += 1
	}))
	defer ts.Close()

	data, err := LoadFromURL(ts.URL)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, data, nil)
	assert.Equal(t, hitCount, 1)

	data, err = LoadFromURL(ts.URL)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, data, nil)
	assert.Equal(t, hitCount, 1)

	InvalidateCache()

	data, err = LoadFromURL(ts.URL)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, data, nil)
	assert.Equal(t, hitCount, 2)
}

// Check that the response from the server is as expected
func CheckResponseFormat(t *testing.T, data *WeatherData) {
	assert.Equal(t, "Gol", data.Name)

	assert.NotEqual(t, data.Time[0], nil)
	assert.Equal(t, 20, data.Time[0].Temperature.Value)
	assert.Equal(t, "2014-06-07T16:00:00", data.Time[0].From)
	assert.Equal(t, "2014-06-07T18:00:00", data.Time[0].To)
	assert.Equal(t, 22, data.Time[0].Symbol.Number)
	assert.Equal(t, "Regn og torden", data.Time[0].Symbol.Name)

	// Check that there's a minimum of three time fields
	assert.NotEqual(t, data.Time[1], nil)
	assert.NotEqual(t, data.Time[2], nil)
}

// More a test of sanity than a test of the implementation
func TestTimeComparison(t *testing.T) {
	t0 := time.Now()
	t1 := time.Now().Add(10 * time.Minute)

	assert.True(t, t0.Before(t1))
	assert.True(t, t1.After(t0))
}
