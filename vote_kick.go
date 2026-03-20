package main

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const voteKickPollDurationHours = 24
const voteKickQuestionPrefix = "vote-kick target="

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
					Text: voteKickQuestionPrefix + targetUserID,
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
		"guild_id", event.GuildID,
		"target_user_id", targetUserID,
	)

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
	msg, err := session.ChannelMessage(channelID, messageID)
	if err != nil {
		return err
	}

	targetUserID, ok := voteKickTargetUserID(msg)
	if !ok {
		return nil
	}

	yesAnswerID, ok := voteKickYesAnswerID(msg)
	if !ok {
		return errors.New("vote-kick poll does not have a Yes answer")
	}

	member, err := session.GuildMember(guildID, userID)
	if err != nil {
		return err
	}

	allowedVoterRoleID := strconv.FormatUint(b.cfg.VoteKick.AllowedVoterRole, 10)
	voterAllowed := hasRole(member.Roles, allowedVoterRoleID)

	yesVotes, err := session.PollAnswerVoters(channelID, messageID, yesAnswerID)
	if err != nil {
		return err
	}

	allowedYesVotes := 0
	for _, voter := range yesVotes {
		voterMember, err := session.GuildMember(guildID, voter.ID)
		if err != nil {
			return err
		}

		if hasRole(voterMember.Roles, allowedVoterRoleID) {
			allowedYesVotes++
		}
	}

	slog.Info(
		"vote-kick poll vote observed",
		"kind", kind,
		"poll_message_id", messageID,
		"channel_id", channelID,
		"guild_id", guildID,
		"target_user_id", targetUserID,
		"voter_user_id", userID,
		"voter_allowed", voterAllowed,
		"answer_id", answerID,
		"yes_answer_id", yesAnswerID,
		"allowed_yes_votes", allowedYesVotes,
		"threshold", b.cfg.VoteKick.Threshold,
	)

	if allowedYesVotes < b.cfg.VoteKick.Threshold {
		return nil
	}

	_, err = session.PollExpire(channelID, messageID)
	if err != nil {
		return err
	}

	err = session.GuildMemberDeleteWithReason(guildID, targetUserID, "vote-kick threshold reached")
	if err != nil {
		return err
	}

	slog.Info(
		"vote-kick executed",
		"poll_message_id", messageID,
		"channel_id", channelID,
		"guild_id", guildID,
		"target_user_id", targetUserID,
		"allowed_yes_votes", allowedYesVotes,
		"threshold", b.cfg.VoteKick.Threshold,
	)

	return nil
}

func voteKickTargetUserID(msg *discordgo.Message) (string, bool) {
	if msg.Poll == nil {
		return "", false
	}

	text := msg.Poll.Question.Text
	if !strings.HasPrefix(text, voteKickQuestionPrefix) {
		return "", false
	}

	targetUserID := strings.TrimPrefix(text, voteKickQuestionPrefix)
	if targetUserID == "" {
		return "", false
	}

	return targetUserID, true
}

func voteKickYesAnswerID(msg *discordgo.Message) (int, bool) {
	if msg.Poll == nil {
		return 0, false
	}

	for _, answer := range msg.Poll.Answers {
		if answer.Media != nil && answer.Media.Text == "Yes" {
			return answer.AnswerID, true
		}
	}

	return 0, false
}

func hasRole(roleIDs []string, wantedRoleID string) bool {
	for _, roleID := range roleIDs {
		if roleID == wantedRoleID {
			return true
		}
	}

	return false
}
