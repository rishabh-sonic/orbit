// Package mock provides a configurable in-memory mock for db.Querier.
// Each method has a corresponding Fn field; if nil, a sensible default is used:
//   - "Get/List/Count/Is/Has/Search" → returns zero value + sql.ErrNoRows / 0, nil
//   - "Create" → returns zero value + nil (success)
//   - mutating methods (Set/Mark/Delete/Upsert/…) → nil error by default
package mock

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
)

// Querier is a configurable mock implementing db.Querier.
// Set the Fn fields you need in each test; all others use safe defaults.
type Querier struct {
	AddParticipantFn                   func(context.Context, db.AddParticipantParams) error
	CountCommentsByAuthorFn            func(context.Context, uuid.UUID) (int64, error)
	CountCommentsInRangeFn             func(context.Context, db.CountCommentsInRangeParams) (int64, error)
	CountDAUInRangeFn                  func(context.Context, db.CountDAUInRangeParams) ([]db.CountDAUInRangeRow, error)
	CountDailyActiveUsersFn            func(context.Context, time.Time) (int64, error)
	CountNewUsersInRangeFn             func(context.Context, db.CountNewUsersInRangeParams) (int64, error)
	CountPostsFn                       func(context.Context) (int64, error)
	CountPostsByAuthorFn               func(context.Context, uuid.UUID) (int64, error)
	CountPostsInRangeFn                func(context.Context, db.CountPostsInRangeParams) (int64, error)
	CountTopLevelCommentsFn            func(context.Context, uuid.UUID) (int64, error)
	CountUnreadNotificationsFn         func(context.Context, uuid.UUID) (int64, error)
	CountUsersFn                       func(context.Context) (int64, error)
	CreateCommentFn                    func(context.Context, db.CreateCommentParams) (db.Comment, error)
	CreateConversationFn               func(context.Context) (db.MessageConversation, error)
	CreateMessageFn                    func(context.Context, db.CreateMessageParams) (db.Message, error)
	CreateNotificationFn               func(context.Context, db.CreateNotificationParams) (db.Notification, error)
	CreateOAuthAccountFn               func(context.Context, db.CreateOAuthAccountParams) error
	CreatePostFn                       func(context.Context, db.CreatePostParams) (db.Post, error)
	CreatePushSubscriptionFn           func(context.Context, db.CreatePushSubscriptionParams) (db.PushSubscription, error)
	CreateUserFn                       func(context.Context, db.CreateUserParams) (db.User, error)
	CreateVerificationCodeFn           func(context.Context, db.CreateVerificationCodeParams) (db.VerificationCode, error)
	DecrementImageRefCountFn           func(context.Context, string) error
	DecrementPostCommentCountFn        func(context.Context, uuid.UUID) error
	DeletePushSubscriptionByEndpointFn func(context.Context, string) error
	DeleteVerificationCodesFn          func(context.Context, db.DeleteVerificationCodesParams) error
	FollowUserFn                       func(context.Context, db.FollowUserParams) error
	GetAdminUsersFn                    func(context.Context) ([]db.User, error)
	GetCommentByIDFn                   func(context.Context, uuid.UUID) (db.Comment, error)
	GetCommentByIDIncludeDeletedFn     func(context.Context, uuid.UUID) (db.Comment, error)
	GetConfigValueFn                   func(context.Context, string) (string, error)
	GetConversationBetweenUsersFn      func(context.Context, db.GetConversationBetweenUsersParams) (db.MessageConversation, error)
	GetConversationByIDFn              func(context.Context, uuid.UUID) (db.MessageConversation, error)
	GetEmailPrefFn                     func(context.Context, db.GetEmailPrefParams) (db.EmailPreference, error)
	GetEmailPreferencesFn              func(context.Context, uuid.UUID) ([]db.EmailPreference, error)
	GetFollowerCountFn                 func(context.Context, uuid.UUID) (int64, error)
	GetFollowerIDsFn                   func(context.Context, uuid.UUID) ([]uuid.UUID, error)
	GetFollowersFn                     func(context.Context, db.GetFollowersParams) ([]db.User, error)
	GetFollowingFn                     func(context.Context, db.GetFollowingParams) ([]db.User, error)
	GetFollowingCountFn                func(context.Context, uuid.UUID) (int64, error)
	GetFollowingIDsFn                  func(context.Context, uuid.UUID) ([]uuid.UUID, error)
	GetImageByURLFn                    func(context.Context, string) (db.Image, error)
	GetMessageByIDFn                   func(context.Context, uuid.UUID) (db.Message, error)
	GetNotificationByIDFn              func(context.Context, uuid.UUID) (db.Notification, error)
	GetNotificationPrefFn              func(context.Context, db.GetNotificationPrefParams) (db.NotificationPreference, error)
	GetNotificationPreferencesFn       func(context.Context, uuid.UUID) ([]db.NotificationPreference, error)
	GetOAuthAccountFn                  func(context.Context, db.GetOAuthAccountParams) (db.OauthAccount, error)
	GetOtherParticipantUserIDFn        func(context.Context, db.GetOtherParticipantUserIDParams) (uuid.UUID, error)
	GetParticipantFn                   func(context.Context, db.GetParticipantParams) (db.MessageParticipant, error)
	GetParticipantsFn                  func(context.Context, uuid.UUID) ([]db.MessageParticipant, error)
	GetPostByIDFn                      func(context.Context, uuid.UUID) (db.Post, error)
	GetPostByIDIncludeDeletedFn        func(context.Context, uuid.UUID) (db.Post, error)
	GetPostSubscribersFn               func(context.Context, uuid.UUID) ([]uuid.UUID, error)
	GetPushSubscriptionsByUserIDFn     func(context.Context, uuid.UUID) ([]db.PushSubscription, error)
	GetSubscribedPostIDsFn             func(context.Context, uuid.UUID) ([]uuid.UUID, error)
	GetTotalUnreadCountFn              func(context.Context, uuid.UUID) (int64, error)
	GetUserByEmailFn                   func(context.Context, string) (db.User, error)
	GetUserByEmailOrUsernameFn         func(context.Context, db.GetUserByEmailOrUsernameParams) (db.User, error)
	GetUserByIDFn                      func(context.Context, uuid.UUID) (db.User, error)
	GetUserByUsernameFn                func(context.Context, string) (db.User, error)
	GetVerificationCodeFn              func(context.Context, db.GetVerificationCodeParams) (db.VerificationCode, error)
	HasUserReadPostFn                  func(context.Context, db.HasUserReadPostParams) (bool, error)
	IncrementPostCommentCountFn        func(context.Context, uuid.UUID) error
	IncrementPostViewsFn               func(context.Context, uuid.UUID) error
	IncrementUnreadCountsFn            func(context.Context, db.IncrementUnreadCountsParams) error
	IsFollowingFn                      func(context.Context, db.IsFollowingParams) (bool, error)
	IsSubscribedToPostFn               func(context.Context, db.IsSubscribedToPostParams) (bool, error)
	ListCommentsByAuthorFn             func(context.Context, db.ListCommentsByAuthorParams) ([]db.Comment, error)
	ListConfigFn                       func(context.Context) ([]db.SiteConfig, error)
	ListConversationsForUserFn         func(context.Context, db.ListConversationsForUserParams) ([]db.MessageConversation, error)
	ListFeaturedPostsFn                func(context.Context, db.ListFeaturedPostsParams) ([]db.Post, error)
	ListMessagesFn                     func(context.Context, db.ListMessagesParams) ([]db.Message, error)
	ListNotificationsFn                func(context.Context, db.ListNotificationsParams) ([]db.Notification, error)
	ListPostsFn                        func(context.Context, db.ListPostsParams) ([]db.Post, error)
	ListPostsByAuthorFn                func(context.Context, db.ListPostsByAuthorParams) ([]db.Post, error)
	ListRecentPostsFn                  func(context.Context, db.ListRecentPostsParams) ([]db.Post, error)
	ListRepliesFn                      func(context.Context, db.ListRepliesParams) ([]db.Comment, error)
	ListTopLevelCommentsFn             func(context.Context, db.ListTopLevelCommentsParams) ([]db.Comment, error)
	ListUsersFn                        func(context.Context, db.ListUsersParams) ([]db.User, error)
	MarkConversationReadFn             func(context.Context, db.MarkConversationReadParams) error
	MarkNotificationReadByIDFn         func(context.Context, uuid.UUID) error
	MarkNotificationsReadFn            func(context.Context, uuid.UUID) error
	MarkVerificationCodeUsedFn         func(context.Context, uuid.UUID) error
	PinCommentFn                       func(context.Context, uuid.UUID) error
	PinPostFn                          func(context.Context, uuid.UUID) error
	RecordPostReadFn                   func(context.Context, db.RecordPostReadParams) error
	RecordUserVisitFn                  func(context.Context, uuid.UUID) error
	SearchUsersFn                      func(context.Context, db.SearchUsersParams) ([]db.User, error)
	SetPostClosedFn                    func(context.Context, db.SetPostClosedParams) error
	SetUserBannedFn                    func(context.Context, db.SetUserBannedParams) error
	SetUserPasswordHashFn              func(context.Context, db.SetUserPasswordHashParams) error
	SetUserRoleFn                      func(context.Context, db.SetUserRoleParams) error
	SetUserVerifiedFn                  func(context.Context, uuid.UUID) error
	SoftDeleteCommentFn                func(context.Context, uuid.UUID) error
	SoftDeleteMessageFn                func(context.Context, uuid.UUID) error
	SoftDeletePostFn                   func(context.Context, uuid.UUID) error
	SubscribeToPostFn                  func(context.Context, db.SubscribeToPostParams) error
	UnfollowUserFn                     func(context.Context, db.UnfollowUserParams) error
	UnpinCommentFn                     func(context.Context, uuid.UUID) error
	UnpinPostFn                        func(context.Context, uuid.UUID) error
	UnsubscribeFromPostFn              func(context.Context, db.UnsubscribeFromPostParams) error
	UpdateCommentFn                    func(context.Context, db.UpdateCommentParams) (db.Comment, error)
	UpdateConversationLastMessageFn    func(context.Context, uuid.UUID) error
	UpdatePostFn                       func(context.Context, db.UpdatePostParams) (db.Post, error)
	UpdateUserFn                       func(context.Context, db.UpdateUserParams) (db.User, error)
	UpsertConfigValueFn                func(context.Context, db.UpsertConfigValueParams) error
	UpsertEmailPreferenceFn            func(context.Context, db.UpsertEmailPreferenceParams) error
	UpsertImageFn                      func(context.Context, string) (db.Image, error)
	UpsertNotificationPreferenceFn     func(context.Context, db.UpsertNotificationPreferenceParams) error
}

