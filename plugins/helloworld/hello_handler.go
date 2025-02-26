package helloworld

import (
	"fmt"
	"math/rand/v2"

	"github.com/anthdm/superkit/kit"
)

func handleHello(kit *kit.Kit) error {
	message := fmt.Sprintf("Hello world! %d", rand.IntN(100))
	createMessage(message)
	return kit.Render(Helloworld())
}

func handleAuthenticatedHello(kit *kit.Kit) error {
	messages, err := listMessages()
	if err != nil {
		panic(err)
	}
	fmt.Println(messages)
	return kit.Render(HelloworldAuth())
}
