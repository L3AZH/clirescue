package trackerapi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"app/cmdutil"
	"app/user"
)

var (
	URL          string     = "https://www.pivotaltracker.com/services/v5/me"
	FileLocation string     = "./.tracker"
	FileCache	string = "./.cache"
	currentUser  *user.User = user.New()
	Stdout       *os.File   = os.Stdout
)

func Me() {
	setCredentials()
	parse(makeRequest())
	fmt.Printf("File location: %v", FileLocation)
	os.WriteFile(FileLocation, []byte(currentUser.APIToken), 0644)
}

func makeRequest() []byte {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", URL, nil)
	req.SetBasicAuth(currentUser.Username, currentUser.Password)
	resp, _ := client.Do(req)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("\n****\nAPI response: \n%s\n", string(body))
	return body
}

func parse(body []byte) {
	var meResp = new(MeResponse)
	err := json.Unmarshal(body, &meResp)
	if err != nil {
		fmt.Println("error:", err)
	}

	currentUser.APIToken = meResp.APIToken
	fmt.Printf("%v", currentUser)
}

func setCredentials() {
	fmt.Println("Loading Username and Password from cache....")
	exist, username, password := checkUserExistInCache()
	if exist {
		currentUser.Login(username, password)
	} else {
		os.Create(FileCache)
		fmt.Fprint(Stdout, "\nUsername: ")
		username = cmdutil.ReadLine()
		cmdutil.Silence()
		fmt.Fprint(Stdout, "Password: ")
		password = cmdutil.ReadLine()
		fmt.Printf("Username %v - Password: %v \n", username, password)
		f, err := os.OpenFile(FileCache, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Open file Cache: %v",err)
		}
		if _, err := f.Write([]byte(username+"\n"+password)); err != nil {
			fmt.Printf("Write Cache: %v",err)
		}
		if err := f.Close(); err != nil {
			fmt.Printf("Write Cache: %v",err)
		}
		currentUser.Login(username, password)
		cmdutil.Unsilence()
	}
	
}

func checkUserExistInCache() (bool, string, string){
	f, err := os.Open(FileCache)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return false, "",""
	}
	defer f.Close()
	bufReader := bufio.NewReader(f)
	usernameData,_,errUsername := bufReader.ReadLine()
	if errUsername != nil {
		fmt.Printf("error: %v\n",errUsername)
		return false, "" ,""
	}
	username := strings.Trim(string(usernameData)," ")
	passwordData,_,errPassword := bufReader.ReadLine()
	if errPassword != nil {
		fmt.Printf("error: %v\n",errUsername)
		return false, "" ,""
	}
	password := strings.Trim(string(passwordData)," ")
	return true, username, password
}


type MeResponse struct {
	APIToken string `json:"api_token"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Initials string `json:"initials"`
	Timezone struct {
		Kind      string `json:"kind"`
		Offset    string `json:"offset"`
		OlsonName string `json:"olson_name"`
	} `json:"time_zone"`
}
