package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type MockRoute struct {
	File    string
	Path    string
	Methods []string
}

const port int = 9000
const mockDir string = "mocks"
const endpoint string = "endpoint.json"

var methods []string = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
var routes []MockRoute

func main() {
	readMocks()

	http.HandleFunc("/", handler)

	fmt.Printf("Mock started. Listening on http://localhost:%d\n", port)

	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	for _, route := range routes {
		if route.Path == req.URL.Path {
			for _, method := range route.Methods {
				if method == req.Method {
					result := readEndpointContent(route.File, method)
					json.NewEncoder(w).Encode(result)
					return
				}
			}
		}
	}
	fmt.Println("Hi")
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

	for _, route := range routes {
		fmt.Printf("%s -", route.Path)
		for _, method := range route.Methods {
			fmt.Printf(" %s", method)
		}
		fmt.Println()
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
