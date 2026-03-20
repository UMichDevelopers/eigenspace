package main

import (
	"errors"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const voteKickPollDurationHours = 24
const voteKickQuestionPrefix = "[vote-kick target="
const voteKickQuestionSuffix = "] "

func (b *bot) handleVoteKickCommand(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
	if event.GuildID == "" {
		return errors.New("this command may only be used in a guild")
	}

	if len(command.Args) < 1 || len(command.Args) > 2 {
		return errors.New("usage: %vote-kick <user-id-or-mention> [:reason]")
	}

	targetUserID, err := normalizeUserID(command.Args[0])
	if err != nil {
		return err
	}

	reason := ""
	if len(command.Args) == 2 {
		reason = strings.TrimSpace(command.Args[1])
	}

	msg, err := session.ChannelMessageSendComplex(
		event.ChannelID,
		&discordgo.MessageSend{
			Reference: event.Reference(),
			Poll: &discordgo.Poll{
				Question: discordgo.PollMedia{
					Text: voteKickQuestion(targetUserID, reason),
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

	noAnswerID, ok := voteKickNoAnswerID(msg)
	if !ok {
		return errors.New("vote-kick poll does not have a No answer")
	}

	member, err := session.GuildMember(guildID, userID)
	if err != nil {
		return err
	}

	allowedVoterRoleID := strconv.FormatUint(b.cfg.VoteKick.AllowedVoterRole, 10)
	voterAllowed := slices.Contains(member.Roles, allowedVoterRoleID)

	yesVotes, err := session.PollAnswerVoters(channelID, messageID, yesAnswerID)
	if err != nil {
		return err
	}

	noVotes, err := session.PollAnswerVoters(channelID, messageID, noAnswerID)
	if err != nil {
		return err
	}

	allowedYesVotes := 0
	for _, voter := range yesVotes {
		voterMember, err := session.GuildMember(guildID, voter.ID)
		if err != nil {
			return err
		}

		if slices.Contains(voterMember.Roles, allowedVoterRoleID) {
			allowedYesVotes++
		}
	}

	allowedNoVotes := 0
	for _, voter := range noVotes {
		voterMember, err := session.GuildMember(guildID, voter.ID)
		if err != nil {
			return err
		}

		if slices.Contains(voterMember.Roles, allowedVoterRoleID) {
			allowedNoVotes++
		}
	}

	score := allowedYesVotes - allowedNoVotes

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
		"no_answer_id", noAnswerID,
		"allowed_yes_votes", allowedYesVotes,
		"allowed_no_votes", allowedNoVotes,
		"score", score,
		"threshold", b.cfg.VoteKick.Threshold,
	)

	if score < b.cfg.VoteKick.Threshold {
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
		"allowed_no_votes", allowedNoVotes,
		"score", score,
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

	rest := strings.TrimPrefix(text, voteKickQuestionPrefix)
	targetUserID, _, ok := strings.Cut(rest, voteKickQuestionSuffix)
	if !ok || targetUserID == "" {
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

func voteKickNoAnswerID(msg *discordgo.Message) (int, bool) {
	if msg.Poll == nil {
		return 0, false
	}

	for _, answer := range msg.Poll.Answers {
		if answer.Media != nil && answer.Media.Text == "No" {
			return answer.AnswerID, true
		}
	}

	return 0, false
}

func voteKickQuestion(targetUserID string, reason string) string {
	text := voteKickQuestionPrefix + targetUserID + voteKickQuestionSuffix + "Kick <@" + targetUserID + ">?"
	if reason != "" {
		text += " Reason: " + reason
	}

	return text
}
