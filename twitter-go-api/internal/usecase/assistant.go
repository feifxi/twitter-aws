package usecase

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/google/generative-ai-go/genai"
	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func (u *AssistantUsecase) Chat(ctx context.Context, input AssistantInput) (io.Reader, error) {
	if u.config.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(u.config.GeminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	var timelineContext string
	if !u.config.EnableRAG {
		log.Debug().Msg("RAG is explicitly disabled in config")
		timelineContext = "Timeline context is currently disabled."
	} else {
		// 1. Vectorize the query
		emModel := client.EmbeddingModel(u.config.GeminiEmbeddingModel)
		res, err := emModel.EmbedContent(ctx, genai.Text(input.Query))
		if err != nil {
			log.Error().Err(err).Msg("failed to vectorize query for RAG, falling back to base chat")
			timelineContext = "Timeline context unavailable (vectorization failed)."
		} else {
			queryVector := pgvector.NewVector(res.Embedding.Values)

			// 2. Retrieve Context (RAG)
			rows, err := u.store.ListRelatedTweetsByEmbedding(ctx, db.ListRelatedTweetsByEmbeddingParams{
				Column1: &queryVector,
				Limit:   5,
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to retrieve context from DB for RAG, falling back to base chat")
				timelineContext = "Timeline context unavailable (DB error)."
			} else if len(rows) == 0 {
				timelineContext = "No recent matching posts found in the timeline."
			} else {
				var contextParts []string
				for _, row := range rows {
					contextParts = append(contextParts, fmt.Sprintf("- %s", row.Content))
				}
				timelineContext = strings.Join(contextParts, "\n")
			}
		}
	}

	// 3. Construct System Prompt
	systemPrompt := `You are the helpful, professional, and polite AI assistant for "Chanom Twitter", a social media platform and Twitter clone. 
	Your goal is to provide accurate, concise, and friendly information to users of Chanom Twitter. 
	You have access to the user's recent [Chat History] and [Timeline Context] from their feed. 
	Prioritize the Timeline Context when answering questions about the app's content or the user's recent experience. 
	Maintain conversational continuity and a helpful tone consistent with a premium social experience.
	CRITICAL: Your memory is limited to the provided history. If asked about older context you don't have, politely state you don't have access to that part of the conversation and ask for clarification.
` + "\n\n[Timeline Context]\n" + timelineContext

	// 4. Call Gemini 1.5 Flash with Streaming
	model := client.GenerativeModel(u.config.GeminiChatModel)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	var history []*genai.Content
	// Limit history to last 30 messages
	start := 0
	if len(input.History) > 30 {
		start = len(input.History) - 30
	}
	for i := start; i < len(input.History); i++ {
		h := input.History[i]
		role := h.Role
		if role == "assistant" || role == "model" {
			role = "model"
		} else {
			role = "user"
		}
		history = append(history, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(h.Text)},
		})
	}

	cs := model.StartChat()
	cs.History = history

	iter := cs.SendMessageStream(ctx, genai.Text(input.Query))

	// Return a reader that pipes tokens from the iterator
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		defer client.Close()
		for {
			resp, err := iter.Next()
			if err == io.EOF || err == iterator.Done {
				break
			}
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}

			if resp != nil {
				for _, cand := range resp.Candidates {
					if cand.Content != nil {
						for _, part := range cand.Content.Parts {
							if text, ok := part.(genai.Text); ok {
								_, _ = fmt.Fprint(pw, string(text))
							}
						}
					}
				}
			}
		}
	}()

	return pr, nil
}
