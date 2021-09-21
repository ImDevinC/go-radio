package jsrlive

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

const urlBase = "https://jetsetradio.live"

func FormatSong(station string, song string) string {
	return urlBase + "/radio/stations/" + url.PathEscape(station) + "/" + url.PathEscape(song) + ".mp3"
}

func GetStatic() string {
	return urlBase + "/radio/stations/static.mp3"
}

func GetStations() []string {
	return []string{
		"outerspace",
		"ultraremixes",
		"summer",
		"halloween",
		"christmas",
		"snowfi",
		"classic",
		"future",
		"garage",
		"ggs",
		"noisetanks",
		"poisonjam",
		"rapid99",
		"loveshockers",
		"immortals",
		"doomriders",
		"goldenrhinos",
		"ganjah",
		"lofi",
		"chiptunes",
		"retroremix",
		"classical",
		"revolutoin",
		"endofdays",
		"crazytaxi",
		"ollieking",
		"toejamandearl",
		"hover",
		"butterflies",
		"bonafidebloom",
		"verafx",
		"djchidow",
	}
}

func GetBumps() ([]string, error) {
	bumpUrl := urlBase + "/radio/stations/bumps/~list.js"
	resp, err := http.Get(bumpUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	songs := parseBumps(string(body))
	return songs, nil
}

func GetSongs(station string) ([]string, error) {
	listUrl := urlBase + "/radio/stations/" + station + "/~list.js"
	resp, err := http.Get(listUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	songs := parseSongs(string(body))
	return songs, nil
}

func parseSongs(body string) []string {
	r := regexp.MustCompile(`this\[stationName\+'_tracks'\]\[this\[stationName\+'_tracks'\].length\] = "(.*)";`)
	results := r.FindAllStringSubmatch(body, -1)
	var songs []string
	for _, v := range results {
		songs = append(songs, v[1])
	}
	return songs
}

func parseBumps(body string) []string {
	r := regexp.MustCompile(`bumpsArray\[bumpsArray\.length\] = "(.*)";`)
	results := r.FindAllStringSubmatch(body, -1)
	var songs []string
	for _, v := range results {
		songs = append(songs, v[1])
	}
	return songs
}

func DownloadSongs() {
	for _, station := range GetStations() {
		if _, err := os.Stat(station); os.IsNotExist(err) {
			os.Mkdir(station, os.ModeDir)
		}
		songs, err := GetSongs(station)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, song := range songs {
			if _, err := os.Stat(station + "/" + song + ".mp3"); !os.IsNotExist(err) {
				fmt.Println(song + " exists for station " + station)
				continue
			}
			fmt.Println("Downloading song " + song + " from " + station)
			err = downloadSong(station, song)
			if err != nil {
				fmt.Println("failed to download song, error:", err)
			}
		}
	}
}

func downloadSong(station string, song string) error {
	songURL := FormatSong(station, song)
	resp, err := http.Get(songURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(station+"/"+song+".mp3", buf, fs.ModeAppend)
	if err != nil {
		return err
	}
	return nil
}
