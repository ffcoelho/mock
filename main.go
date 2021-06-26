package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"text/template"
	"time"
)

type MockRoute struct {
	File    string
	Path    string
	Methods []string
}

type MockEndpoints struct {
	Items []MockRoute
}

const port int = 9000
const mockDir string = "mocks"
const endpoint string = "endpoint.json"
const endpointsTemplate = `{{range .Items}}{{.Path}}	{{.Methods}}
{{end}}
`

var methods []string = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
var routes []MockRoute

func main() {
	setupCloseHandler()
	readMocks()

	http.HandleFunc("/", handler)

	fmt.Printf("\nMock started.\n")

	fmt.Printf("\nListening on http://localhost:%d\n", port)

	ips := listIPs()
	for _, ip := range ips {
		fmt.Printf("          or http://%s:%d\n", ip, port)
	}

	fmt.Printf("\nEndpoints:\n")
	printEndpoints()

	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, nil)

	fmt.Println("here")
}

func handler(w http.ResponseWriter, req *http.Request) {
	for _, route := range routes {
		if route.Path == req.URL.Path {
			for _, method := range route.Methods {
				if method == req.Method {
					now := time.Now()
					fmt.Printf("%s %s %s\n", now.Format("15:04:05"), req.Method, req.URL.Path)
					result := readEndpointContent(route.File, method)
					json.NewEncoder(w).Encode(result)
					return
				}
			}
		}
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func readMocks() error {
	dir := fmt.Sprintf(".%v%s", string(os.PathSeparator), mockDir)

	e := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(f.Name()) == ".json" && strings.HasSuffix(path, endpoint) {
			var route MockRoute
			jsonFilePath := strings.SplitAfterN(path, string(os.PathSeparator), 2)[1]
			route.File = jsonFilePath
			toTrim := fmt.Sprintf("%v%s", string(os.PathSeparator), endpoint)
			route.Path = fmt.Sprintf("%v%s", string(os.PathSeparator), strings.TrimSuffix(jsonFilePath, toTrim))

			// TODO fix path handle
			if route.Path == "/endpoint.json" {
				route.Path = "/"
			}

			route.Methods = readRouteMethods(jsonFilePath)
			if len(route.Methods) > 0 {
				routes = append(routes, route)
			}
		}

		return err
	})

	if e != nil {
		panic(e)
	}

	return nil
}

func readRouteMethods(filePath string) []string {
	path := fmt.Sprintf(".%v%s%v%s", string(os.PathSeparator), mockDir, string(os.PathSeparator), filePath)
	jsonFile, err := os.Open(path)

	if err != nil {
		log.Fatalf(err.Error())
	}

	defer jsonFile.Close()

	bytevalue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	json.Unmarshal(bytevalue, &result)

	routeMethods := make([]string, 0)
	for key := range result {
		if validMethodKey(key) {
			routeMethods = append(routeMethods, key)
		}
	}

	return routeMethods
}

func readEndpointContent(file, method string) interface{} {
	path := fmt.Sprintf(".%v%s%v%s", string(os.PathSeparator), mockDir, string(os.PathSeparator), file)
	jsonFile, err := os.Open(path)

	if err != nil {
		log.Fatalf(err.Error())
	}

	defer jsonFile.Close()

	bytevalue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	json.Unmarshal(bytevalue, &result)

	return result[method]
}

func validMethodKey(key string) bool {
	for _, method := range methods {
		if method == key {
			return true
		}
	}

	return false
}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\rMock is down.")
		os.Exit(0)
	}()
}

func listIPs() []net.IP {
	ifaces, err := net.Interfaces()

	if err != nil {
		panic(err)
	}

	ips := make([]net.IP, 0)
	for _, i := range ifaces {
		addrs, e := i.Addrs()

		if e != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			ip = ip.To4()
			if ip == nil || ip.String() == "127.0.0.1" {
				continue
			}

			ips = append(ips, ip)
		}
	}

	return ips
}

func printEndpoints() {
	data := MockEndpoints{
		Items: routes,
	}

	t := template.New("test")
	t, _ = t.Parse(endpointsTemplate)
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 8, ' ', 0)

	if err := t.Execute(w, data); err != nil {
		log.Fatal(err)
	}

	w.Flush()
}