// Compile-time check.
var _ db.Querier = (*Querier)(nil)

func (m *Querier) AddParticipant(ctx context.Context, arg db.AddParticipantParams) error {
	if m.AddParticipantFn != nil {
		return m.AddParticipantFn(ctx, arg)
	}
	return nil
}

func (m *Querier) CountCommentsByAuthor(ctx context.Context, authorID uuid.UUID) (int64, error) {
	if m.CountCommentsByAuthorFn != nil {
		return m.CountCommentsByAuthorFn(ctx, authorID)
	}
	return 0, nil
}

func (m *Querier) CountCommentsInRange(ctx context.Context, arg db.CountCommentsInRangeParams) (int64, error) {
	if m.CountCommentsInRangeFn != nil {
		return m.CountCommentsInRangeFn(ctx, arg)
	}
	return 0, nil
}

func (m *Querier) CountDAUInRange(ctx context.Context, arg db.CountDAUInRangeParams) ([]db.CountDAUInRangeRow, error) {
	if m.CountDAUInRangeFn != nil {
		return m.CountDAUInRangeFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) CountDailyActiveUsers(ctx context.Context, visitDate time.Time) (int64, error) {
	if m.CountDailyActiveUsersFn != nil {
		return m.CountDailyActiveUsersFn(ctx, visitDate)
	}
	return 0, nil
}

func (m *Querier) CountNewUsersInRange(ctx context.Context, arg db.CountNewUsersInRangeParams) (int64, error) {
	if m.CountNewUsersInRangeFn != nil {
		return m.CountNewUsersInRangeFn(ctx, arg)
	}
	return 0, nil
}

func (m *Querier) CountPosts(ctx context.Context) (int64, error) {
	if m.CountPostsFn != nil {
		return m.CountPostsFn(ctx)
	}
	return 0, nil
}

func (m *Querier) CountPostsByAuthor(ctx context.Context, authorID uuid.UUID) (int64, error) {
	if m.CountPostsByAuthorFn != nil {
		return m.CountPostsByAuthorFn(ctx, authorID)
	}
	return 0, nil
}

func (m *Querier) CountPostsInRange(ctx context.Context, arg db.CountPostsInRangeParams) (int64, error) {
	if m.CountPostsInRangeFn != nil {
		return m.CountPostsInRangeFn(ctx, arg)
	}
	return 0, nil
}

func (m *Querier) CountTopLevelComments(ctx context.Context, postID uuid.UUID) (int64, error) {
	if m.CountTopLevelCommentsFn != nil {
		return m.CountTopLevelCommentsFn(ctx, postID)
	}
	return 0, nil
}

func (m *Querier) CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.CountUnreadNotificationsFn != nil {
		return m.CountUnreadNotificationsFn(ctx, userID)
	}
	return 0, nil
}

