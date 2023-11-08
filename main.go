package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/Coaltergeist/discordgo-embeds/colors"
	"github.com/Coaltergeist/discordgo-embeds/embed"
	"github.com/bwmarrin/discordgo"
)

func main() {
	s, err := discordgo.New("Bot " + os.Getenv("TOKEN"))

	if err != nil {
		log.Fatal(err)
	}

	s.AddHandler(func(s *discordgo.Session, e *discordgo.Ready) {
		fmt.Println("Ready: " + s.State.User.Username)
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if !strings.HasPrefix(m.Content, "$") {
			return
		}

		fmt.Println(m.Message.Content)

		args := strings.Split(strings.Trim(m.Content, "$"), " ")

		fmt.Println(args)

		switch args[0] {
		case "sell":
			if !strings.HasPrefix(args[1], "hangar=") {
				s.ChannelMessageSendReply(
					m.ChannelID,
					"Please specify the hangar your goods are in by specifying `hangar=hangarname` after the command.",
					m.MessageReference,
				)

				return 
			}

			err := processSell(s, m, args)

			s.ChannelMessageSendReply(
				m.ChannelID,
				"There was an error processing your command: "+err.Error(),
				m.MessageReference,
			)
		case "request":
			err := processRequest(s, m, args)

			s.ChannelMessageSendReply(
				m.ChannelID,
				"There was an error processing your command: "+err.Error(),
				m.MessageReference,
			)
		}
	})

	err = s.Open()

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

func processSell(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	hangar := strings.Split(strings.Join(strings.Split(strings.Join(args[1:], " "), "="), ""), "\n")

	em := embed.New()

	em.SetTitle(fmt.Sprintf("Sell Order (%s)", m.Member.Nick))
	em.SetDescription(
		fmt.Sprintf(
			"%s has requested the following items to be sold:\n```%s```",
			m.Member.Nick,
			strings.Join(args[1:], " "),
		),
	)
	em.AddField("Hangar", strings.TrimPrefix(hangar[0], "hangar"), false)
	em.SetColor(colors.Red())

	request, err := s.ChannelMessageSendEmbed(m.ChannelID, em.MessageEmbed)

	if err != nil {
		return err
	}

	err = s.ChannelMessageDelete(m.ChannelID, m.ID)

	if err != nil {
		return err
	}

	err = s.MessageReactionAdd(request.ChannelID, request.ID, "📥")

	if err != nil {
		return err
	}

	state := make(chan string)

	handleReaction(s, *request, state, *m.Message, "sell")

	st := <-state

	if len(st) > 0 {
		return errors.New(st)
	}

	return nil
}

func processRequest(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	em := embed.New()

	em.SetTitle(fmt.Sprintf("Item Request (%s)", m.Member.Nick))
	em.SetDescription(fmt.Sprintf("%s requested an item:\n```%s```", m.Member.Nick, strings.Join(args[1:], " ")))
	em.SetColor(colors.Green())

	request, err := s.ChannelMessageSendEmbed(m.ChannelID, em.MessageEmbed)

	if err != nil {
		return err
	}

	err = s.ChannelMessageDelete(m.ChannelID, m.ID)

	if err != nil {
		return err
	}

	err = s.MessageReactionAdd(request.ChannelID, request.ID, "📥")

	if err != nil {
		return err
	}

	state := make(chan string)

	handleReaction(s, *request, state, *m.Message, "buy")

	st := <-state

	if len(st) > 0 {
		return errors.New(st)
	}

	return nil
}

func handleReaction(s *discordgo.Session, request discordgo.Message, state chan string, m discordgo.Message, requestType string) {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if r.MessageID != request.ID || r.UserID == s.State.User.ID {
			return
		}

		err := s.ChannelMessageDelete(request.ChannelID, request.ID)

		if err != nil {
			go func() { state <- err.Error() }()

			return
		}

		dm, err := s.UserChannelCreate(r.UserID)

		if err != nil {
			go func() { state <- err.Error() }()

			return
		}

		_, err = s.ChannelMessageSend(dm.ID, fmt.Sprintf("**Your request %s has been fulfilled:**\n%s", requestType, m.Content))

		if err != nil {
			go func() { state <- err.Error() }()

			return
		}

		err = s.ChannelMessageDelete(request.ChannelID, request.ID)

		if err != nil {
			go func() { state <- err.Error() }()

			return
		}
	})
}
