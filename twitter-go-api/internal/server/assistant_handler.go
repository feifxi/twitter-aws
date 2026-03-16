package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

func (server *Server) assistant(ctx *gin.Context) {
	var input usecase.AssistantInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		writeError(ctx, err)
		return
	}

	reader, err := server.assistantUC.Chat(ctx.Request.Context(), input)
	if err != nil {
		writeError(ctx, err)
		return
	}

	// Set SSE headers before writing status code
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no") // Disable nginx/proxy buffering
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Writer.Flush()

	buffer := make([]byte, 1024)
	for {
		n, readErr := reader.Read(buffer)
		if n > 0 {
			// Write proper SSE frame: "data: <chunk>\n\n"
			chunk := string(buffer[:n])
			_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", chunk)
			ctx.Writer.Flush()
		}
		if readErr != nil {
			if readErr != io.EOF {
				// Send error as SSE event before closing
				_, _ = fmt.Fprintf(ctx.Writer, "event: error\ndata: %s\n\n", readErr.Error())
				ctx.Writer.Flush()
			}
			break
		}
	}

	// Gracefully close — no trailing writes. This ensures API Gateway
	// doesn't hit the 29-second idle timeout on a still-open connection.
}
