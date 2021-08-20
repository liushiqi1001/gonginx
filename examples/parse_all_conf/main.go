package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/liushiqi1001/gonginx"
	"github.com/liushiqi1001/gonginx/parser"
)

type Server struct {
	ServiceName []string
	Locations   []*Location
}

type Location struct {
	Uri       string
	ProxyPass string
	UpStream  *UpstreamServer
}

type UpstreamServer struct {
	UpstreamName string
	Hosts        []string
}

func main() {

	confs := make([]*gonginx.Config, 0)
	usServers := make(map[string]*UpstreamServer)
	servers := make([]*Server, 0)

	fileName := "/Users/shiqi/go/src/github.com/liushiqi1001/gonginx/examples/nginx_conf"
	pattern := ".*.conf$"

	reg, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println(err)
		return
	}

	filepath.Walk(fileName,
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}

			if f.IsDir() {
				return nil
			}

			// 匹配目录
			matched := reg.MatchString(f.Name())

			if matched {
				fmt.Println(path)
				content, err := parser.ReadConfContent("", path)

				if err != nil {
					return err
				}

				p := parser.NewStringParser(content)

				defer func() {
					if err := recover(); err != nil {
						fmt.Println(err)
					}
				}()

				c := p.Parse()
				confs = append(confs, c)

			}

			return nil

		})

	fmt.Println("length of confs:", len(confs))

	for _, c := range confs {
		ups := c.FindUpstreams()
		for _, up := range ups {
			tmpUp := &UpstreamServer{
				UpstreamName: up.UpstreamName,
				Hosts:        make([]string, 0),
			}
			ss := up.UpstreamServers
			for _, s := range ss {
				tmpUp.Hosts = append(tmpUp.Hosts, s.Address)
				//fmt.Println(up.UpstreamName + "------------" + s.Address)
			}

			usServers[tmpUp.UpstreamName] = tmpUp
		}
	}

	//fmt.Println("usServers:", usServers[0])

	for _, c := range confs {
		ds := c.FindDirectives("server")
		for _, direc := range ds {
			secondDirecs := direc.GetBlock().FindDirectives("server_name")
			for _, sd := range secondDirecs {
				//var param string
				//for _, p := range sd.GetParameters() {
				//	param = fmt.Sprintf("%s----%s", param, p)
				//
				//}
				//fmt.Println(sd.GetName() + "-----" + param)
				tmpServer := &Server{
					ServiceName: make([]string, 0),
					Locations:   make([]*Location, 0),
				}
				params := sd.GetParameters()
				for _, p := range params {
					tmpServer.ServiceName = append(tmpServer.ServiceName, p)
				}

				secondDirecs := direc.GetBlock().FindDirectives("location")
				locas := make([]*Location, 0)
				for _, sd := range secondDirecs {
					var param string
					for _, p := range sd.GetParameters() {
						if param == "" {
							param = p
							continue
						}
						param = fmt.Sprintf("%s %s", param, p)
					}
					tmplocal := &Location{
						Uri: param,
					}
					thirdDirec := sd.GetBlock().FindDirectives("proxy_pass")
					for _, td := range thirdDirec {
						tmplocal.ProxyPass = td.GetParameters()[0]
						//fmt.Print(td.GetName()+ "-----")
						//fmt.Println(td.GetParameters())
					}
					//fmt.Println(sd.GetParameters())
					locas = append(locas, tmplocal)

				}
				tmpServer.Locations = locas
				servers = append(servers, tmpServer)
			}
		}
	}

	fmt.Println("the length of servers:", len(servers))

	// 匹配 location
	for _, s := range servers {
		for _, l := range s.Locations {
			split := strings.Split(l.ProxyPass, "://")
			if len(split) != 2 {
				continue
			}

			if ups, ok := usServers[split[1]]; ok {
				l.UpStream = ups
			}
		}
	}

	datas, _ := json.Marshal(servers[:10])
	fmt.Println(string(datas))
}
