package speedtest

import (
	"io/ioutil"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

var dlSizes = [...]int{350, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500, 4000}
var ulSizes = [...]int{100, 300, 500, 800, 1000, 1500, 2500, 3000, 3500, 4000} //kB

// DownloadTest executes the test to measure download speed
func (s *Server) DownloadTest(savingMode bool) error {
	sockets, err := GetSockets(s.Host, 10)
	if err != nil {
		return err
	}
	eg := errgroup.Group{}
	stats := make([]float64, len(sockets))
	for i, socket := range sockets {
		index := i
		s := socket
		eg.Go(func() error {
			duration := time.Duration(1e10)
			test := time.Duration(7e9)
			warmup := time.Duration(3e9)
			content := make([]byte, 4096)
			count := int64(0)
			t1 := time.Now()
			t2 := time.Now()
			for t2.Sub(t1) < duration {
				s.Write([]byte("DOWNLOAD 4096\n"))
				length, err := s.Read(content)
				if err == nil && t2.Sub(t1) > warmup {
					count += int64(length) + (int64)(math.Ceil(float64(length/1460)))*54 // 40 bytes L3 + L4 header and 14 bytes L2 header
				}
				t2 = time.Now()
			}
			stats[index] = float64(count) * 8 / test.Seconds()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	CloseSockets(sockets)

	dlSpeed := float64(0)
	for _, speed := range stats {
		dlSpeed += speed
	}

	s.DLSpeed = dlSpeed / 1024 / 1024
	return nil
}

// UploadTest executes the test to measure upload speed
func (s *Server) UploadTest(savingMode bool) error {
	sockets, err := GetSockets(s.Host, 10)
	if err != nil {
		return err
	}
	eg := errgroup.Group{}
	stats := make([]float64, len(sockets))
	msg := []byte("UPLOAD 4096\n")
	content := []byte(strings.Repeat("1", 4096-len(msg)-1) + "\n")

	for i, socket := range sockets {
		index := i
		s := socket
		eg.Go(func() error {
			duration := time.Duration(1e10)
			test := time.Duration(7e9)
			warmup := time.Duration(3e9)
			response := make([]byte, 64)
			count := int64(0)
			t1 := time.Now()
			t2 := time.Now()
			for t2.Sub(t1) < duration {
				s.Write(msg)
				s.Write(content)
				length, err := s.Read(response)
				if err == nil && t2.Sub(t1) > warmup {
					bytes, err := strconv.ParseInt(strings.Split(string(response)[:length], " ")[1], 10, 64)
					if err == nil {
						count += bytes + (int64)(math.Ceil(float64(len(content)/1460)))*54
					}
				}
				t2 = time.Now()
			}
			stats[index] = float64(count) * 8 / test.Seconds()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	CloseSockets(sockets)

	ulSpeed := float64(0)
	for _, speed := range stats {
		ulSpeed += speed
	}

	s.ULSpeed = ulSpeed / 1024 / 1024
	return nil
}

func dlWarmUp(dlURL string) error {
	size := dlSizes[2]
	xdlURL := dlURL + "/random" + strconv.Itoa(size) + "x" + strconv.Itoa(size) + ".jpg"

	resp, err := client.Get(xdlURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)

	return nil
}

func ulWarmUp(ulURL string) error {
	size := ulSizes[4]
	v := url.Values{}
	v.Add("content", strings.Repeat("0123456789", size*100-51))

	resp, err := client.PostForm(ulURL, v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)

	return nil
}

func downloadRequest(dlURL string, w int) error {
	size := dlSizes[w]
	xdlURL := dlURL + "/random" + strconv.Itoa(size) + "x" + strconv.Itoa(size) + ".jpg"

	resp, err := client.Get(xdlURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)

	return nil
}

func uploadRequest(ulURL string, w int) error {
	size := ulSizes[w]
	v := url.Values{}
	v.Add("content", strings.Repeat("0123456789", size*100-51))

	resp, err := client.PostForm(ulURL, v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)

	return nil
}

// PingTest executes test to measure latency
func (s *Server) PingTest() error {
	sockets, err := GetSockets(s.Host, 1)
	if err != nil {
		return err
	}
	socket := sockets[0]
	total := int64(0)
	count := int64(0)
	for i := 0; i < 10; i++ {
		t1 := time.Now().UnixNano()
		str := "PING " + strconv.FormatInt(t1, 10) + "\n"
		socket.Write([]byte(str))
		result := make([]byte, 40)
		_, err := socket.Read(result)
		if err == nil {
			t2 := time.Now().UnixNano()
			count++
			total += (t2 - t1)
		}
	}
	if count > 0 {
		s.Latency = time.Duration(total / count)
	}
	CloseSockets(sockets)
	return nil
}
