package server

import (
	"mime/multipart"
	"net/http"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type createTweetRequest struct {
	Content  *string               `form:"content" binding:"required_without=Media,omitempty,max=280"`
	ParentID *int64                `form:"parentId" binding:"omitempty,min=1"`
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

	input := usecase.CreateTweetInput{UserID: userID, Content: req.Content, ParentID: req.ParentID}

	if req.Media != nil {
		if server.config.MaxMediaBytes > 0 && req.Media.Size > server.config.MaxMediaBytes {
			writeValidationError(ctx, "media", "file size exceeds limit")
			return
		}

		if !hasAllowedExtension(req.Media.Filename, mediaAllowedExts) {
			writeValidationError(ctx, "media", "unsupported file extension")
			return
		}

		file, reader, detectedContentType, err := openAndDetectUpload(req.Media)
		if err != nil {
			writeError(ctx, apperr.BadRequest("failed to inspect media file"))
			return
		}
		defer file.Close()

		if !isAllowedType(detectedContentType, mediaAllowedMIMEs) {
			writeValidationError(ctx, "media", "unsupported media type")
			return
		}

		input.Media = &usecase.MediaUpload{
			Filename:    req.Media.Filename,
			ContentType: detectedContentType,
			Reader:      reader,
		}
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
	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}
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
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}
	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}
	tweets, err := server.tweetUC.ListReplies(ctx, req.ID, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.tweetUC.CountReplies(ctx, req.ID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newTweetResponseList(tweets)
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
