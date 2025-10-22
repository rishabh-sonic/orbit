package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rishabh-sonic/orbit/internal/db"
)

type Service struct {
	q   db.Querier
	rdb *redis.Client
}

func NewService(q db.Querier, rdb *redis.Client) *Service {
	return &Service{q: q, rdb: rdb}
}

// ── Response types ────────────────────────────────────────────────────────────

// MeResponse is returned from /api/users/me and profile update endpoints.
// It includes private fields (email, role) that only the owner should see.
type MeResponse struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Avatar       *string   `json:"avatar"`
	Introduction *string   `json:"introduction"`
	Role         string    `json:"role"`
	Verified     bool      `json:"verified"`
	Banned       bool      `json:"banned"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserResponse is returned from GET /api/users/{identifier}.
// Includes public stats and the viewer's follow relationship.
type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	Avatar         *string   `json:"avatar"`
	Introduction   *string   `json:"introduction"`
	FollowerCount  int64     `json:"follower_count"`
	FollowingCount int64     `json:"following_count"`
	PostCount      int64     `json:"post_count"`
	IsFollowing    bool      `json:"is_following"`
	CreatedAt      time.Time `json:"created_at"`
}

// PublicUser is the minimal shape used in followers / following lists.
type PublicUser struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Avatar   *string   `json:"avatar"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func nullableStr(s sql.NullString) *string {
	if !s.Valid || s.String == "" {
		return nil
	}
	v := s.String
	return &v
}

func toMeResponse(u db.User) MeResponse {
	return MeResponse{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		Avatar:       nullableStr(u.Avatar),
		Introduction: nullableStr(u.Introduction),
		Role:         string(u.Role),
		Verified:     u.Verified,
		Banned:       u.Banned,
		CreatedAt:    u.CreatedAt,
	}
}

func toPublicUser(u db.User) PublicUser {
	return PublicUser{
		ID:       u.ID,
		Username: u.Username,
		Avatar:   nullableStr(u.Avatar),
	}
}

// ── Service methods ───────────────────────────────────────────────────────────

// GetByID returns the private profile used for the /me endpoint.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (MeResponse, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		return MeResponse{}, err
	}
	return toMeResponse(u), nil
}

// GetByIdentifier returns the public profile for any user, enriched with
// follower/following/post counts and whether viewerID is following them.
func (s *Service) GetByIdentifier(ctx context.Context, identifier string, viewerID *uuid.UUID) (UserResponse, error) {
	var (
		u   db.User
		err error
	)
	if id, parseErr := uuid.Parse(identifier); parseErr == nil {
		u, err = s.q.GetUserByID(ctx, id)
	} else {
		u, err = s.q.GetUserByUsername(ctx, identifier)
	}
	if err != nil {
		return UserResponse{}, err
	}

	followerCount, _ := s.q.GetFollowerCount(ctx, u.ID)
	followingCount, _ := s.q.GetFollowingCount(ctx, u.ID)
	postCount, _ := s.q.CountPostsByAuthor(ctx, u.ID)

	var isFollowing bool
	if viewerID != nil && *viewerID != u.ID {
		isFollowing, _ = s.q.IsFollowing(ctx, db.IsFollowingParams{
			FollowerID:  *viewerID,
			FollowingID: u.ID,
		})
	}

	return UserResponse{
		ID:             u.ID,
		Username:       u.Username,
		Avatar:         nullableStr(u.Avatar),
		Introduction:   nullableStr(u.Introduction),
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		PostCount:      postCount,
		IsFollowing:    isFollowing,
		CreatedAt:      u.CreatedAt,
	}, nil
}

// UpdateProfile applies partial updates and returns the updated MeResponse.
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (MeResponse, error) {
	var username, intro, avatar sql.NullString
	if in.Username != nil {
		existing, err := s.q.GetUserByUsername(ctx, *in.Username)
		if err == nil && existing.ID != userID {
			return MeResponse{}, fmt.Errorf("username already taken")
		}
		username = sql.NullString{String: *in.Username, Valid: true}
	}
	if in.Introduction != nil {
		intro = sql.NullString{String: *in.Introduction, Valid: true}
	}
	if in.Avatar != nil {
		avatar = sql.NullString{String: *in.Avatar, Valid: true}
	}

	u, err := s.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:           userID,
		Username:     username,
		Introduction: intro,
		Avatar:       avatar,
	})
	if err != nil {
		return MeResponse{}, err
	}
	return toMeResponse(u), nil
}

// GetFollowers returns the minimal public profile for each follower.
func (s *Service) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]PublicUser, error) {
	users, err := s.q.GetFollowers(ctx, db.GetFollowersParams{
		FollowingID: userID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, err
	}
	result := make([]PublicUser, len(users))
	for i, u := range users {
		result[i] = toPublicUser(u)
	}
	return result, nil
}

// GetFollowing returns the minimal public profile for each followed user.
func (s *Service) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]PublicUser, error) {
	users, err := s.q.GetFollowing(ctx, db.GetFollowingParams{
		FollowerID: userID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	result := make([]PublicUser, len(users))
	for i, u := range users {
		result[i] = toPublicUser(u)
	}
	return result, nil
}

func (s *Service) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	return s.q.FollowUser(ctx, db.FollowUserParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

func (s *Service) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	return s.q.UnfollowUser(ctx, db.UnfollowUserParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

func (s *Service) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return s.q.IsFollowing(ctx, db.IsFollowingParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

func (s *Service) RecordVisit(ctx context.Context, userID uuid.UUID) error {
	return s.q.RecordUserVisit(ctx, userID)
}

type UpdateProfileInput struct {
	Username     *string
	Introduction *string
	Avatar       *string
}