func (m *Querier) CountUsers(ctx context.Context) (int64, error) {
	if m.CountUsersFn != nil {
		return m.CountUsersFn(ctx)
	}
	return 0, nil
}

func (m *Querier) CreateComment(ctx context.Context, arg db.CreateCommentParams) (db.Comment, error) {
	if m.CreateCommentFn != nil {
		return m.CreateCommentFn(ctx, arg)
	}
	return db.Comment{ID: uuid.New(), Content: arg.Content, AuthorID: arg.AuthorID, PostID: arg.PostID}, nil
}

func (m *Querier) CreateConversation(ctx context.Context) (db.MessageConversation, error) {
	if m.CreateConversationFn != nil {
		return m.CreateConversationFn(ctx)
	}
	return db.MessageConversation{ID: uuid.New()}, nil
}

func (m *Querier) CreateMessage(ctx context.Context, arg db.CreateMessageParams) (db.Message, error) {
	if m.CreateMessageFn != nil {
		return m.CreateMessageFn(ctx, arg)
	}
	return db.Message{ID: uuid.New(), ConversationID: arg.ConversationID, SenderID: arg.SenderID, Content: arg.Content}, nil
}

func (m *Querier) CreateNotification(ctx context.Context, arg db.CreateNotificationParams) (db.Notification, error) {
	if m.CreateNotificationFn != nil {
		return m.CreateNotificationFn(ctx, arg)
	}
	return db.Notification{ID: uuid.New()}, nil
}

