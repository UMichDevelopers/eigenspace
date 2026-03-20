package main

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const voteKickPollDurationHours = 24

type voteKickPoll struct {
	GuildID          string
	ChannelID        string
	TargetUserID     string
	InitiatorUserID  string
	AllowedVoterRole string
	Threshold        int
	YesAnswerID      int
	NoAnswerID       int
}

func (b *bot) handleVoteKickCommand(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
	if event.GuildID == "" {
		return errors.New("this command may only be used in a guild")
	}

	if len(command.Args) != 1 {
		return errors.New("usage: %vote-kick <user-id-or-mention>")
	}

	targetUserID, err := normalizeUserID(command.Args[0])
	if err != nil {
		return err
	}

	msg, err := session.ChannelMessageSendComplex(
		event.ChannelID,
		&discordgo.MessageSend{
			Reference: event.Reference(),
			Poll: &discordgo.Poll{
				Question: discordgo.PollMedia{
					Text: "Kick user " + targetUserID + "?",
				},
				Answers: []discordgo.PollAnswer{
					{Media: &discordgo.PollMedia{Text: "Yes"}},
					{Media: &discordgo.PollMedia{Text: "No"}},
				},
				AllowMultiselect: false,
				Duration:         voteKickPollDurationHours,
			},
		},
	)
	if err != nil {
		return err
	}

	slog.Info(
		"vote-kick poll created",
		"poll_message_id", msg.ID,
		"channel_id", msg.ChannelID,
		"guild_id", msg.GuildID,
		"target_user_id", targetUserID,
	)

	pollState := &voteKickPoll{
		GuildID:          event.GuildID,
		ChannelID:        event.ChannelID,
		TargetUserID:     targetUserID,
		InitiatorUserID:  event.Author.ID,
		AllowedVoterRole: strconv.FormatUint(b.cfg.VoteKick.AllowedVoterRole, 10),
		Threshold:        b.cfg.VoteKick.Threshold,
	}

	if msg.Poll != nil && len(msg.Poll.Answers) >= 2 {
		pollState.YesAnswerID = msg.Poll.Answers[0].AnswerID
		pollState.NoAnswerID = msg.Poll.Answers[1].AnswerID
	}

	b.voteKickMu.Lock()
	b.voteKickPolls[msg.ID] = pollState
	b.voteKickMu.Unlock()

	return nil
}

func normalizeUserID(s string) (string, error) {
	if strings.HasPrefix(s, "<@") && strings.HasSuffix(s, ">") {
		s = strings.TrimPrefix(s, "<@")
		s = strings.TrimSuffix(s, ">")
		s = strings.TrimPrefix(s, "!")
	}

	if s == "" {
		return "", errors.New("invalid user id")
	}

	if _, err := strconv.ParseUint(s, 10, 64); err != nil {
		return "", errors.New("invalid user id")
	}

	return s, nil
}

func (b *bot) handleMessagePollVoteAdd(session *discordgo.Session, event *discordgo.MessagePollVoteAdd) error {
	return b.handleMessagePollVote("add", session, event.GuildID, event.ChannelID, event.MessageID, event.UserID, event.AnswerID)
}

func (b *bot) handleMessagePollVoteRemove(session *discordgo.Session, event *discordgo.MessagePollVoteRemove) error {
	return b.handleMessagePollVote("remove", session, event.GuildID, event.ChannelID, event.MessageID, event.UserID, event.AnswerID)
}

func (b *bot) handleMessagePollVote(kind string, session *discordgo.Session, guildID string, channelID string, messageID string, userID string, answerID int) error {
	b.voteKickMu.Lock()
	pollState, ok := b.voteKickPolls[messageID]
	b.voteKickMu.Unlock()
	if !ok {
		return nil
	}

	member, err := session.GuildMember(guildID, userID)
	if err != nil {
		return err
	}

	voterAllowed := false
	for _, roleID := range member.Roles {
		if roleID == pollState.AllowedVoterRole {
			voterAllowed = true
			break
		}
	}

	yesVotes, err := session.PollAnswerVoters(channelID, messageID, pollState.YesAnswerID)
	if err != nil {
		return err
	}

	allowedYesVotes := 0
	for _, voter := range yesVotes {
		voterMember, err := session.GuildMember(guildID, voter.ID)
		if err != nil {
			return err
		}

		for _, roleID := range voterMember.Roles {
			if roleID == pollState.AllowedVoterRole {
				allowedYesVotes++
				break
			}
		}
	}

	slog.Info(
		"vote-kick poll vote observed",
		"kind", kind,
		"poll_message_id", messageID,
		"channel_id", channelID,
		"guild_id", guildID,
		"target_user_id", pollState.TargetUserID,
		"initiator_user_id", pollState.InitiatorUserID,
		"voter_user_id", userID,
		"voter_allowed", voterAllowed,
		"answer_id", answerID,
		"yes_answer_id", pollState.YesAnswerID,
		"allowed_yes_votes", allowedYesVotes,
		"threshold", pollState.Threshold,
	)

	return nil
}
