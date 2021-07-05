package speedtest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"time"
)

// Server information
type Server struct {
	URL      string        `xml:"url,attr" json:",omitempty"`
	Lat      string        `xml:"lat,attr" json:"-"`
	Lon      string        `xml:"lon,attr" json:"-"`
	Name     string        `xml:"name,attr" json:"name"`
	Country  string        `xml:"country,attr" json:"country"`
	Sponsor  string        `xml:"sponsor,attr" json:"sponsor"`
	ID       string        `xml:"id,attr" json:"id"`
	URL2     string        `xml:"url2,attr" json:",omitempty"`
	Host     string        `xml:"host,attr" json:"host"`
	Distance float64       `json:",omitempty"`
	Latency  time.Duration `json:",omitempty"`
	DLSpeed  float64       `json:",omitempty"`
	ULSpeed  float64       `json:",omitempty"`
}

// ServerList list of Server
type ServerList struct {
	Servers []*Server `xml:"servers>server" json:"servers"`
}

// Servers for sorting servers.
type Servers []*Server

// ByDistance for sorting servers.
type ByDistance struct {
	Servers
}

// Len finds length of servers. For sorting servers.
func (svrs Servers) Len() int {
	return len(svrs)
}

// Swap swaps i-th and j-th. For sorting servers.
func (svrs Servers) Swap(i, j int) {
	svrs[i], svrs[j] = svrs[j], svrs[i]
}

// Less compares the distance. For sorting servers.
func (b ByDistance) Less(i, j int) bool {
	return b.Servers[i].Distance < b.Servers[j].Distance
}

// FetchServerList retrieves a list of available servers or a specific server if serverId is specified
func FetchServerList(serverId *int) (ServerList, error) {
	// Fetch xml server data
	params := ""
	if serverId != nil {
		params = "?serverid=" + strconv.Itoa(*serverId)
	}
	resp, err := client.Get("https://cli.speedtest.net/api/cli/config" + params)
	if err != nil {
		return ServerList{}, errors.New("failed to retrieve speedtest servers")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ServerList{}, errors.New("failed to read response body")
	}
	defer resp.Body.Close()

	list := ServerList{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		return ServerList{}, errors.New("failed to parse response body")
	}

	for _, s := range list.Servers {
		s.URL = "http://" + s.Host + "/speedtest/upload.php"
	}

	if len(list.Servers) <= 0 {
		return list, errors.New("unable to retrieve server list")
	}

	return list, nil
}

func distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radius := 6378.137

	a1 := lat1 * math.Pi / 180.0
	b1 := lon1 * math.Pi / 180.0
	a2 := lat2 * math.Pi / 180.0
	b2 := lon2 * math.Pi / 180.0

	x := math.Sin(a1)*math.Sin(a2) + math.Cos(a1)*math.Cos(a2)*math.Cos(b2-b1)
	return radius * math.Acos(x)
}

// FindServer finds server by serverID
func (l *ServerList) FindServer(sid int) (Servers, error) {
	servers := Servers{}

	if len(l.Servers) <= 0 {
		return servers, errors.New("no servers available")
	}

	for _, s := range l.Servers {
		id, _ := strconv.Atoi(s.ID)
		if sid == id {
			servers = append(servers, s)
		}
	}

	if len(servers) == 0 {
		servers = append(servers, l.Servers[0])
	}

	return servers, nil
}

// String representation of ServerList
func (l *ServerList) String() string {
	slr := ""
	for _, s := range l.Servers {
		slr += s.String()
	}
	return slr
}

// String representation of Server
func (s *Server) String() string {
	return fmt.Sprintf("[%4s] %8.2fkm \n%s (%s) by %s\n", s.ID, s.Distance, s.Name, s.Country, s.Sponsor)
}

// CheckResultValid checks that results are logical given UL and DL speeds
func (s Server) CheckResultValid() bool {
	return !(s.DLSpeed*100 < s.ULSpeed) || !(s.DLSpeed > s.ULSpeed*100)
}
