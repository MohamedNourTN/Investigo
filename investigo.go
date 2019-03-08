package main

import (
	"encoding/json"
	"fmt"
	color "github.com/fatih/color"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var sns = map[string]string{}
var snsCaseLower = map[string]string{}

func getPageSource(response *http.Response) string {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	return string(bodyBytes)
}

func httpRequest(url string) (
	response *http.Response, respondedURL string, err error) {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	client := &http.Client{}
	response, clientError := client.Do(request)
	if clientError == nil {
		respondedURL = response.Request.URL.String()
		err = nil
	} else {
		respondedURL = ""
		err = clientError
	}

	return
}

// Check if username is exist on snsName
func isUserExist(snsName string, username string, caseLower bool) bool {
	url := sns[snsName]
	if caseLower {
		url = snsCaseLower[strings.ToLower(snsName)]
	}

	response, respondedURL, err := httpRequest(strings.Replace(url, "?", username, 1))
	if err != nil {
		fmt.Print("You can not access " + snsName + " in your country. ")
		// fmt.Println(err)
		return false
	}

	snsName = strings.ToLower(snsName)

	switch snsName {
	case "wordpress":
		if respondedURL == url {
			return true
		}
		return false
	case "steam":
		if !strings.Contains(
			getPageSource(response),
			"The specified profile could not be found.") {
			return true
		}
		return false
	case "pinterest":
		if url == respondedURL || strings.Contains(respondedURL, username) {
			return true
		}
		return false
	case "gitlab":
		if url == respondedURL {
			return true
		}
		return false
	case "egloos":
		if !strings.Contains(
			getPageSource(response),
			"블로그가 존재하지 않습니다") {
			return true
		}
		return false
	}

	if response.StatusCode == 200 {
		return true
	}
	return false
}

// Array, element
func contains(array []string, str string) (bool, int) {
	for index, item := range array {
		if item == str {
			return true, index
		}
	}
	return false, 0 // Index
}

func loadSNSList() {
	jsonFile, err := os.Open("sites.json")
	
	if err != nil {
		udpateSNSList()
		jsonFile, _ = os.Open("sites.json")
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var snsInterface map[string]interface{}
	json.Unmarshal([]byte(byteValue), &snsInterface)
	// Json to map
	for k, v := range snsInterface {
		sns[k] = v.(string)
	}
}

func udpateSNSList() {
	fmt.Println("Updating sites.json")
	response, _, _ := httpRequest("https://raw.githubusercontent.com/tdh8316/Investigo/master/sites.json")
	jsonData := getPageSource(response)

	fileName := "sites.json"
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		if err = os.Remove(fileName); err != nil {
			panic(err)
		}
	}
	dataFile, _ := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	defer dataFile.Close()

	if _, err := dataFile.WriteString(jsonData); err != nil {
		fmt.Fprintf(color.Output, color.RedString("Failed to update data\n"))
	}
}

func main() {
	args := os.Args[1:]
	disableColor, _ := contains(args, "--no-color")
	disableQuiet, _ := contains(args, "--verbose")
	updateData, _ := contains(args, "--update")
	specificSite, siteIndex := contains(args, "--site")
	specifiedSite := ""
	if specificSite {
		specifiedSite = args[siteIndex+1]
	}

	if updateData {
		udpateSNSList()
	}

	loadSNSList()

	for _, username := range args {
		if isOpt, _ := contains([]string{"--no-color", "--verbose", specifiedSite, "--site", "-update"}, username); isOpt {
			continue
		}
		if disableColor {
			fmt.Printf("Searching username %s\n", username)
		} else {
			fmt.Fprintf(color.Output, "%s %s\n", color.HiMagentaString("Searching username"), username)
		}
		if specificSite {
			// Case ignore
			for k, v := range sns {
				snsCaseLower[strings.ToLower(k)] = v
			}
			if isUserExist(strings.ToLower(specifiedSite), username, true) {
				if disableColor {
					fmt.Printf(
						"[+] %s: %s\n", specifiedSite, strings.Replace(
							snsCaseLower[strings.ToLower(specifiedSite)], "?", username, 1))
				} else {
					fmt.Fprintf(color.Output,
						"[%s] %s: %s\n",
						color.HiGreenString("+"), color.HiWhiteString(specifiedSite),
						color.WhiteString(
							strings.Replace(snsCaseLower[strings.ToLower(specifiedSite)],
								"?", username, 1)))
				}
			} else {
				if disableColor {
					fmt.Printf(
						"[-] %s: Not found!\n", specifiedSite)
				} else {
					fmt.Fprintf(color.Output,
						"[%s] %s: %s\n",
						color.HiRedString("-"), color.HiWhiteString(specifiedSite),
						color.HiYellowString("Not found!"))
				}
			}
			break
		}

		fileName := username + ".txt"
		if _, err := os.Stat(fileName); !os.IsNotExist(err) {
			if err = os.Remove(fileName); err != nil {
				panic(err)
			}
		}
		resFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer resFile.Close()

		for site := range sns {
			if isUserExist(site, username, false) {
				if disableColor {
					fmt.Printf(
						"[+] %s: %s\n", site, strings.Replace(sns[site], "?", username, 1))
				} else {
					fmt.Fprintf(color.Output,
						"[%s] %s: %s\n",
						color.HiGreenString("+"), color.HiWhiteString(site),
						color.WhiteString(strings.Replace(sns[site], "?", username, 1)))
				}
				if _, err = resFile.WriteString(site + ": " + strings.Replace(sns[site], "?", username, 1) + "\n"); err != nil {
					panic(err)
				}
			} else {
				if !disableQuiet {
					continue
				}

				if disableColor {
					fmt.Printf(
						"[-] %s: Not found!\n", site)
				} else {
					fmt.Fprintf(color.Output,
						"[%s] %s: %s\n",
						color.HiRedString("-"), color.HiWhiteString(site),
						color.HiYellowString("Not found!"))
				}
			}
		}
		fmt.Println("\nYour search results have been saved to " + fileName)
	}
}
