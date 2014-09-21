/* greet.go */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Application struct {
	Name     string `json:"name"`
	Creator  User   `json:"_creator"`
	Running  bool   `json:"running"`
	DockerID string `json:"dockerId"`
	Ip       string `json:"ip"`
}

type Token struct {
	Value string `json:"value"`
}

func askQuestion(title string) string {
	var answer string
	fmt.Println(title)
	_, err := fmt.Scanln(&answer)
	if err != nil {
		panic(err)
	}
	return answer
}

func userJsonFromParams(username string, password string) io.Reader {
	user := &User{username, password}
	buf, _ := json.Marshal(user)
	body := bytes.NewBuffer(buf)
	return body
}

func tokenFromJson(jsonString []byte) *Token {
	token := &Token{}
	err := json.Unmarshal(jsonString, &token)
	if err != nil {
		panic(err)
	}
	return token
}

func getToolFolder() string {
	user, err := user.Current()

	if err != nil {
		panic(err)
	}
	//TODO os specific, fix me
	path := "/home/" + user.Username + "/.modernweb"
	return path
}

func createUserFolder() {
	path := getToolFolder()
	_ = os.Mkdir(path, 0777)
}

func saveTokenToDisk(token Token) error {
	buf, _ := json.Marshal(token)
	path := getToolFolder() + "/currentUser"
	err := ioutil.WriteFile(path, buf, 0644)
	return err
}

func getTokenFromDisk() string {
	token := &Token{}

	path := getToolFolder() + "/currentUser"

	file, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(file, &token)

	if err != nil {
		panic(err)
	}
	return token.Value
}

func applicationJsonFromName(name string) io.Reader {
	application := &Application{}
	application.Name = name
	buf, _ := json.Marshal(application)
	body := bytes.NewBuffer(buf)
	return body
}

func main() {

	api := "http://192.168.33.10:3000"

	createUserFolder()

	app := cli.NewApp()
	app.Name = "mw"
	app.Version = "0.0.1"
	app.Usage = "CLI interface for interacting with ModernWeb API"
	app.Action = func(c *cli.Context) {
		println("Hello friend!")
	}

	app.Commands = []cli.Command{
		{
			Name:      "register",
			ShortName: "r",
			Usage:     "Register as a user",
			Action: func(c *cli.Context) {

				username := askQuestion("What username would you like to have?")

				password := askQuestion("Enter your password")

				body := userJsonFromParams(username, password)

				_, err := http.Post(api+"/users/register", "application/json", body)
				if err != nil {
					fmt.Println("Something went very wrong")
					panic(err)
				}
				fmt.Println("Your account have been created and you have been logged in as '" + username + "'")
			},
		},
		{
			Name:      "login",
			ShortName: "l",
			Usage:     "Login to your user account",
			Action: func(c *cli.Context) {
				username := askQuestion("What is your username?")

				password := askQuestion("Enter your password")

				body := userJsonFromParams(username, password)

				response, err := http.Post(api+"/users/login", "application/json", body)
				if err != nil {
					fmt.Println("Something went very wrong")
					panic(err)
				}
				contents, err := ioutil.ReadAll(response.Body)

				token := tokenFromJson(contents)
				saveTokenToDisk(*token)

				if response.StatusCode != 200 {
					fmt.Println("Wrong username/password, please try again")
				} else {
					fmt.Println("You have been logged in as '" + username + "'")
				}
			},
		},
		{
			Name:      "init",
			ShortName: "i",
			Usage:     "Initialize new application in current folder",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					fmt.Println("You need to provide application name as first argument")
					return
				}

				token := getTokenFromDisk()

				if token == "" {
					fmt.Println("You need to be logged in")
					return
				}

				applicationName := c.Args()[0]

				body := applicationJsonFromName(applicationName)

				fmt.Println(body)

				req, reqErr := http.NewRequest("POST", api+"/applications/init", body)
				if reqErr != nil {
					panic(reqErr)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("x-token", token)

				client := &http.Client{}
				response, err := client.Do(req)

				if err != nil {
					fmt.Println("Something went very wrong")
					panic(err)
				}

				if response.StatusCode == 201 {
					fmt.Println("Created application '" + applicationName + "'")
				}
				if response.StatusCode == 401 {
					fmt.Println("You are not logged in")
				}
				if response.StatusCode == 400 {
					fmt.Println("Application '" + applicationName + "' already exists.")
					fmt.Println("Please chose another name")
				}
				if response.StatusCode != 201 && response.StatusCode != 400 && response.StatusCode != 401 {
					fmt.Println("Something went very wrong")
					panic(response)
				}
			},
		},
	}

	app.Run(os.Args)
}
