package helloworld

import (
	"github.com/anthdm/superkit/kit"
)

func handleHello(kit *kit.Kit) error {
	createMessage("")
	return kit.Render(Helloworld())
}

func handleAuthenticatedHello(kit *kit.Kit) error {
	return kit.Render(HelloworldAuth())
}
func handleReadHello(kit *kit.Kit) error {
	messages, err := listMessages()
	if err != nil {
		panic(err)
	}
	return kit.Render(HelloworldRead(messages))
}
