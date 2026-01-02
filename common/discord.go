package common

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func DiscordSendMessage(s *discordgo.Session, serverID, channelID, message, recipientType, recipientID string) error {
	// Get the channel to send the message
	channel, err := s.Channel(channelID)
	if err != nil {
		LogError("DISCORD_CHANNEL_ERROR", map[string]interface{}{"error": err.Error(), "channel_id": channelID}, "")
		return err
	}

	// Create a message content based on the recipient type
	var content string
	switch recipientType {
	case "everyone":
		content = message
	case "role":
		role, _ := getRoleByID(s, serverID, recipientID)
		content = fmt.Sprintf("<@&%s> %s", recipientID, message)
		if role != nil {
			Log("Discord message to role", map[string]interface{}{"role": role.Name, "channel": channel.Name}, "")
		}
	case "user":
		user, _ := getUserByID(s, recipientID)
		content = fmt.Sprintf("<@%s> %s", recipientID, message)
		if user != nil {
			Log("Discord message to user", map[string]interface{}{"user": user.Username, "channel": channel.Name}, "")
		}
	default:
		return errors.New("invalid recipient type: " + recipientType)
	}

	// Send the message to the channel
	_, err = s.ChannelMessageSend(channel.ID, content)
	if err != nil {
		LogError("DISCORD_SEND_ERROR", map[string]interface{}{"error": err.Error(), "channel": channel.Name}, "")
		return err
	}

	return nil
}

func DiscordCheckOnlineStatus(status string) string {
	strList := []string{string(discordgo.StatusOnline), string(discordgo.StatusDoNotDisturb), string(discordgo.StatusIdle)}

	if StringExistsInList(string(status), strList) {
		return "online"
	}
	return "offline"
}

func DiscordUsername(s *discordgo.Session, userID string) string {
	user, err := s.User(userID)
	if err != nil {
		return ""
	}
	return user.Username
}

func DiscordChannelName(s *discordgo.Session, channelID string) string {
	channel, err := s.Channel(channelID)
	if err != nil {
		return ""
	}
	return channel.Name
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
