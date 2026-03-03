package server

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type createTweetRequest struct {
	Content  *string               `form:"content" binding:"required_without=Media,omitempty,max=280"`
	ParentID *string               `form:"parentId" binding:"omitempty,numeric"`
	Media    *multipart.FileHeader `form:"media"`
}

func (server *Server) createTweet(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	var req createTweetRequest
	if err := ctx.ShouldBind(&req); err != nil {
		writeError(ctx, err)
		return
	}

	var parentID *int64
	if req.ParentID != nil {
		id, err := strconv.ParseInt(*req.ParentID, 10, 64)
		if err != nil || id <= 0 {
			writeValidationError(ctx, "ParentID", "must be a positive number")
			return
		}
		parentID = &id
	}

	input := usecase.CreateTweetInput{UserID: userID, Content: req.Content, ParentID: parentID}

	if req.Media != nil {
		file, err := req.Media.Open()
		if err != nil {
			writeError(ctx, apperr.BadRequest("failed to open media file"))
			return
		}
		defer file.Close()

		input.Media = &usecase.MediaUpload{
			Filename:    req.Media.Filename,
			ContentType: req.Media.Header.Get("Content-Type"),
			Reader:      file,
		}
	}

	tweet, err := server.usecase.CreateTweet(ctx, input)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, newTweetResponse(tweet))
}

func (server *Server) deleteTweet(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	if err := server.usecase.DeleteTweet(ctx, userID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (server *Server) getTweet(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}
	tweet, err := server.usecase.GetTweet(ctx, req.ID, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, newTweetResponse(tweet))
}

func (server *Server) getReplies(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}
	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}
	tweets, err := server.usecase.ListReplies(ctx, req.ID, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountReplies(ctx, req.ID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := make([]tweetResponse, 0, len(tweets))
	for _, t := range tweets {
		response = append(response, newTweetResponse(t))
	}
	ctx.JSON(http.StatusOK, buildPageResponse(response, page, size, total))
}

func (server *Server) likeTweet(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	if err := server.usecase.LikeTweet(ctx, userID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (server *Server) unlikeTweet(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	if err := server.usecase.UnlikeTweet(ctx, userID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (server *Server) retweet(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	tweet, err := server.usecase.Retweet(ctx, userID, req.ID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, newTweetResponse(tweet))
}

func (server *Server) undoRetweet(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	if err := server.usecase.UndoRetweet(ctx, userID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
