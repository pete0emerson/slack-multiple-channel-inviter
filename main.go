package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

var userMap map[string]string
var channelMap map[string]string

func getChannelMap(api *slack.Client) (map[string]string, error) {
	log.Infof("Getting channel ID <=> Name mapping")
	channelMap := map[string]string{}
	for cursor := "init"; cursor == "init" || cursor != ""; {
		if cursor == "init" {
			cursor = ""
		}
		var channels []slack.Channel
		var err error
		params := &slack.GetConversationsParameters{
			Cursor:          cursor,
			ExcludeArchived: true,
			Types:           []string{"public_channel", "private_channel"},
		}
		channels, cursor, err = api.GetConversations(params)
		if err != nil {
			return nil, err
		}

		for _, channel := range channels {
			channelMap[channel.Name] = channel.ID
			log.Infof("Got channel '%s' (%s)", channel.Name, channel.ID)
		}
	}
	return channelMap, nil
}

func getUserMap(api *slack.Client) (map[string]string, error) {
	log.Infof("Getting user ID <=> Name mapping")
	userMap := map[string]string{}
	users, err := api.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		log.Infof("Got user '%s' (%s)", user.Name, user.ID)
		userMap[user.Name] = user.ID
	}
	return userMap, nil
}

func leaveChannel(api *slack.Client, self, channel string) {
	log.Infof("Having %s (%s) leave channel %s (%s)", self, userMap[self], channel, channelMap[channel])
	_, err := api.LeaveConversation(channelMap[channel])
	if err != nil {
		log.Fatalf("Error leaving channel: %#v", err)
	}

}

func inviteUsersToChannel(api *slack.Client, self string, channel string, users []string) error {

	params := &slack.GetUsersInConversationParameters{
		ChannelID: channelMap[channel],
	}

	usersInChannel, _, _ := api.GetUsersInConversation(params)

	// See if the bot is a member of the channel
	foundSelf := false
	for _, u := range usersInChannel {
		if u == userMap[self] {
			foundSelf = true
			log.Infof("User %s (%s) is already in the channel %s (%s)", self, userMap[self], channel, channelMap[channel])
			break
		}
	}

	// Create a list of users to invite based on if they're already in the channel or not
	var newUsers []string
	var logString string
	for _, user := range users {
		foundUser := false
		for _, u := range usersInChannel {
			if userMap[user] == u {
				log.Infof("Found user %s (%s) in channel %s (%s)", user, userMap[user], channel, channelMap[channel])
				foundUser = true
				break
			}
		}
		if !foundUser {
			log.Infof("No user %s (%s) in channel %s (%s)", user, userMap[user], channel, channelMap[channel])
			newUsers = append(newUsers, userMap[user])
			if logString == "" {
				logString = fmt.Sprintf("%s (%s)", user, userMap[user])
			} else {
				logString = logString + fmt.Sprintf(", %s (%s)", user, userMap[user])
			}
		}
	}

	if len(newUsers) == 0 {
		log.Info("No users to invite.")
		return nil
	}

	// Invite the users
	if !foundSelf {
		log.Infof("Inviting %s to join channel %s (%s)", logString, channel, channelMap[channel])
		_, _, _, err := api.JoinConversation(channelMap[channel])
		if err != nil {
			return err
		}
		defer leaveChannel(api, self, channel)
	}

	usersString := strings.Join(newUsers, ",")
	_, err := api.InviteUsersToConversation(channelMap[channel], usersString)
	if err != nil {
		return err
	}

	return nil
}

func getEnvVar(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("No %s set.", name)
	}
	return value
}

func main() {

	// Make sure our environment variables are set
	slackToken := getEnvVar("SLACK_TOKEN")
	slackChannelUsersString := getEnvVar("SLACK_CHANNEL_USERS")
	slackChannelString := getEnvVar("SLACK_CHANNELS")

	api := slack.New(
		slackToken,
	)

	// Get bot self info
	self, err := api.AuthTest()
	if err != nil {
		log.Fatalf("Error getting myself: %#v\n", err)
	}

	// Get user -> ID mapping
	userMap, err = getUserMap(api)
	if err != nil {
		log.Fatalf("Unable to get user map: %#v\n", err)
	}

	// Get channel -> ID mapping
	channelMap, err = getChannelMap(api)
	if err != nil {
		log.Fatalf("Unable to get channel map: %#v\n", err)
	}

	// Make sure each channel that we're trying to manipulate actually exists
	var slackChannels []string
	for _, channel := range strings.Split(slackChannelString, ",") {
		if _, ok := channelMap[channel]; ok {
			slackChannels = append(slackChannels, channel)
		} else {
			log.Fatalf("%s is not a valid channel.", channel)
		}
	}

	// Make sure each user we're trying to add actually exists
	var slackChannelUsers []string
	for _, user := range strings.Split(slackChannelUsersString, ",") {
		if _, ok := userMap[user]; ok {
			slackChannelUsers = append(slackChannelUsers, user)
		} else {
			log.Fatalf("%s is not a valid user.", user)
		}
	}

	for _, channel := range slackChannels {
		err := inviteUsersToChannel(api, self.User, channel, slackChannelUsers)
		if err != nil {
			log.Fatalf("Unable to invite users to channel %s (%s): %#v\n", channel, channelMap[channel], err)
		}
	}
}
