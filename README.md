# slack-multiple-channel-inviter

This golang code invites multiple people to multiple channels.

It uses the https://github.com/slack-go/slack library.

You'll need to create a Slack app first.

https://api.slack.com/apps/

* Create New App
* From scratch
* Name something like "ChannelInviter" && Pick a workspace && Create App
* Basic Information -> Add features and functionality
	* Bots
	* Review Scopes to Add
	* Scopes -> Add an OAuth Scope

```
channels:join
channels:manage
channels:read
groups:read
groups:write
users:read
```

	* scroll up to OAuth Tokens for Your Workspace
		* Install to Workspace
		* Capture the "Bot User OAuth Token" for use with this code




```
channels:join
channels:manage
channels:read
groups:read
groups:write
users:read
```

Capture the `Bot User OAuth Token` from https://api.slack.com/apps/ (choose your app, click on OAuth & Permissions, Bot User OAuth Token).

Prior to running this script, export the following environment variables:

```
# Set the verbosity of the output if you want to
export INVITER_VERBOSE=true

# The Bot User OAuth Toekn
export SLACK_TOKEN=

# Comma separated user names to invite
export SLACK_CHANNEL_USERS=

# Comma separated channel names to invite the SLACK_CHANNEL_USER to
export SLACK_CHANNELS=
```

Example:

```
go build
export SLACK_TOKEN=xoxb-...
export SLACK_CHANNEL_USERS=alice,bob,charlie,dawn
export SLACK_CHANNELS=random,memes,sales

# Show what will happen, but don't actually do it
./slack-channel-inviter --dry-run

# Make it happen!
./slack-channel-inviter
```