func (m *Querier) CreateOAuthAccount(ctx context.Context, arg db.CreateOAuthAccountParams) error {
	if m.CreateOAuthAccountFn != nil {
		return m.CreateOAuthAccountFn(ctx, arg)
	}
	return nil
}

func (m *Querier) CreatePost(ctx context.Context, arg db.CreatePostParams) (db.Post, error) {
	if m.CreatePostFn != nil {
		return m.CreatePostFn(ctx, arg)
	}
	return db.Post{ID: uuid.New(), Title: arg.Title, Content: arg.Content, AuthorID: arg.AuthorID}, nil
}

func (m *Querier) CreatePushSubscription(ctx context.Context, arg db.CreatePushSubscriptionParams) (db.PushSubscription, error) {
	if m.CreatePushSubscriptionFn != nil {
		return m.CreatePushSubscriptionFn(ctx, arg)
	}
	return db.PushSubscription{ID: uuid.New()}, nil
}

func (m *Querier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, arg)
	}
	return db.User{ID: uuid.New(), Username: arg.Username, Email: arg.Email, Role: arg.Role, Verified: arg.Verified}, nil
}

func (m *Querier) CreateVerificationCode(ctx context.Context, arg db.CreateVerificationCodeParams) (db.VerificationCode, error) {
	if m.CreateVerificationCodeFn != nil {
		return m.CreateVerificationCodeFn(ctx, arg)
	}
	return db.VerificationCode{ID: uuid.New(), Email: arg.Email, Code: arg.Code, Type: arg.Type}, nil
}

func (m *Querier) DecrementImageRefCount(ctx context.Context, url string) error {
	if m.DecrementImageRefCountFn != nil {
		return m.DecrementImageRefCountFn(ctx, url)
	}
	return nil
}

func (m *Querier) DecrementPostCommentCount(ctx context.Context, id uuid.UUID) error {
	if m.DecrementPostCommentCountFn != nil {
		return m.DecrementPostCommentCountFn(ctx, id)
	}
	return nil
}

func (m *Querier) DeletePushSubscriptionByEndpoint(ctx context.Context, endpoint string) error {
	if m.DeletePushSubscriptionByEndpointFn != nil {
		return m.DeletePushSubscriptionByEndpointFn(ctx, endpoint)
	}
	return nil
}

func (m *Querier) DeleteVerificationCodes(ctx context.Context, arg db.DeleteVerificationCodesParams) error {
	if m.DeleteVerificationCodesFn != nil {
		return m.DeleteVerificationCodesFn(ctx, arg)
	}
	return nil
}

func (m *Querier) FollowUser(ctx context.Context, arg db.FollowUserParams) error {
	if m.FollowUserFn != nil {
		return m.FollowUserFn(ctx, arg)
	}
	return nil
}

func (m *Querier) GetAdminUsers(ctx context.Context) ([]db.User, error) {
	if m.GetAdminUsersFn != nil {
		return m.GetAdminUsersFn(ctx)
	}
	return nil, nil
}

func (m *Querier) GetCommentByID(ctx context.Context, id uuid.UUID) (db.Comment, error) {
	if m.GetCommentByIDFn != nil {
		return m.GetCommentByIDFn(ctx, id)
	}
	return db.Comment{}, sql.ErrNoRows
}

func (m *Querier) GetCommentByIDIncludeDeleted(ctx context.Context, id uuid.UUID) (db.Comment, error) {
	if m.GetCommentByIDIncludeDeletedFn != nil {
		return m.GetCommentByIDIncludeDeletedFn(ctx, id)
	}
	return db.Comment{}, sql.ErrNoRows
}

