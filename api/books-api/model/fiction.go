package model

import "time"

type Fiction struct {
	ID                string    `json:"id"`
	AuthorID          string    `json:"authorId"`
	AuthorName        string    `json:"authorName"`
	Title             string    `json:"title"`
	Synopsis          string    `json:"synopsis"`
	CoverURL          string    `json:"coverUrl"`
	Genres            []string  `json:"genres"`
	Tags              []string  `json:"tags"`
	GenresWereOmitted bool      `json:"-"`
	TagsWereOmitted   bool      `json:"-"`
	Status            string    `json:"status"` // ongoing, completed, hiatus
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	FollowerCount     int       `json:"followerCount"`
	ViewCount         int       `json:"viewCount"`
	ChapterCount      int       `json:"chapterCount"`
}
