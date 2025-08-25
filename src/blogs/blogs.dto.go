package blogs

import "github.com/epsierra/phinex-blog-api/src/models"

// CreateBlogDto defines the input for creating a blog

type CreateBlogDto struct {
	Title              string   `json:"title" validate:"required" example:"My First Blog Post"`
	ExternalLink       string   `json:"externalLink" example:"https://example.com/related-article"`
	ExternalLinkTitle  string   `json:"externalLinkTitle" example:"Related Article"`
	Text               string   `json:"text" validate:"required" example:"This is the content of my first blog post."`
	Images             []string `json:"images" example:"https://example.com/image1.jpg,https://example.com/image2.jpg"`
	Video              string   `json:"video" example:"https://example.com/video.mp4"`
	Audio              string   `json:"audio" example:"https://example.com/updated-audio.mp3"`
	RepostedFromBlogId string   `json:"RepostedFromBlogId,omitempty" example:"some-other-blog-id"`
	Pinned             bool     `json:"pinned,omitempty" example:"false"`
	PinnedNumerOfDays  int      `json:"pinnedNumberOfDays,omitempty" example:"7"`
}

// UpdateBlogDto defines the input for updating a blog
type UpdateBlogDto struct {
	Title             string   `json:"title" example:"My Updated Blog Post"`
	ExternalLink      string   `json:"externalLink" example:"https://example.com/another-related-article"`
	ExternalLinkTitle string   `json:"externalLinkTitle" example:"Another Related Article"`
	Text              string   `json:"text" example:"This is the updated content of my blog post."`
	Images            []string `json:"images" example:"https://example.com/image3.jpg"`
	Video             string   `json:"video" example:"https://example.com/updated-video.mp4"`
	Audio             string   `json:"audio" example:"https://example.com/updated-audio.mp3"`
}

// FollowUnfollowDto defines the input for follow/unfollow

// BlogWithMeta represents a blog with metadata
type BlogWithMeta struct {
	Blog          models.Blog `json:"blog"`
	Liked         bool        `json:"liked"`
	Reposted      bool        `json:"reposted"`
	LikesCount    int64       `json:"likesCount"`
	RepostsCount  int64       `json:"repostsCount"`
	CommentsCount int64       `json:"commentsCount"`
	ViewsCount    int64       `json:"viewsCount"`
}

// CreateCommentDto defines the input for creating a comment
type CreateCommentDto struct {
	Text    string `json:"text" validate:"required" example:"This is a great post!"`
	Image   string `json:"image,omitempty" example:"https://example.com/comment-image.jpg"`
	Video   string `json:"video,omitempty" example:"https://example.com/comment-video.mp4"`
	Audio   string `json:"audio" example:"https://example.com/updated-audio.mp3"`
	Sticker string `json:"sticker,omitempty" example:"smiley_face"`
}

// CreateReplyDto defines the input for creating a reply
type CreateReplyDto struct {
	Text    string `json:"text" validate:"required" example:"I agree!"`
	Image   string `json:"image,omitempty" example:"https://example.com/reply-image.jpg"`
	Video   string `json:"video,omitempty" example:"https://example.com/reply-video.mp4"`
	Audio   string `json:"audio" example:"https://example.com/updated-audio.mp3"`
	Sticker string `json:"sticker,omitempty" example:"thumbs_up"`
}

// CommentWithMeta represents a comment with metadata
type CommentWithMeta struct {
	Comment      models.Comment `json:"comment"`
	RepliesCount int64          `json:"repliesCount"`
	LikesCount   int64          `json:"likesCount"`
	Liked        bool           `json:"liked"`
}

// LikeResponse represents the response for like/unlike actions
type LikeResponse struct {
	Liked      bool  `json:"liked"`
	LikesCount int64 `json:"likesCount"`
}

// FollowResponse represents the response for follow/unfollow actions
type FollowResponse struct {
	Followed bool `json:"followed"`
}

// MutationResponse represents the standard mutation response
type MutationResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
