package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type sendMessageRequest struct {
	Content string `json:"content" binding:"required,max=2000"`
}

// listConversations godoc
// @Summary		List Conversations
// @Description	Get a paginated list of the current user's conversations.
// @Tags			Messages
// @Produce		json
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[ConversationResponse]
// @Security		BearerAuth
// @Failure		401		{object}	ErrorResponse
// @Router			/messages/conversations [get]
func (server *Server) listConversations(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size

	items, err := server.messageUC.ListConversations(ctx, userID, page, size+1)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newConversationResponseList(items)
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}

// listConversationMessages godoc
// @Summary		List Conversation Messages
// @Description	Get a paginated list of messages in a specific conversation.
// @Tags			Messages
// @Produce		json
// @Param			id		path		int		true	"Conversation ID"
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[MessageResponse]
// @Security		BearerAuth
// @Failure		401		{object}	ErrorResponse
// @Router			/messages/conversations/{id}/messages [get]
func (server *Server) listConversationMessages(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}

	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size

	items, err := server.messageUC.ListMessages(ctx, userID, req.ID, page, size+1)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newMessageResponseList(items)
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}

// sendMessageToUser godoc
// @Summary		Send Message to User
// @Description	Start a new conversation or send a message to a user by their ID.
// @Tags			Messages
// @Accept			json
// @Produce		json
// @Param			id		path		int64					true	"Recipient User ID"
// @Param			request	body			sendMessageRequest	true	"Message content"
// @Success		201		{object}	MessageResponse
// @Security		BearerAuth
// @Failure		401		{object}	ErrorResponse
// @Router			/messages/users/{id}/messages [post]
func (server *Server) sendMessageToUser(ctx *gin.Context) {
	senderID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	var reqID idURIRequest
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		writeError(ctx, err)
		return
	}

	var req sendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, err)
		return
	}

	item, participants, err := server.messageUC.SendMessageToUser(ctx, senderID, reqID.ID, req.Content)
	if err != nil {
		writeError(ctx, err)
		return
	}

	resp := newMessageResponse(item)
	server.sendDirectMessageWS(participants, resp)
	ctx.JSON(http.StatusCreated, resp)
}

// sendMessageToConversation godoc
// @Summary		Send Message to Conversation
// @Description	Send a message to an existing conversation thread.
// @Tags			Messages
// @Accept			json
// @Produce		json
// @Param			id		path		int64					true	"Conversation ID"
// @Param			request	body			sendMessageRequest	true	"Message content"
// @Success		201		{object}	MessageResponse
// @Security		BearerAuth
// @Failure		401		{object}	ErrorResponse
// @Router			/messages/conversations/{id}/messages [post]
func (server *Server) sendMessageToConversation(ctx *gin.Context) {
	senderID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	var reqID idURIRequest
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		writeError(ctx, err)
		return
	}

	var req sendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, err)
		return
	}

	item, participants, err := server.messageUC.SendMessageToConversation(ctx, senderID, reqID.ID, req.Content)
	if err != nil {
		writeError(ctx, err)
		return
	}

	resp := newMessageResponse(item)
	server.sendDirectMessageWS(participants, resp)
	ctx.JSON(http.StatusCreated, resp)
}