func (m *Querier) GetConfigValue(ctx context.Context, key string) (string, error) {
	if m.GetConfigValueFn != nil {
		return m.GetConfigValueFn(ctx, key)
	}
	return "", sql.ErrNoRows
}

func (m *Querier) GetConversationBetweenUsers(ctx context.Context, arg db.GetConversationBetweenUsersParams) (db.MessageConversation, error) {
	if m.GetConversationBetweenUsersFn != nil {
		return m.GetConversationBetweenUsersFn(ctx, arg)
	}
	return db.MessageConversation{}, sql.ErrNoRows
}

func (m *Querier) GetConversationByID(ctx context.Context, id uuid.UUID) (db.MessageConversation, error) {
	if m.GetConversationByIDFn != nil {
		return m.GetConversationByIDFn(ctx, id)
	}
	return db.MessageConversation{}, sql.ErrNoRows
}

func (m *Querier) GetEmailPref(ctx context.Context, arg db.GetEmailPrefParams) (db.EmailPreference, error) {
	if m.GetEmailPrefFn != nil {
		return m.GetEmailPrefFn(ctx, arg)
	}
	return db.EmailPreference{}, sql.ErrNoRows
}

func (m *Querier) GetEmailPreferences(ctx context.Context, userID uuid.UUID) ([]db.EmailPreference, error) {
	if m.GetEmailPreferencesFn != nil {
		return m.GetEmailPreferencesFn(ctx, userID)
	}
	return nil, nil
}

func (m *Querier) GetFollowerCount(ctx context.Context, followingID uuid.UUID) (int64, error) {
	if m.GetFollowerCountFn != nil {
		return m.GetFollowerCountFn(ctx, followingID)
	}
	return 0, nil
}

func (m *Querier) GetFollowerIDs(ctx context.Context, followingID uuid.UUID) ([]uuid.UUID, error) {
	if m.GetFollowerIDsFn != nil {
		return m.GetFollowerIDsFn(ctx, followingID)
	}
	return nil, nil
}

func (m *Querier) GetFollowers(ctx context.Context, arg db.GetFollowersParams) ([]db.User, error) {
	if m.GetFollowersFn != nil {
		return m.GetFollowersFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) GetFollowing(ctx context.Context, arg db.GetFollowingParams) ([]db.User, error) {
	if m.GetFollowingFn != nil {
		return m.GetFollowingFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) GetFollowingCount(ctx context.Context, followerID uuid.UUID) (int64, error) {
	if m.GetFollowingCountFn != nil {
		return m.GetFollowingCountFn(ctx, followerID)
	}
	return 0, nil
}

func (m *Querier) GetFollowingIDs(ctx context.Context, followerID uuid.UUID) ([]uuid.UUID, error) {
	if m.GetFollowingIDsFn != nil {
		return m.GetFollowingIDsFn(ctx, followerID)
	}
	return nil, nil
}

func (m *Querier) GetImageByURL(ctx context.Context, url string) (db.Image, error) {
	if m.GetImageByURLFn != nil {
		return m.GetImageByURLFn(ctx, url)
	}
	return db.Image{}, sql.ErrNoRows
}

func (m *Querier) GetMessageByID(ctx context.Context, id uuid.UUID) (db.Message, error) {
	if m.GetMessageByIDFn != nil {
		return m.GetMessageByIDFn(ctx, id)
	}
	return db.Message{}, sql.ErrNoRows
}

func (m *Querier) GetNotificationByID(ctx context.Context, id uuid.UUID) (db.Notification, error) {
	if m.GetNotificationByIDFn != nil {
		return m.GetNotificationByIDFn(ctx, id)
	}
	return db.Notification{}, sql.ErrNoRows
}

func (m *Querier) GetNotificationPref(ctx context.Context, arg db.GetNotificationPrefParams) (db.NotificationPreference, error) {
	if m.GetNotificationPrefFn != nil {
		return m.GetNotificationPrefFn(ctx, arg)
	}
	return db.NotificationPreference{}, sql.ErrNoRows
}

func (m *Querier) GetNotificationPreferences(ctx context.Context, userID uuid.UUID) ([]db.NotificationPreference, error) {
	if m.GetNotificationPreferencesFn != nil {
		return m.GetNotificationPreferencesFn(ctx, userID)
	}
	return nil, nil
}

func (m *Querier) GetOAuthAccount(ctx context.Context, arg db.GetOAuthAccountParams) (db.OauthAccount, error) {
	if m.GetOAuthAccountFn != nil {
		return m.GetOAuthAccountFn(ctx, arg)
	}
	return db.OauthAccount{}, sql.ErrNoRows
}

func (m *Querier) GetOtherParticipantUserID(ctx context.Context, arg db.GetOtherParticipantUserIDParams) (uuid.UUID, error) {
	if m.GetOtherParticipantUserIDFn != nil {
		return m.GetOtherParticipantUserIDFn(ctx, arg)
	}
	return uuid.UUID{}, sql.ErrNoRows
}

func (m *Querier) GetParticipant(ctx context.Context, arg db.GetParticipantParams) (db.MessageParticipant, error) {
	if m.GetParticipantFn != nil {
		return m.GetParticipantFn(ctx, arg)
	}
	return db.MessageParticipant{}, sql.ErrNoRows
}

func (m *Querier) GetParticipants(ctx context.Context, conversationID uuid.UUID) ([]db.MessageParticipant, error) {
	if m.GetParticipantsFn != nil {
		return m.GetParticipantsFn(ctx, conversationID)
	}
	return nil, nil
}

func (m *Querier) GetPostByID(ctx context.Context, id uuid.UUID) (db.Post, error) {
	if m.GetPostByIDFn != nil {
		return m.GetPostByIDFn(ctx, id)
	}
	return db.Post{}, sql.ErrNoRows
}

func (m *Querier) GetPostByIDIncludeDeleted(ctx context.Context, id uuid.UUID) (db.Post, error) {
	if m.GetPostByIDIncludeDeletedFn != nil {
		return m.GetPostByIDIncludeDeletedFn(ctx, id)
	}
	return db.Post{}, sql.ErrNoRows
}

func (m *Querier) GetPostSubscribers(ctx context.Context, postID uuid.UUID) ([]uuid.UUID, error) {
	if m.GetPostSubscribersFn != nil {
		return m.GetPostSubscribersFn(ctx, postID)
	}
	return nil, nil
}

func (m *Querier) GetPushSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]db.PushSubscription, error) {
	if m.GetPushSubscriptionsByUserIDFn != nil {
		return m.GetPushSubscriptionsByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *Querier) GetSubscribedPostIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if m.GetSubscribedPostIDsFn != nil {
		return m.GetSubscribedPostIDsFn(ctx, userID)
	}
	return nil, nil
}

