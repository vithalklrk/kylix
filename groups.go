package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type channel struct {
	name   string
	logo   string
	icon   string
	tvType int
	group  string
	title  string
	url    string
	tvg    tvg
}

type group struct {
	title    string
	channels map[string]*channel
}

const (
	channels = iota
	movies
	series
)

type grMap map[string]*group

var tv = [3]grMap{make(grMap), make(grMap), make(grMap)}
var groups grMap = tv[channels]
var ctryIds map[string]string = make(map[string]string)

func saveChannel(ch1 *channel) {
	groupTitle := ch1.group
	if groupTitle != "" {
		if g, ok := groups[groupTitle]; !ok {
			g = &group{groupTitle, make(map[string]*channel)}
			groups[groupTitle] = g
			g.channels[ch1.title] = ch1
		} else {
			if _, ok := g.channels[ch1.title]; ok {
				// log.Println("Duplicate channel name:", ch1.title)
			}
			g.channels[ch1.title] = ch1
		}
	}
}

var reChannel = regexp.MustCompile(`^#EXTINF: ?(?P<duration>-?\d+?) ?(?P<params>.*),(?P<title>.*?)$`)
var reParams = regexp.MustCompile(`(.*?=".*?") ?`)
var reNameVal = regexp.MustCompile(`((?P<name>.*?)="(?P<val>.*?)") ?`)

func processChannel(line string, url string) {
	fields := reChannel.FindStringSubmatch(line)
	if fields == nil {
		// log.Println("Skipping:", line)
		return
	}
	ind := reChannel.SubexpIndex("params")
	ch1 := new(channel)
	params := reParams.FindAllString(fields[ind], -1)
	for _, s := range params {
		nameVals := reNameVal.FindStringSubmatch(s)
		name := nameVals[reNameVal.SubexpIndex("name")]
		val := nameVals[reNameVal.SubexpIndex("val")]
		switch name {
		case "tvg-name":
			ch1.name = val
		case "tvg-logo":
			ch1.logo = val
		case "group-title":
			ch1.group = val
		}
	}
	ind = reChannel.SubexpIndex("title")
	ch1.title = sanitize(fields[ind])
	ch1.url = url
	saveChannel(ch1)
}

func loadFromURL(url string) {
	i := strings.LastIndex(url, "/")
	m3u := url[i+1:]
	fname := filepath.Join(getConfigPath(), m3u)
	var writer *bufio.Writer = nil
	var scanner *bufio.Scanner
	if _, err := os.Stat(fname); err == nil {
		f, _ := os.Open(fname)
		defer f.Close()
		scanner = bufio.NewScanner(f)
	} else {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		scanner = bufio.NewScanner(resp.Body)
		f, _ := os.Create(fname)
		writer = bufio.NewWriter(f)
		defer f.Close()
	}
	var channelLine string = ""
	nLines := 0
	for scanner.Scan() {
		nLines++
		line := scanner.Text()
		if strings.HasPrefix(line, "#EXTM3U") {
			continue
		}
		if strings.HasPrefix(line, "#EXTINF") {
			channelLine = line
			if writer != nil {
				writer.WriteString(line + "\n")
			}
			continue
		}
		if strings.Contains(line, `//`) {
			processChannel(channelLine, line)
			if writer != nil {
				writer.WriteString(line + "\n")
			}
		} else {
			// fmt.Println("skipping:", line)
		}
	}
	for title := range groups {
		g := groups[title]
		if len(g.channels) == 0 {
			delete(groups, title)
			continue
		}
		for _, ch1 := range g.channels {
			if ch1.logo == "" {
				fmt.Println(g.title, ch1.title)
			}
		}
	}
}

type lang struct {
	Code string
	Name string
}
type ctry struct {
	Code string
	Name string
}
type tvg struct {
	Id   string
	Name string
	Url  string
}
type jsChan struct {
	Name  string
	Logo  string
	Url   string
	Catg  string `json:"category"`
	Langs []lang `json:"languages"`
	Ctrys []ctry `json:"countries"`
	Tvg   tvg
}

func loadFromJson(url string) {
	var inp []byte
	i := strings.LastIndex(url, "/")
	jsonFile := url[i+1:]
	fname := filepath.Join(getConfigPath(), jsonFile)
	var stat fs.FileInfo
	var err error
	if stat, err = os.Stat(fname); err != nil {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		f, _ := os.Create(fname)
		io.Copy(f, resp.Body)
		resp.Body.Close()
		f.Close()
		stat, _ = os.Stat(fname)
	}
	inp = make([]byte, stat.Size())
	f, _ := os.Open(fname)
	f.Read(inp)
	f.Close()

	var chs []jsChan
	err = json.Unmarshal([]byte(inp), &chs)

	skip := 0
	for _, jsCh := range chs {
		if jsCh.Name == "" || jsCh.Url == "" || len(jsCh.Ctrys) == 0 {
			skip++
			continue
		}
		for _, ctry := range jsCh.Ctrys {
			ch1 := new(channel)
			nameLC := strings.ToLower(jsCh.Name)
			if strings.Contains(nameLC, "adult") {
				skip++
				continue
			}
			ch1.title = jsCh.Name
			ch1.url = jsCh.Url
			ch1.logo = jsCh.Logo
			ch1.group = ctry.Name
			ch1.tvg = jsCh.Tvg
			if jsCh.Catg == "Movies" {
				groups = tv[movies]
				ch1.tvType = movies
			} else if strings.Contains(nameLC, "series") {
				groups = tv[series]
				ch1.tvType = series
			} else {
				groups = tv[channels]
				ch1.tvType = channels
			}
			saveChannel(ch1)
			ctryIds[ctry.Name] = ctry.Code
		}
	}
	fmt.Println("skipped", skip)
	groups = tv[channels]
	/*	var catgs = make(map[string]bool)
		for _, ch1 := range chs {
			if ch1.Catg != "" {
				catgs[ch1.Catg] = true
			}
		}
		for c := range catgs {
			fmt.Println(c)
		}*/
	/*	for _, ch1 := range chs {
			if strings.Contains(line, `//`) {
				processChannel(channelLine, line)
				if writer != nil {
					writer.WriteString(line + "\n")
				}
			} else {
				// fmt.Println("skipping:", line)
			}
		}
		for title := range groups {
			g := groups[title]
			if len(g.channels) == 0 {
				delete(groups, title)
				continue
			}
			for _, ch1 := range g.channels {
				if ch1.logo == "" {
					fmt.Println(g.title, ch1.title)
				}
			}
		}*/
	return
}
