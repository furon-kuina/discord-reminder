package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v9"
)

type Config struct {
	discordToken string `env:"DISCORD_TOKEN,required"`
	guildId      string `env:"DISCORD_GUILD_ID,required"`
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("invalid config: %v\n", err)
		os.Exit(1)
	}
	if err := run(cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cfg Config) error {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		log.Fatal("DISCORD_TOKEN not set")
	}
	fmt.Printf("DISCORD_TOKEN: %s\n", discordToken)
	guildId := os.Getenv("DISCORD_GUILD_ID")
	if guildId == "" {
		log.Fatal("DISCORD_GUILD_ID not set")
	}
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return err
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return err
	}
	defer dg.Close()

	command := &discordgo.ApplicationCommand{
		Name:        "hello",
		Description: "Hello command",
	}
	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, guildId, command)
	if err != nil {
		log.Fatal("failed to create command: %v", err)
	}
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hello from Bot!",
			},
		})
	})

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	return nil
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("Received message: %v\n", *m.Message)

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}
