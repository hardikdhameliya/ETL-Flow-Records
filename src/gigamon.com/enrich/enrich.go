package enrich

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"gigamon.com/config"
)

var (
	db DataInterface
)

//Initialize database with specific parameters and load enrichment info
func init() {
	var add, net, url string
	var it, ct, rt, wt time.Duration
	var mi int
	var err error

	if add, err = config.GetSectionParamString("vendor.database.address"); err != nil {
		add = "localhost:6379"
	}

	if net, err = config.GetSectionParamString("vendor.database.network"); err != nil {
		net = "tcp"
	}

	if mi, err = config.GetSectionParamInt("vendor.maxidleconn"); err != nil {
		mi = 3
	}

	if it, err = config.GetSectionParamDuration("vendor.idletimeout"); err != nil {
		it = 5 * time.Second
	}

	if ct, err = config.GetSectionParamDuration("vendor.conntimeout"); err != nil {
		ct = 5 * time.Second
	}

	if rt, err = config.GetSectionParamDuration("vendor.readtimeout"); err != nil {
		rt = 5 * time.Second
	}

	if wt, err = config.GetSectionParamDuration("vendor.writetimeout"); err != nil {
		wt = 5 * time.Second
	}

	if url, err = config.GetSectionParamString("vendor.url"); err != nil {
		url = "http://standards-oui.ieee.org/oui.txt"
	}

	db = NewRedisImpl(add, net, mi, it, ct, rt, wt)

	if err := loadFile(url); err != nil {
		fmt.Println("Cannot load the enrichment file into Database")
	}
}

// Load files from the url
func loadFile(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	m, err := parseFile(resp)
	if err != nil {
		return err
	}
	err = db.SetBulk(*m)
	if err != nil {
		return err
	}

	return nil
}

// Parse file to get vendor info
func parseFile(resp *http.Response) (*map[string]string, error) {
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	scan := bufio.NewScanner(bytes.NewReader(s))
	scan.Split(bufio.ScanLines)
	line := ""
	oui_info := make(map[string]string)
	r := regexp.MustCompile("(?P<Mac>[\\w]*)(?P<base>[\\s]*\\([\\s]*base[\\s]*16[\\s]*\\)[\\s]*)(?P<company>.*)")
	for scan.Scan() {
		line = scan.Text()
		if ss := r.FindStringSubmatch(line); len(ss) != 0 {
			oui_info[ss[1]] = ss[3]
		}
	}

	return &oui_info, nil
}

// Get the vendor info from mac
func GetVendor(mac string) (string, error) {
	if mac == "" {
		return "", errors.New("Empty mac address")
	}
	v, err := db.Get(mac)
	if err != nil {
		return "", err
	}
	return v, nil

}
