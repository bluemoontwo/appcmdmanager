package appcmdmanager

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"sync"
)

type AppCmdInterface interface {
	ApplicationCommandStruct() *discordgo.ApplicationCommand
	ApplicationCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate)
	MessageComponentHandler() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
	ModalSubmitHandler() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func RegisterApplicationCommands(cmds []AppCmdInterface, s *discordgo.Session, guildID string) (appCmds []*discordgo.ApplicationCommand) {
	var (
		mutex sync.Mutex
		wg    sync.WaitGroup
	)

	for _, command := range cmds {
		wg.Add(1)

		go func(cmd AppCmdInterface) {
			appCmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd.ApplicationCommandStruct())
			if err != nil {
				log.Panicf("Cannot create '%v' command: %v", cmd.ApplicationCommandStruct().Name, err)
			}

			mutex.Lock()
			appCmds = append(appCmds, appCmd)
			mutex.Unlock()

			wg.Done()
		}(command)
	}
	wg.Wait()

	return appCmds
}

func RegisterHandler(apps []AppCmdInterface) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			for _, app := range apps {
				if app.ApplicationCommandStruct().Name == i.ApplicationCommandData().Name {
					app.ApplicationCommandHandler(s, i)
				}
			}

		case discordgo.InteractionMessageComponent:
			for _, app := range apps {
				if handler, ok := app.MessageComponentHandler()[i.MessageComponentData().CustomID]; ok {
					handler(s, i)
				}
			}

		case discordgo.InteractionModalSubmit:
			for _, app := range apps {
				if handler, ok := app.ModalSubmitHandler()[i.ModalSubmitData().CustomID]; ok {
					handler(s, i)
				}
			}
		}
	}
}
