package server

import (
	"net/http"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type createTweetRequest struct {
	Content   *string `json:"content" binding:"required_without=MediaKey,omitempty,max=280"`
	ParentID  *int64  `json:"parentId" binding:"omitempty,min=1"`
	MediaKey  *string `json:"mediaKey" binding:"omitempty"`
	MediaType *string `json:"mediaType" binding:"required_with=MediaKey,omitempty,oneof=IMAGE VIDEO"`
}

func (server *Server) createTweet(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	var req createTweetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, err)
		return
	}

	// Validate media key looks like an S3 object key (folder/uuid_file)
	if req.MediaKey != nil && *req.MediaKey != "" {
		if !strings.Contains(*req.MediaKey, "/") {
			writeValidationError(ctx, "mediaKey", "invalid media key format")
			return
		}
	}

	input := usecase.CreateTweetInput{
		UserID:    userID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		MediaKey:  req.MediaKey,
		MediaType: req.MediaType,
	}

	tweet, err := server.tweetUC.CreateTweet(ctx, input)
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
	if err := server.tweetUC.DeleteTweet(ctx, userID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, successResponse())
}

func (server *Server) getTweet(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	viewerID := optionalViewerID(ctx)
	tweet, err := server.tweetUC.GetTweet(ctx, req.ID, viewerID)
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
	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size
	viewerID := optionalViewerID(ctx)
	tweets, err := server.tweetUC.ListReplies(ctx, req.ID, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newTweetResponseList(tweets)
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
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
	if err := server.tweetUC.LikeTweet(ctx, userID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, successResponse())
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
	if err := server.tweetUC.UnlikeTweet(ctx, userID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, successResponse())
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
	tweet, err := server.tweetUC.Retweet(ctx, userID, req.ID)
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
	if err := server.tweetUC.UndoRetweet(ctx, userID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, successResponse())
}