func (m *Querier) GetTotalUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.GetTotalUnreadCountFn != nil {
		return m.GetTotalUnreadCountFn(ctx, userID)
	}
	return 0, nil
}

func (m *Querier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(ctx, email)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *Querier) GetUserByEmailOrUsername(ctx context.Context, arg db.GetUserByEmailOrUsernameParams) (db.User, error) {
	if m.GetUserByEmailOrUsernameFn != nil {
		return m.GetUserByEmailOrUsernameFn(ctx, arg)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *Querier) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *Querier) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	if m.GetUserByUsernameFn != nil {
		return m.GetUserByUsernameFn(ctx, username)
	}
	return db.User{}, sql.ErrNoRows
}

func (m *Querier) GetVerificationCode(ctx context.Context, arg db.GetVerificationCodeParams) (db.VerificationCode, error) {
	if m.GetVerificationCodeFn != nil {
		return m.GetVerificationCodeFn(ctx, arg)
	}
	return db.VerificationCode{}, sql.ErrNoRows
}

func (m *Querier) HasUserReadPost(ctx context.Context, arg db.HasUserReadPostParams) (bool, error) {
	if m.HasUserReadPostFn != nil {
		return m.HasUserReadPostFn(ctx, arg)
	}
	return false, nil
}

func (m *Querier) IncrementPostCommentCount(ctx context.Context, id uuid.UUID) error {
	if m.IncrementPostCommentCountFn != nil {
		return m.IncrementPostCommentCountFn(ctx, id)
	}
	return nil
}

func (m *Querier) IncrementPostViews(ctx context.Context, id uuid.UUID) error {
	if m.IncrementPostViewsFn != nil {
		return m.IncrementPostViewsFn(ctx, id)
	}
	return nil
}

func (m *Querier) IncrementUnreadCounts(ctx context.Context, arg db.IncrementUnreadCountsParams) error {
	if m.IncrementUnreadCountsFn != nil {
		return m.IncrementUnreadCountsFn(ctx, arg)
	}
	return nil
}

func (m *Querier) IsFollowing(ctx context.Context, arg db.IsFollowingParams) (bool, error) {
	if m.IsFollowingFn != nil {
		return m.IsFollowingFn(ctx, arg)
	}
	return false, nil
}

