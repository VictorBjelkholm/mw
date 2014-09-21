/* greet.go */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"net/http"
	"os"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

func saveUserLocally() bool {
	return true
}

func main() {

	api := "http://192.168.33.10:3000"

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

				_, err := http.Post(api+"/register", "application/json", body)
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

				response, err := http.Post(api+"/login", "application/json", body)
				if err != nil {
					fmt.Println("Something went very wrong")
					panic(err)
				}
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
				println("completed task: ", c.Args().First())
			},
		},
	}

	app.Run(os.Args)
}
