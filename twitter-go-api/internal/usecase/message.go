package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/jackc/pgx/v5"
)

var roomKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

func normalizeMessageContent(content string) (string, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "", apperr.BadRequest("message content is required")
	}
	if len(trimmed) > 2000 {
		return "", apperr.BadRequest("message content exceeds 2000 characters")
	}
	return trimmed, nil
}

func normalizeRoomKey(roomKey string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(roomKey))
	if normalized == "" {
		normalized = "global"
	}
	if !roomKeyPattern.MatchString(normalized) {
		return "", apperr.BadRequest("invalid room key")
	}
	return normalized, nil
}

func (u *MessageUsecase) ListConversations(ctx context.Context, userID int64, page, size int32) ([]ConversationItem, error) {
	rows, err := u.store.ListUserConversations(ctx, db.ListUserConversationsParams{
		UserID: userID,
		Limit:  size,
		Offset: page * size,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []ConversationItem{}, nil
	}

	idSet := make(map[int64]struct{}, len(rows)*2)
	for _, r := range rows {
		idSet[r.User.ID] = struct{}{}
		idSet[r.LastSenderID] = struct{}{}
	}

	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	viewerID := userID
	users, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
		ViewerID: &viewerID,
		UserIds:  ids,
	})
	if err != nil {
		return nil, err
	}

	userMap := make(map[int64]UserItem, len(users))
	for _, row := range users {
		userMap[row.User.ID] = newUserItemFromDB(row.User, row.IsFollowing)
	}

	items := make([]ConversationItem, 0, len(rows))
	for _, r := range rows {
		peer, ok := userMap[r.User.ID]
		if !ok {
			continue
		}
		lastSender, ok := userMap[r.LastSenderID]
		if !ok {
			continue
		}

		items = append(items, ConversationItem{
			ID:        r.ConversationID,
			Peer:      peer,
			UpdatedAt: r.ConversationUpdatedAt,
			LastMessage: MessageItem{
				ID:             r.LastMessageID,
				ConversationID: r.ConversationID,
				Sender:         lastSender,
				Content:        r.LastMessageContent,
				CreatedAt:      r.LastMessageCreatedAt,
			},
		})
	}

	return items, nil
}

func (u *MessageUsecase) ListMessages(ctx context.Context, userID, conversationID int64, page, size int32) ([]MessageItem, error) {
	allowed, err := u.store.IsConversationParticipant(ctx, db.IsConversationParticipantParams{
		ConversationID: conversationID,
		UserID:         userID,
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, apperr.Forbidden("not allowed to access this conversation")
	}

	rows, err := u.store.ListConversationMessages(ctx, db.ListConversationMessagesParams{
		ConversationID: conversationID,
		Limit:          size,
		Offset:         page * size,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []MessageItem{}, nil
	}

	ids := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.SenderID]; ok {
			continue
		}
		seen[row.SenderID] = struct{}{}
		ids = append(ids, row.SenderID)
	}

	viewerID := userID
	users, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
		ViewerID: &viewerID,
		UserIds:  ids,
	})
	if err != nil {
		return nil, err
	}

	userMap := make(map[int64]UserItem, len(users))
	for _, row := range users {
		userMap[row.User.ID] = newUserItemFromDB(row.User, row.IsFollowing)
	}

	items := make([]MessageItem, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		row := rows[i] // rows are DESC in SQL; reverse to ASC for chat rendering.
		sender, ok := userMap[row.SenderID]
		if !ok {
			continue
		}

		items = append(items, MessageItem{
			ID:             row.ID,
			ConversationID: row.ConversationID,
			Sender:         sender,
			Content:        row.Content,
			CreatedAt:      row.CreatedAt,
		})
	}

	return items, nil
}

func (u *MessageUsecase) SendMessageToUser(ctx context.Context, senderID, recipientID int64, content string) (MessageItem, []int64, error) {
	if senderID == recipientID {
		return MessageItem{}, nil, apperr.BadRequest("cannot message yourself")
	}

	normalized, err := normalizeMessageContent(content)
	if err != nil {
		return MessageItem{}, nil, err
	}

	if _, err := u.store.GetUser(ctx, db.GetUserParams{ID: recipientID, ViewerID: nil}); err != nil {
		return MessageItem{}, nil, err
	}

	var created db.DirectMessage
	var conversationID int64
	err = u.store.ExecTx(ctx, func(q *db.Queries) error {
		conversation, findErr := q.FindDirectConversation(ctx, db.FindDirectConversationParams{
			UserID:   senderID,
			UserID_2: recipientID,
		})
		if findErr != nil {
			if !errors.Is(findErr, pgx.ErrNoRows) {
				return findErr
			}

			newConversation, createErr := q.CreateConversation(ctx)
			if createErr != nil {
				return createErr
			}
			conversationID = newConversation.ID

			if addErr := q.AddConversationParticipant(ctx, db.AddConversationParticipantParams{
				ConversationID: conversationID,
				UserID:         senderID,
			}); addErr != nil {
				return addErr
			}
			if addErr := q.AddConversationParticipant(ctx, db.AddConversationParticipantParams{
				ConversationID: conversationID,
				UserID:         recipientID,
			}); addErr != nil {
				return addErr
			}
		} else {
			conversationID = conversation.ID
		}

		message, createErr := q.CreateDirectMessage(ctx, db.CreateDirectMessageParams{
			ConversationID: conversationID,
			SenderID:       senderID,
			Content:        normalized,
		})
		if createErr != nil {
			return createErr
		}
		created = message

		return q.TouchConversation(ctx, conversationID)
	})
	if err != nil {
		return MessageItem{}, nil, err
	}

	participants, err := u.store.ListConversationParticipantIDs(ctx, conversationID)
	if err != nil {
		return MessageItem{}, nil, err
	}

	viewerID := senderID
	senderRows, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
		ViewerID: &viewerID,
		UserIds:  []int64{senderID},
	})
	if err != nil {
		return MessageItem{}, nil, err
	}
	if len(senderRows) == 0 {
		return MessageItem{}, nil, apperr.Internal("sender not found", nil)
	}
	sender := newUserItemFromDB(senderRows[0].User, senderRows[0].IsFollowing)

	return MessageItem{
		ID:             created.ID,
		ConversationID: created.ConversationID,
		Sender:         sender,
		Content:        created.Content,
		CreatedAt:      created.CreatedAt,
	}, participants, nil
}

