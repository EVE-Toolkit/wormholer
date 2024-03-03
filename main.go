package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Coaltergeist/discordgo-embeds/colors"
	"github.com/Coaltergeist/discordgo-embeds/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Signature struct {
	ID   string
	Type string
	Name string
}

func (s *Signature) String() string {
	return fmt.Sprintf("%s (%s) - %s", s.ID, s.Type, s.Name)
}

func main() {
	if strings.ToLower(os.Getenv("ENV")) != "prod" {
		err := godotenv.Load("./.env")

		if err != nil {
			log.Fatal(err)
		}
	}

	s, err := discordgo.New("Bot " + os.Getenv("TOKEN"))

	if err != nil {
		log.Fatal(err)
	}

	s.AddHandler(func(s *discordgo.Session, e *discordgo.Ready) {
		fmt.Println("Ready: " + s.State.User.Username)
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		fmt.Println(m.Content)
		
		if !strings.HasPrefix(m.Content, "$") {
			return
		}

		args := strings.Split(strings.Trim(m.Content, "$"), " ")

		fmt.Println(args)

		switch args[0][:strings.Index(stringified, "=")] {
		case "system":
			err := processScan(s, m)

			if err != nil {
				s.ChannelMessageSendReply(
					m.ChannelID,
					"There was an error processing your scan.",
					m.MessageReference,
				)

				return
			}

			return
		case "sell":
			if !strings.HasPrefix(args[1], "hangar=") {
				s.ChannelMessageSendReply(
					m.ChannelID,
					"Please specify the hangar your goods are in by specifying `hangar=hangarname` after the command.",
					m.MessageReference,
				)

				return
			}

			if !strings.Contains(m.Content, "items=") {
				s.ChannelMessageSendReply(
					m.ChannelID,
					"Please specify the items you are selling with `items=your items here` after the hangar option.",
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

func processScan(s *discordgo.Session, m *discordgo.MessageCreate) error {
	lines := strings.Split(m.Content, "\n")

	fmt.Println(lines[0])

	if !strings.HasPrefix(lines[0], "system=") {
		return nil
	}

	system := strings.Split(lines[0], "=")[1]

	var scanned []string
	var unscanned []string

	for i, line := range lines[1:] {
		if line == "" || line == "\n" {
			lines = append(lines[:i], lines[i+1:]...)

			continue
		}

		fields := strings.Split(line, "  ")

		for j, field := range fields {
			if field == "" || field == " " {
				fields = append(fields[:j], fields[j+1:]...)

				continue
			}
		}

		signature := Signature{}

		if len(fields) == 6 && len(strings.Trim(fields[2], " ")) > 0 && len(strings.Trim(fields[3], " ")) > 0 {
			signature.ID = fields[0]
			signature.Type = fields[2]
			signature.Name = fields[3]

			scanned = append(scanned, signature.String())
		} else if len(strings.Trim(fields[2], " ")) == 0 || len(strings.Trim(fields[3], " ")) == 0 {
			signature.ID = fields[0]
			signature.Type = "Unknown"
			signature.Name = "Unknown"

			unscanned = append(unscanned, signature.String())
		}
	}

	em := embed.New()

	em.SetTitle(fmt.Sprintf("New Scan Report - %s (%s)", system, m.Member.Nick))
	em.SetTimestamp(time.Now())
	em.SetDescription(fmt.Sprintf("New Scan Report for system **%s** at %s", system, em.Timestamp))
	em.AddField("Scanned", fmt.Sprintf("```\n%s```", strings.Join(scanned, "\n")), false)
	em.AddField("Unscanned", fmt.Sprintf("```\n%s```", strings.Join(unscanned, "\n")), false)
	em.SetColor(colors.Blue())

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, em.MessageEmbed)

	if err != nil {
		return err
	}

	err = s.ChannelMessageDelete(m.ChannelID, m.ID)

	if err != nil {
		return err
	}

	return nil
}

func processSell(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	stringified := strings.Join(args, " ")

	hangar := stringified[strings.Index(stringified, "hangar="):strings.Index(stringified, "items=")] 
	items := stringified[strings.Index(stringified, "items="):]

	fmt.Printf("hangar: %s\n", hangar)

	em := embed.New()

	em.SetTitle(fmt.Sprintf("Sell Order (%s)", m.Member.Nick))
	em.SetDescription(
		fmt.Sprintf(
			"%s has requested the following items to be sold:\n```%s```",
			m.Member.Nick,
			strings.TrimPrefix(items, "items="),
		),
	)
	em.AddField("Hangar", strings.TrimPrefix(hangar, "hangar="), false)
	em.SetColorRGB(237, 40, 122)

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
