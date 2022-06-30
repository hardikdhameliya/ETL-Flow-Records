package fetch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gigamon.com/config"
	"gigamon.com/enrich"
	"github.com/olivere/ndjson"
)

// Fetch struct to fetch data files
type fetch struct {
	timeout  time.Duration
	tick     *time.Ticker
	path     string
	maxFiles int
}

// Create a new instance of fetch with specific configuration
func NewFetch() *fetch {

	inter, err := config.GetSectionParamDuration("fetch.interval")
	if err != nil {
		inter = 5 * time.Second
	}

	to, err := config.GetSectionParamDuration("fetch.timeout")
	if err != nil {
		to = 10 * time.Second
	}

	pt, err := config.GetSectionParamString("fetch.file_path")
	if err != nil {
		pt, _ = os.UserHomeDir()
	}

	mf, err := config.GetSectionParamInt("fetch.maximum_file")
	if err != nil {
		mf = 5
	}
	p := &fetch{
		timeout:  to,
		tick:     time.NewTicker(inter),
		path:     pt,
		maxFiles: mf,
	}
	return p
}

// Fetch the data files at specific interval, enrich it and delete files from the disk
func (p *fetch) Run(c chan map[string][]interface{}, done chan bool) {
	// timeout configuration
	toutStart := func(t bool, tf *time.Ticker) *time.Ticker {
		if !t {
			return time.NewTicker(p.timeout)
		}
		return tf
	}
	tout := false
	// set new ticker for timeout
	tf := new(time.Ticker)
	for {
		select {
		case t := <-p.tick.C: // this portion of code get executed at specific interval
			fmt.Printf("Read file(s) at  %+v\n", t)
			m, _ := listFiles(p.path, p.maxFiles)
			flows, err := readFiles(m)
			if err != nil {
				return
			}
			if len(flows) != 0 {
				select {
				case c <- flows: // push documents into the channel and delete files from the disk
					fmt.Println("Documents are pushed into the channel")
					deleteFiles(m)
					tout = false
					tf.Stop()
				default: // If channel is blocked, start the timeout timer
					tf = toutStart(tout, tf)
					tout = true
				}
			} else {
				// If flows record is not there, start the timeout timer
				tf = toutStart(tout, tf)
				tout = true
			}
		case <-tf.C: // Timeout occured
			fmt.Println("Terminated due to timeout")
			done <- true
		}
	}
}

// List all the files in the dir
func listFiles(dir string, maxFiles int) (map[string]bool, error) {
	toVisit := make(map[string]bool)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if len(toVisit) >= maxFiles {
			return errors.New("maxFilesReached")
		}
		if err != nil {
			return err
		}
		if filepath.Ext(path) != ".ndjson" {
			return nil
		}
		toVisit[path] = false
		fmt.Printf("Visited file : %q\n", path)
		return nil
	})
	return toVisit, err
}

// Read files, generate flow record and enrich it
func readFiles(m map[string]bool) (map[string][]interface{}, error) {
	if m == nil {
		return nil, errors.New("No input")
	}
	flows := make(map[string][]interface{})
	for f, b := range m {
		// if file has read already
		if b {
			continue
		}
		r, err := os.Open(f)
		defer r.Close()

		if err != nil {
			fmt.Println("Cannot open a file", f)
			continue
		}

		//set the flag for sucessful read
		m[f] = true
		flow_id := strings.Split(f[strings.LastIndex(f, "/")+1:], "T")[0]
		s := ndjson.NewReader(r)
		for s.Next() {
			var flow Flow
			if err := s.Decode(&flow); err != nil {
				fmt.Println("Decode failed:", err)
				continue
			}
			if flow.SrcMac != "" {
				flow.SrcVendor = getVendor(flow.SrcMac)
			}
			if flow.DstMac != "" {
				flow.DstVendor = getVendor(flow.DstMac)
			}
			flows[flow_id] = append(flows[flow_id], flow)
		}
	}
	return flows, nil
}

// Get vendor info
func getVendor(mac string) string {
	if mac == "" {
		return ""
	}

	mac = strings.Join(strings.SplitN(mac, ":", 4)[:3], "")
	s, err := enrich.GetVendor(strings.ToUpper(mac))
	if err != nil {
		return ""
	}

	return s
}

// Delete files those are being read
func deleteFiles(m map[string]bool) {
	if m == nil {
		return
	}
	for f, b := range m {
		if !b {
			continue
		}
		err := os.Remove(f)
		if err != nil {
			fmt.Println("File cannot be removed")
		} else {
			fmt.Printf("Removed file : %s\n", f)
		}
	}
}