func (m *Querier) IsSubscribedToPost(ctx context.Context, arg db.IsSubscribedToPostParams) (bool, error) {
	if m.IsSubscribedToPostFn != nil {
		return m.IsSubscribedToPostFn(ctx, arg)
	}
	return false, nil
}

func (m *Querier) ListCommentsByAuthor(ctx context.Context, arg db.ListCommentsByAuthorParams) ([]db.Comment, error) {
	if m.ListCommentsByAuthorFn != nil {
		return m.ListCommentsByAuthorFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListConfig(ctx context.Context) ([]db.SiteConfig, error) {
	if m.ListConfigFn != nil {
		return m.ListConfigFn(ctx)
	}
	return nil, nil
}

func (m *Querier) ListConversationsForUser(ctx context.Context, arg db.ListConversationsForUserParams) ([]db.MessageConversation, error) {
	if m.ListConversationsForUserFn != nil {
		return m.ListConversationsForUserFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListFeaturedPosts(ctx context.Context, arg db.ListFeaturedPostsParams) ([]db.Post, error) {
	if m.ListFeaturedPostsFn != nil {
		return m.ListFeaturedPostsFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListMessages(ctx context.Context, arg db.ListMessagesParams) ([]db.Message, error) {
	if m.ListMessagesFn != nil {
		return m.ListMessagesFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListNotifications(ctx context.Context, arg db.ListNotificationsParams) ([]db.Notification, error) {
	if m.ListNotificationsFn != nil {
		return m.ListNotificationsFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListPosts(ctx context.Context, arg db.ListPostsParams) ([]db.Post, error) {
	if m.ListPostsFn != nil {
		return m.ListPostsFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListPostsByAuthor(ctx context.Context, arg db.ListPostsByAuthorParams) ([]db.Post, error) {
	if m.ListPostsByAuthorFn != nil {
		return m.ListPostsByAuthorFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListRecentPosts(ctx context.Context, arg db.ListRecentPostsParams) ([]db.Post, error) {
	if m.ListRecentPostsFn != nil {
		return m.ListRecentPostsFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListReplies(ctx context.Context, arg db.ListRepliesParams) ([]db.Comment, error) {
	if m.ListRepliesFn != nil {
		return m.ListRepliesFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListTopLevelComments(ctx context.Context, arg db.ListTopLevelCommentsParams) ([]db.Comment, error) {
	if m.ListTopLevelCommentsFn != nil {
		return m.ListTopLevelCommentsFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	if m.ListUsersFn != nil {
		return m.ListUsersFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) MarkConversationRead(ctx context.Context, arg db.MarkConversationReadParams) error {
	if m.MarkConversationReadFn != nil {
		return m.MarkConversationReadFn(ctx, arg)
	}
	return nil
}

func (m *Querier) MarkNotificationReadByID(ctx context.Context, id uuid.UUID) error {
	if m.MarkNotificationReadByIDFn != nil {
		return m.MarkNotificationReadByIDFn(ctx, id)
	}
	return nil
}

func (m *Querier) MarkNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	if m.MarkNotificationsReadFn != nil {
		return m.MarkNotificationsReadFn(ctx, userID)
	}
	return nil
}

func (m *Querier) MarkVerificationCodeUsed(ctx context.Context, id uuid.UUID) error {
	if m.MarkVerificationCodeUsedFn != nil {
		return m.MarkVerificationCodeUsedFn(ctx, id)
	}
	return nil
}

func (m *Querier) PinComment(ctx context.Context, id uuid.UUID) error {
	if m.PinCommentFn != nil {
		return m.PinCommentFn(ctx, id)
	}
	return nil
}

func (m *Querier) PinPost(ctx context.Context, id uuid.UUID) error {
	if m.PinPostFn != nil {
		return m.PinPostFn(ctx, id)
	}
	return nil
}

func (m *Querier) RecordPostRead(ctx context.Context, arg db.RecordPostReadParams) error {
	if m.RecordPostReadFn != nil {
		return m.RecordPostReadFn(ctx, arg)
	}
	return nil
}

func (m *Querier) RecordUserVisit(ctx context.Context, userID uuid.UUID) error {
	if m.RecordUserVisitFn != nil {
		return m.RecordUserVisitFn(ctx, userID)
	}
	return nil
}

func (m *Querier) SearchUsers(ctx context.Context, arg db.SearchUsersParams) ([]db.User, error) {
	if m.SearchUsersFn != nil {
		return m.SearchUsersFn(ctx, arg)
	}
	return nil, nil
}

func (m *Querier) SetPostClosed(ctx context.Context, arg db.SetPostClosedParams) error {
	if m.SetPostClosedFn != nil {
		return m.SetPostClosedFn(ctx, arg)
	}
	return nil
}

func (m *Querier) SetUserBanned(ctx context.Context, arg db.SetUserBannedParams) error {
	if m.SetUserBannedFn != nil {
		return m.SetUserBannedFn(ctx, arg)
	}
	return nil
}

func (m *Querier) SetUserPasswordHash(ctx context.Context, arg db.SetUserPasswordHashParams) error {
	if m.SetUserPasswordHashFn != nil {
		return m.SetUserPasswordHashFn(ctx, arg)
	}
	return nil
}

func (m *Querier) SetUserRole(ctx context.Context, arg db.SetUserRoleParams) error {
	if m.SetUserRoleFn != nil {
		return m.SetUserRoleFn(ctx, arg)
	}
	return nil
}

func (m *Querier) SetUserVerified(ctx context.Context, id uuid.UUID) error {
	if m.SetUserVerifiedFn != nil {
		return m.SetUserVerifiedFn(ctx, id)
	}
	return nil
}

func (m *Querier) SoftDeleteComment(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeleteCommentFn != nil {
		return m.SoftDeleteCommentFn(ctx, id)
	}
	return nil
}

func (m *Querier) SoftDeleteMessage(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeleteMessageFn != nil {
		return m.SoftDeleteMessageFn(ctx, id)
	}
	return nil
}

func (m *Querier) SoftDeletePost(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeletePostFn != nil {
		return m.SoftDeletePostFn(ctx, id)
	}
	return nil
}

func (m *Querier) SubscribeToPost(ctx context.Context, arg db.SubscribeToPostParams) error {
	if m.SubscribeToPostFn != nil {
		return m.SubscribeToPostFn(ctx, arg)
	}
	return nil
}

func (m *Querier) UnfollowUser(ctx context.Context, arg db.UnfollowUserParams) error {
	if m.UnfollowUserFn != nil {
		return m.UnfollowUserFn(ctx, arg)
	}
	return nil
}

func (m *Querier) UnpinComment(ctx context.Context, id uuid.UUID) error {
	if m.UnpinCommentFn != nil {
		return m.UnpinCommentFn(ctx, id)
	}
	return nil
}

func (m *Querier) UnpinPost(ctx context.Context, id uuid.UUID) error {
	if m.UnpinPostFn != nil {
		return m.UnpinPostFn(ctx, id)
	}
	return nil
}

func (m *Querier) UnsubscribeFromPost(ctx context.Context, arg db.UnsubscribeFromPostParams) error {
	if m.UnsubscribeFromPostFn != nil {
		return m.UnsubscribeFromPostFn(ctx, arg)
	}
	return nil
}

func (m *Querier) UpdateComment(ctx context.Context, arg db.UpdateCommentParams) (db.Comment, error) {
	if m.UpdateCommentFn != nil {
		return m.UpdateCommentFn(ctx, arg)
	}
	return db.Comment{}, nil
}

func (m *Querier) UpdateConversationLastMessage(ctx context.Context, id uuid.UUID) error {
	if m.UpdateConversationLastMessageFn != nil {
		return m.UpdateConversationLastMessageFn(ctx, id)
	}
	return nil
}

func (m *Querier) UpdatePost(ctx context.Context, arg db.UpdatePostParams) (db.Post, error) {
	if m.UpdatePostFn != nil {
		return m.UpdatePostFn(ctx, arg)
	}
	return db.Post{ID: arg.ID}, nil
}

func (m *Querier) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	if m.UpdateUserFn != nil {
		return m.UpdateUserFn(ctx, arg)
	}
	return db.User{ID: arg.ID}, nil
}

func (m *Querier) UpsertConfigValue(ctx context.Context, arg db.UpsertConfigValueParams) error {
	if m.UpsertConfigValueFn != nil {
		return m.UpsertConfigValueFn(ctx, arg)
	}
	return nil
}

func (m *Querier) UpsertEmailPreference(ctx context.Context, arg db.UpsertEmailPreferenceParams) error {
	if m.UpsertEmailPreferenceFn != nil {
		return m.UpsertEmailPreferenceFn(ctx, arg)
	}
	return nil
}

func (m *Querier) UpsertImage(ctx context.Context, url string) (db.Image, error) {
	if m.UpsertImageFn != nil {
		return m.UpsertImageFn(ctx, url)
	}
	return db.Image{ID: uuid.New(), Url: url}, nil
}

func (m *Querier) UpsertNotificationPreference(ctx context.Context, arg db.UpsertNotificationPreferenceParams) error {
	if m.UpsertNotificationPreferenceFn != nil {
		return m.UpsertNotificationPreferenceFn(ctx, arg)
	}
	return nil
}