func (u *MessageUsecase) SendMessageToConversation(ctx context.Context, senderID, conversationID int64, content string) (MessageItem, []int64, error) {
	normalized, err := normalizeMessageContent(content)
	if err != nil {
		return MessageItem{}, nil, err
	}

	allowed, err := u.store.IsConversationParticipant(ctx, db.IsConversationParticipantParams{
		ConversationID: conversationID,
		UserID:         senderID,
	})
	if err != nil {
		return MessageItem{}, nil, err
	}
	if !allowed {
		return MessageItem{}, nil, apperr.Forbidden("not allowed to send messages to this conversation")
	}

	var created db.DirectMessage
	err = u.store.ExecTx(ctx, func(q *db.Queries) error {
		var createErr error
		created, createErr = q.CreateDirectMessage(ctx, db.CreateDirectMessageParams{
			ConversationID: conversationID,
			SenderID:       senderID,
			Content:        normalized,
		})
		if createErr != nil {
			return createErr
		}
		return q.TouchConversation(ctx, conversationID)
	})
	if err != nil {
		return MessageItem{}, nil, err
	}

	participants, err := u.store.ListConversationParticipantIDs(ctx, conversationID)
	if err != nil {
		return MessageItem{}, nil, err
	}

	viewerID := senderID
	senderRows, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
		ViewerID: &viewerID,
		UserIds:  []int64{senderID},
	})
	if err != nil {
		return MessageItem{}, nil, err
	}
	if len(senderRows) == 0 {
		return MessageItem{}, nil, apperr.Internal("sender not found", nil)
	}
	sender := newUserItemFromDB(senderRows[0].User, senderRows[0].IsFollowing)

	return MessageItem{
		ID:             created.ID,
		ConversationID: created.ConversationID,
		Sender:         sender,
		Content:        created.Content,
		CreatedAt:      created.CreatedAt,
	}, participants, nil
}

func (u *MessageUsecase) ListPublicRoomMessages(ctx context.Context, roomKey string, page, size int32, viewerID *int64) ([]PublicRoomMessageItem, error) {
	normalizedRoomKey, err := normalizeRoomKey(roomKey)
	if err != nil {
		return nil, err
	}

	rows, err := u.store.ListPublicRoomMessages(ctx, db.ListPublicRoomMessagesParams{
		RoomKey: normalizedRoomKey,
		Limit:   size,
		Offset:  page * size,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []PublicRoomMessageItem{}, nil
	}

	ids := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.SenderID]; ok {
			continue
		}
		seen[row.SenderID] = struct{}{}
		ids = append(ids, row.SenderID)
	}

	users, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
		ViewerID: viewerID,
		UserIds:  ids,
	})
	if err != nil {
		return nil, err
	}

	userMap := make(map[int64]UserItem, len(users))
	for _, row := range users {
		userMap[row.User.ID] = newUserItemFromDB(row.User, row.IsFollowing)
	}

	items := make([]PublicRoomMessageItem, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		row := rows[i]
		sender, ok := userMap[row.SenderID]
		if !ok {
			continue
		}

		items = append(items, PublicRoomMessageItem{
			ID:        row.ID,
			RoomKey:   row.RoomKey,
			Sender:    sender,
			Content:   row.Content,
			CreatedAt: row.CreatedAt,
		})
	}
	return items, nil
}

func (u *MessageUsecase) SendPublicRoomMessage(ctx context.Context, senderID int64, roomKey, content string) (PublicRoomMessageItem, error) {
	normalizedRoomKey, err := normalizeRoomKey(roomKey)
	if err != nil {
		return PublicRoomMessageItem{}, err
	}

	normalizedContent, err := normalizeMessageContent(content)
	if err != nil {
		return PublicRoomMessageItem{}, err
	}

	created, err := u.store.CreatePublicRoomMessage(ctx, db.CreatePublicRoomMessageParams{
		RoomKey:  normalizedRoomKey,
		SenderID: senderID,
		Content:  normalizedContent,
	})
	if err != nil {
		return PublicRoomMessageItem{}, err
	}

	viewerID := senderID
	senderRows, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
		ViewerID: &viewerID,
		UserIds:  []int64{senderID},
	})
	if err != nil {
		return PublicRoomMessageItem{}, err
	}
	if len(senderRows) == 0 {
		return PublicRoomMessageItem{}, apperr.Internal("sender not found", nil)
	}
	sender := newUserItemFromDB(senderRows[0].User, senderRows[0].IsFollowing)

	return PublicRoomMessageItem{
		ID:        created.ID,
		RoomKey:   created.RoomKey,
		Sender:    sender,
		Content:   created.Content,
		CreatedAt: created.CreatedAt,
	}, nil
}
