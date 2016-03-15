package command

import (
	"github.com/timakin/op/client"

	"fmt"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

func CmdNotification(c *cli.Context) {
	notifications := client.GetNotifications()
	for _, notification := range notifications {
		color.Cyan("-------------------------------------------------")
		fmt.Print("\n")
		fmt.Print("Title: \t" + *notification.Subject.Title)
		fmt.Print("\n")
		fmt.Print("Url: \t" + *notification.Subject.URL)
		fmt.Print("\n")
	}
}