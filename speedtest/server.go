package speedtest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
)

// Server information
type Server struct {
	URL     string  `xml:"url,attr" json:",omitempty"`
	Lat     string  `xml:"lat,attr" json:"-"`
	Lon     string  `xml:"lon,attr" json:"-"`
	Name    string  `xml:"name,attr" json:"name"`
	Country string  `xml:"country,attr" json:"country"`
	Sponsor string  `xml:"sponsor,attr" json:"sponsor"`
	ID      string  `xml:"id,attr" json:"id"`
	Host    string  `xml:"host,attr" json:"host"`
	Latency float64 `json:"latency,omitempty"`
	DLSpeed float64 `json:"dl_speed,omitempty"`
	ULSpeed float64 `json:"ul_speed,omitempty"`
}

// ServerList list of Server
type ServerList struct {
	Servers []*Server `xml:"servers>server" json:"servers"`
}

// Servers for sorting servers.
type Servers []*Server

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
	return fmt.Sprintf("[%4s] \n%s (%s) by %s\n", s.ID, s.Name, s.Country, s.Sponsor)
}

// CheckResultValid checks that results are logical given UL and DL speeds
func (s Server) CheckResultValid() bool {
	return !(s.DLSpeed*100 < s.ULSpeed) || !(s.DLSpeed > s.ULSpeed*100)
}
