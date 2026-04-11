package kit

import (
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
)

func DiscordSendMessage(s *discordgo.Session, serverID, channelID, message, recipientType, recipientID string) error {
	channel, err := s.Channel(channelID)
	if err != nil {
		return err
	}

	var content string
	switch recipientType {
	case "everyone":
		content = message
	case "role":
		content = fmt.Sprintf("<@&%s> %s", recipientID, message)
		if _, err := getRoleByID(s, serverID, recipientID); err != nil {
			return err
		}
	case "user":
		content = fmt.Sprintf("<@%s> %s", recipientID, message)
		if _, err := s.User(recipientID); err != nil {
			return fmt.Errorf("retrieving user: %w", err)
		}
	default:
		return fmt.Errorf("invalid recipient type: %s", recipientType)
	}

	_, err = s.ChannelMessageSend(channel.ID, content)
	return err
}

func DiscordCheckOnlineStatus(status string) string {
	if slices.Contains([]string{
		string(discordgo.StatusOnline),
		string(discordgo.StatusDoNotDisturb),
		string(discordgo.StatusIdle),
	}, status) {
		return "online"
	}
	return "offline"
}

func DiscordUsername(s *discordgo.Session, userID string) (string, error) {
	user, err := s.User(userID)
	if err != nil {
		return "", fmt.Errorf("retrieving user: %w", err)
	}
	return user.Username, nil
}

func DiscordChannelName(s *discordgo.Session, channelID string) (string, error) {
	channel, err := s.Channel(channelID)
	if err != nil {
		return "", fmt.Errorf("retrieving channel: %w", err)
	}
	return channel.Name, nil
}

func getRoleByID(s *discordgo.Session, serverID, roleID string) (*discordgo.Role, error) {
	guild, err := s.State.Guild(serverID)
	if err != nil {
		return nil, fmt.Errorf("retrieving guild: %w", err)
	}

	for _, role := range guild.Roles {
		if role.ID == roleID {
			return role, nil
		}
	}

	return nil, fmt.Errorf("role not found: %s", roleID)
}
