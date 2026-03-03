package usecase

import (
	"context"
	"database/sql"

	"github.com/chanombude/twitter-go-api/internal/db"
)

type NotificationItem struct {
	db.Notification
	Actor                 UserItem
	TweetContent          *string
	TweetMediaUrl         *string
	OriginalTweetID       *int64
	OriginalTweetContent  *string
	OriginalTweetMediaUrl *string
}

func (u *Usecase) ListNotifications(ctx context.Context, userID int64, page, size int32) ([]NotificationItem, error) {
	rows, err := u.store.ListNotifications(ctx, db.ListNotificationsParams{
		RecipientID: userID,
		Limit:       size,
		Offset:      page * size,
	})
	if err != nil {
		return nil, err
	}
	return u.hydrateNotifications(ctx, rows)
}

func (u *Usecase) CountNotifications(ctx context.Context, userID int64) (int64, error) {
	return u.store.CountNotifications(ctx, userID)
}

func (u *Usecase) CountUnreadNotifications(ctx context.Context, userID int64) (int64, error) {
	return u.store.GetUnreadNotificationCount(ctx, userID)
}

func (u *Usecase) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return u.store.MarkAllNotificationsRead(ctx, userID)
}

func (u *Usecase) HydrateNotification(ctx context.Context, notification db.Notification) (NotificationItem, error) {
	items, err := u.hydrateNotifications(ctx, []db.Notification{notification})
	if err != nil {
		return NotificationItem{}, err
	}
	if len(items) == 0 {
		return NotificationItem{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (u *Usecase) hydrateNotifications(ctx context.Context, notifications []db.Notification) ([]NotificationItem, error) {
	if len(notifications) == 0 {
		return []NotificationItem{}, nil
	}

	actorIDsMap := make(map[int64]struct{}, len(notifications))
	tweetIDsMap := make(map[int64]struct{}, len(notifications))
	for _, n := range notifications {
		actorIDsMap[n.ActorID] = struct{}{}
		if n.TweetID != nil {
			tweetIDsMap[*n.TweetID] = struct{}{}
		}
	}

	actorIDs := make([]int64, 0, len(actorIDsMap))
	for id := range actorIDsMap {
		actorIDs = append(actorIDs, id)
	}

	tweetIDs := make([]int64, 0, len(tweetIDsMap))
	for id := range tweetIDsMap {
		tweetIDs = append(tweetIDs, id)
	}

	actors := make(map[int64]UserItem, len(actorIDs))
	if len(actorIDs) > 0 {
		rawActors, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
			UserIds:  actorIDs,
			ViewerID: nil,
		})
		if err != nil {
			return nil, err
		}
		for _, raw := range rawActors {
			actors[raw.User.ID] = UserItem{User: raw.User, IsFollowing: raw.IsFollowing}
		}
	}

	type tweetPreview struct {
		Content  *string
		Media    *string
		ParentID *int64
	}
	tweets := make(map[int64]tweetPreview, len(tweetIDs))
	if len(tweetIDs) > 0 {
		rawTweets, err := u.store.GetTweetsByIDs(ctx, db.GetTweetsByIDsParams{
			TweetIds: tweetIDs,
			ViewerID: nil,
		})
		if err != nil {
			return nil, err
		}

		parentIDsTemp := make(map[int64]struct{})
		for _, raw := range rawTweets {
			var content *string
			if raw.Tweet.Content != nil {
				content = raw.Tweet.Content
			}
			var media *string
			if raw.Tweet.MediaUrl != nil {
				media = raw.Tweet.MediaUrl
			}
			var parentId *int64
			if raw.Tweet.ParentID != nil {
				id := *raw.Tweet.ParentID
				parentId = &id
				parentIDsTemp[id] = struct{}{}
			}
			tweets[raw.Tweet.ID] = tweetPreview{Content: content, Media: media, ParentID: parentId}
		}

		var parentIDs []int64
		for pid := range parentIDsTemp {
			if _, ok := tweets[pid]; !ok {
				parentIDs = append(parentIDs, pid)
			}
		}

		if len(parentIDs) > 0 {
			parentRaw, err := u.store.GetTweetsByIDs(ctx, db.GetTweetsByIDsParams{
				TweetIds: parentIDs,
				ViewerID: nil,
			})
			if err == nil {
				for _, raw := range parentRaw {
					var content *string
					if raw.Tweet.Content != nil {
						content = raw.Tweet.Content
					}
					var media *string
					if raw.Tweet.MediaUrl != nil {
						media = raw.Tweet.MediaUrl
					}
					tweets[raw.Tweet.ID] = tweetPreview{Content: content, Media: media}
				}
			}
		}
	}

	items := make([]NotificationItem, 0, len(notifications))
	for _, n := range notifications {
		item := NotificationItem{
			Notification: n,
		}
		if actor, ok := actors[n.ActorID]; ok {
			item.Actor = actor
		}
		if n.TweetID != nil {
			if preview, ok := tweets[*n.TweetID]; ok {
				item.TweetContent = preview.Content
				item.TweetMediaUrl = preview.Media
				if n.Type == NotifTypeReply && preview.ParentID != nil {
					if parentPreview, ok := tweets[*preview.ParentID]; ok {
						item.OriginalTweetID = preview.ParentID
						item.OriginalTweetContent = parentPreview.Content
						item.OriginalTweetMediaUrl = parentPreview.Media
					}
				}
			}
		}
		items = append(items, item)
	}

	return items, nil
}
