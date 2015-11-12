package main // import "fknsrs.biz/p/xacom/cmd/xacom"

import (
	"fmt"
	"os"

	"fknsrs.biz/p/xacom"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app     = kingpin.New("xacom", "Send pager messages via a XACOM system.")
	server  = app.Flag("server", "Server address for XACOM system.").OverrideDefaultFromEnvar("XACOM_URL").Required().String()
	number  = app.Arg("number", "Pager number to send the message to.").Required().Uint64()
	message = app.Arg("message", "Message content.").Required().String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	c, err := xacom.NewClient(*server)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	go c.Run()
	ok, err := c.SendMessage(fmt.Sprintf("%010d", *number), *message)
	if err != nil {
		panic(err)
	}

	if !ok {
		fmt.Printf("failed sending message\n")
		os.Exit(1)
	} else {
		fmt.Printf("message queued successfully\n")
	}
}
