package common

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func DiscordSendMessage(s *discordgo.Session, serverID, channelID, message, recipientType, recipientID string) error {
	// Get the channel to send the message
	channel, err := s.Channel(channelID)
	if err != nil {
		fmt.Println("Error retrieving channel: ", err)
		return err
	}

	// Create a message content based on the recipient type
	var to string
	var content string
	switch recipientType {
	case "everyone":
		content = message
	case "role":
		role, _ := getRoleByID(s, serverID, recipientID)
		to = fmt.Sprintf("@%s ", role.Name)
		content = fmt.Sprintf("<@&%s> %s", recipientID, message)
	case "user":
		user, _ := getUserByID(s, recipientID)
		to = fmt.Sprintf("@%s ", user.Username)
		content = fmt.Sprintf("<@%s> %s", recipientID, message)
	default:
		return fmt.Errorf("invalid recipient type: %s", recipientType)
	}

	// Send the message to the channel
	_, err = s.ChannelMessageSend(channel.ID, content)
	if err != nil {
		fmt.Println("Error sending message: ", err)
		return err
	}

	fmt.Printf("Discord => [%s] %s%s\n", channel.Name, to, message)
	return nil
}

func getUserByID(s *discordgo.Session, userID string) (*discordgo.User, error) {
	user, err := s.User(userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}
	return user, nil
}

func getRoleByID(s *discordgo.Session, serverID, roleID string) (*discordgo.Role, error) {
	// Retrieve the guild from the session's state
	guild, err := s.State.Guild(serverID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving guild: %v", err)
	}

	for _, role := range guild.Roles {
		if role.ID == roleID {
			return role, nil
		}
	}

	return nil, fmt.Errorf("role not found: %s", roleID)
}
