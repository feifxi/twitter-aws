package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type sendMessageRequest struct {
	Content string `json:"content" binding:"required,max=2000"`
}

type roomURIRequest struct {
	Room string `uri:"room" binding:"required,max=64"`
}

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
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}

func (server *Server) listConversationMessages(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
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
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}

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

func (server *Server) listPublicRoomMessages(ctx *gin.Context) {
	var req roomURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}

	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size

	viewerID := optionalViewerID(ctx)

	items, err := server.messageUC.ListPublicRoomMessages(ctx, req.Room, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newPublicMessageResponseList(items)
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}

func (server *Server) sendPublicRoomMessage(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	var roomReq roomURIRequest
	if err := ctx.ShouldBindUri(&roomReq); err != nil {
		writeError(ctx, err)
		return
	}

	var req sendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, err)
		return
	}

	item, err := server.messageUC.SendPublicRoomMessage(ctx, userID, roomReq.Room, req.Content)
	if err != nil {
		writeError(ctx, err)
		return
	}

	resp := newPublicMessageResponse(item)
	server.sendPublicRoomWS(resp)
	ctx.JSON(http.StatusCreated, resp)
}
