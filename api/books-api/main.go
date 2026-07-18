package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/guhaag/writerslife-books-api/db"
	"github.com/guhaag/writerslife-books-api/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// ── JWT ──────────────────────────────────────────────────────────────────────

var jwtSecret = []byte("writerslife-dev-secret")

type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func makeToken(user *model.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

var store *db.Store

// ── Middleware ────────────────────────────────────────────────────────────────

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := parseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func getUser(c *gin.Context) (*model.User, error) {
	uid, ok := c.Get("userID")
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return store.FindUserByID(c.Request.Context(), uid.(string))
}

func databaseError(c *gin.Context, operation string, err error) {
	log.Printf("database %s: %v", operation, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func validFictionStatus(status string) bool {
	return status == "ongoing" || status == "completed" || status == "hiatus"
}

func validChapterStatus(status string) bool {
	return status == "draft" || status == "published"
}

// ── Auth handlers ─────────────────────────────────────────────────────────────

func registerHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Bio      string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := store.FindUserByName(c.Request.Context(), req.Username); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		databaseError(c, "find username during registration", err)
		return
	}
	if _, err := store.FindUserByEmail(c.Request.Context(), req.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		databaseError(c, "find email during registration", err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}

	u := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Bio:          req.Bio,
		CreatedAt:    time.Now(),
	}
	if err := store.CreateUser(c.Request.Context(), u); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" {
				c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
			} else {
				c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			}
			return
		}
		databaseError(c, "create user", err)
		return
	}

	token, _ := makeToken(u)
	c.JSON(http.StatusCreated, gin.H{"token": token, "username": u.Username})
}

func loginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := store.FindUserByName(c.Request.Context(), req.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	if err != nil {
		databaseError(c, "find user during login", err)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	token, _ := makeToken(u)
	c.JSON(http.StatusOK, gin.H{"token": token, "username": u.Username})
}

// ── Fiction handlers ──────────────────────────────────────────────────────────

func listFictionsHandler(c *gin.Context) {
	search := strings.ToLower(c.Query("search"))
	genre := strings.ToLower(c.Query("genre"))
	sort := c.Query("sort") // recent | popular

	result, err := store.ListFictions(c.Request.Context(), search, genre, sort)
	if err != nil {
		databaseError(c, "list fictions", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func getFictionHandler(c *gin.Context) {
	id := c.Param("id")
	f, err := store.IncrementFictionViews(c.Request.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fiction not found"})
		return
	}
	if err != nil {
		databaseError(c, "increment fiction views", err)
		return
	}
	c.JSON(http.StatusOK, f)
}

func createFictionHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var req struct {
		Title    string   `json:"title" binding:"required"`
		Synopsis string   `json:"synopsis"`
		CoverURL string   `json:"coverUrl"`
		Genres   []string `json:"genres"`
		Tags     []string `json:"tags"`
		Status   string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = "ongoing"
	}
	if !validFictionStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fiction status"})
		return
	}
	f := &model.Fiction{
		AuthorID:   u.ID,
		AuthorName: u.Username,
		Title:      req.Title,
		Synopsis:   req.Synopsis,
		CoverURL:   req.CoverURL,
		Genres:     req.Genres,
		Tags:       req.Tags,
		Status:     req.Status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := store.CreateFiction(c.Request.Context(), f); err != nil {
		databaseError(c, "create fiction", err)
		return
	}
	c.JSON(http.StatusCreated, f)
}

func updateFictionHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	id := c.Param("id")

	f, err := store.GetFiction(c.Request.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fiction not found"})
		return
	}
	if err != nil {
		databaseError(c, "get fiction for update", err)
		return
	}
	if f.AuthorID != u.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your fiction"})
		return
	}

	var req struct {
		Title    string   `json:"title"`
		Synopsis string   `json:"synopsis"`
		CoverURL string   `json:"coverUrl"`
		Genres   []string `json:"genres"`
		Tags     []string `json:"tags"`
		Status   string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		f.Title = req.Title
	}
	if req.Synopsis != "" {
		f.Synopsis = req.Synopsis
	}
	if req.CoverURL != "" {
		f.CoverURL = req.CoverURL
	}
	if req.Genres != nil {
		f.Genres = req.Genres
		f.GenresWereOmitted = false
	}
	if req.Tags != nil {
		f.Tags = req.Tags
		f.TagsWereOmitted = false
	}
	if req.Status != "" {
		if !validFictionStatus(req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fiction status"})
			return
		}
		f.Status = req.Status
	}
	f.UpdatedAt = time.Now()

	if err := store.UpdateFiction(c.Request.Context(), f); err != nil {
		databaseError(c, "update fiction", err)
		return
	}
	c.JSON(http.StatusOK, f)
}

// ── Chapter handlers ──────────────────────────────────────────────────────────

func listChaptersHandler(c *gin.Context) {
	fictionID := c.Param("id")

	// Determine if the requester is the author (optional auth).
	var callerID string
	if header := c.GetHeader("Authorization"); strings.HasPrefix(header, "Bearer ") {
		if claims, err := parseToken(strings.TrimPrefix(header, "Bearer ")); err == nil {
			callerID = claims.UserID
		}
	}

	f, err := store.GetFiction(c.Request.Context(), fictionID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fiction not found"})
		return
	}
	if err != nil {
		databaseError(c, "get fiction chapter list", err)
		return
	}
	isAuthor := callerID != "" && callerID == f.AuthorID
	result, err := store.ListChapters(c.Request.Context(), fictionID, isAuthor)
	if err != nil {
		databaseError(c, "list chapters", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func getChapterHandler(c *gin.Context) {
	fictionID := c.Param("id")
	num, err := strconv.Atoi(c.Param("num"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter number"})
		return
	}

	// Allow the author to retrieve their own draft chapters.
	var callerID string
	if header := c.GetHeader("Authorization"); strings.HasPrefix(header, "Bearer ") {
		if claims, err := parseToken(strings.TrimPrefix(header, "Bearer ")); err == nil {
			callerID = claims.UserID
		}
	}

	f, err := store.GetFiction(c.Request.Context(), fictionID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fiction not found"})
		return
	}
	if err != nil {
		databaseError(c, "get fiction chapter", err)
		return
	}
	isAuthor := callerID != "" && callerID == f.AuthorID
	ch, err := store.GetChapter(c.Request.Context(), fictionID, num, isAuthor)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}
	if err != nil {
		databaseError(c, "get chapter", err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

func createChapterHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	fictionID := c.Param("id")

	f, err := store.GetFiction(c.Request.Context(), fictionID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fiction not found"})
		return
	}
	if err != nil {
		databaseError(c, "get fiction for chapter creation", err)
		return
	}
	if f.AuthorID != u.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your fiction"})
		return
	}

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
		Status  string `json:"status"` // draft | published
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if !validChapterStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter status"})
		return
	}

	ch := &model.Chapter{
		FictionID: fictionID,
		Title:     req.Title,
		Content:   req.Content,
		Status:    req.Status,
		CreatedAt: time.Now(),
	}
	if req.Status == "published" {
		ch.PublishedAt = ch.CreatedAt
	}
	if err := store.CreateChapter(c.Request.Context(), ch); err != nil {
		databaseError(c, "create chapter", err)
		return
	}
	c.JSON(http.StatusCreated, ch)
}

func updateChapterHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	fictionID := c.Param("id")
	num, err := strconv.Atoi(c.Param("num"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter number"})
		return
	}

	f, err := store.GetFiction(c.Request.Context(), fictionID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fiction not found"})
		return
	}
	if err != nil {
		databaseError(c, "get fiction for chapter update", err)
		return
	}
	if f.AuthorID != u.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your fiction"})
		return
	}

	target, err := store.GetChapter(c.Request.Context(), fictionID, num, true)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}
	if err != nil {
		databaseError(c, "get chapter for update", err)
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Title != "" {
		target.Title = req.Title
	}
	if req.Content != "" {
		target.Content = req.Content
	}
	if req.Status != "" {
		if !validChapterStatus(req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter status"})
			return
		}
		target.Status = req.Status
	}
	if err := store.UpdateChapter(c.Request.Context(), target); err != nil {
		databaseError(c, "update chapter", err)
		return
	}
	c.JSON(http.StatusOK, target)
}

// ── Follow handlers ───────────────────────────────────────────────────────────

func followHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	fictionID := c.Param("id")

	_, err = store.GetFiction(c.Request.Context(), fictionID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fiction not found"})
		return
	}
	if err != nil {
		databaseError(c, "get fiction for follow", err)
		return
	}
	if err := store.Follow(c.Request.Context(), u.ID, fictionID); err != nil {
		databaseError(c, "follow fiction", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"following": true})
}

func unfollowHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	fictionID := c.Param("id")

	_, err = store.GetFiction(c.Request.Context(), fictionID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fiction not found"})
		return
	}
	if err != nil {
		databaseError(c, "get fiction for unfollow", err)
		return
	}
	if err := store.Unfollow(c.Request.Context(), u.ID, fictionID); err != nil {
		databaseError(c, "unfollow fiction", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"following": false})
}

// ── User handlers ─────────────────────────────────────────────────────────────

func publicProfileHandler(c *gin.Context) {
	username := c.Param("username")

	u, err := store.FindUserByName(c.Request.Context(), username)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		databaseError(c, "get public profile user", err)
		return
	}
	myFictions, err := store.ListFictionsByAuthor(c.Request.Context(), u.ID)
	if err != nil {
		databaseError(c, "list public profile fictions", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":     u,
		"fictions": myFictions,
	})
}

func meHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func myFictionsHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	result, err := store.ListFictionsByAuthor(c.Request.Context(), u.ID)
	if err != nil {
		databaseError(c, "list user fictions", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func myFollowsHandler(c *gin.Context) {
	u, err := getUser(c)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			databaseError(c, "get authenticated user", err)
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	result, err := store.ListFollowedFictions(c.Request.Context(), u.ID)
	if err != nil {
		databaseError(c, "list follows", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func followStatusHandler(c *gin.Context) {
	fictionID := c.Param("id")
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		c.JSON(http.StatusOK, gin.H{"following": false})
		return
	}
	claims, err := parseToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"following": false})
		return
	}
	following, err := store.IsFollowing(c.Request.Context(), claims.UserID, fictionID)
	if err != nil {
		databaseError(c, "get follow status", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"following": following})
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	pool, err := OpenPool(context.Background(), config)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	if err := db.Migrate(context.Background(), pool); err != nil {
		panic(err)
	}
	if err := db.Seed(context.Background(), pool, config.AppEnv); err != nil {
		panic(err)
	}
	store = db.NewStore(pool)

	router := gin.Default()
	router.Use(corsMiddleware())

	api := router.Group("/api")

	// Auth
	api.POST("/auth/register", registerHandler)
	api.POST("/auth/login", loginHandler)

	// Fictions (public reads)
	api.GET("/fictions", listFictionsHandler)
	api.GET("/fictions/:id", getFictionHandler)
	api.GET("/fictions/:id/chapters", listChaptersHandler)
	api.GET("/fictions/:id/chapters/:num", getChapterHandler)
	api.GET("/fictions/:id/follow/status", followStatusHandler)

	// Fictions (auth required)
	auth := api.Group("/", authRequired())
	auth.POST("/fictions", createFictionHandler)
	auth.PUT("/fictions/:id", updateFictionHandler)
	auth.POST("/fictions/:id/chapters", createChapterHandler)
	auth.PUT("/fictions/:id/chapters/:num", updateChapterHandler)
	auth.POST("/fictions/:id/follow", followHandler)
	auth.DELETE("/fictions/:id/follow", unfollowHandler)

	// User
	api.GET("/users/:username", publicProfileHandler)
	auth.GET("/user/me", meHandler)
	auth.GET("/user/fictions", myFictionsHandler)
	auth.GET("/user/follows", myFollowsHandler)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
