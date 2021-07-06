package speedtest

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// DownloadTest executes the test to measure download speed
func (s *Server) DownloadTest() error {
	sockets, err := GetSockets(s.Host, 10)
	if err != nil {
		return err
	}
	if len(sockets) == 0 {
		return errors.New("Failed to do download test, unabled to connect " + s.Host)
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
			content := make([]byte, 65536)
			count := int64(0)
			t1 := time.Now()
			t2 := time.Now()
			for t2.Sub(t1) < duration {
				s.Write([]byte("DOWNLOAD 65536\n"))
				length, err := s.Read(content)
				if err == nil && t2.Sub(t1) > warmup {
					count += int64(length) + (int64(math.Ceil(float64(length/1460))))*54 // 40 bytes L3 + L4 header and 14 bytes L2 header
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
func (s *Server) UploadTest() error {
	sockets, err := GetSockets(s.Host, 10)
	if err != nil {
		return err
	}
	if len(sockets) == 0 {
		return errors.New("Failed to do upload test, unabled to connect " + s.Host)
	}
	eg := errgroup.Group{}
	stats := make([]float64, len(sockets))
	line1 := "UPLOAD 65536 0\n"
	msg := []byte(line1 + strings.Repeat("1", 65536-len(line1)-1) + "\n")

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
				length, err := s.Read(response)
				if err == nil && t2.Sub(t1) > warmup {
					bytes, err := strconv.ParseInt(strings.Split(string(response)[:length], " ")[1], 10, 64)
					if err == nil {
						count += bytes + (int64(math.Ceil(float64(len(msg)/1460))))*54
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

// PingTest executes test to measure latency
func (s *Server) PingTest() error {
	sockets, err := GetSockets(s.Host, 1)
	if err != nil {
		return err
	}
	if len(sockets) == 0 {
		return errors.New("Failed to do latency test, unabled to connect " + s.Host)
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
		s.Latency = float64(time.Duration(total/count).Nanoseconds()) / 1e6
	}
	CloseSockets(sockets)
	return nil
}
