package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guhaag/writerslife-books-api/db"
	"github.com/guhaag/writerslife-books-api/model"
	"github.com/stretchr/testify/require"
)

func integrationTestConfig() (Config, bool, error) {
	if os.Getenv("DATABASE_URL") == "" {
		return Config{}, true, nil
	}
	config, err := LoadConfig()
	if err != nil {
		return Config{}, false, err
	}
	if err := RequireTestDatabase(config); err != nil {
		return Config{}, false, err
	}
	return config, false, nil
}

func setupIntegrationStore(t *testing.T) (context.Context, *db.Store) {
	t.Helper()
	config, skipped, err := integrationTestConfig()
	if skipped {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	require.NoError(t, err)
	ctx := context.Background()
	pool, err := OpenPool(ctx, config)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, db.Migrate(ctx, pool))
	_, err = pool.Exec(ctx, `TRUNCATE users CASCADE`)
	require.NoError(t, err)
	require.NoError(t, db.Seed(ctx, pool, "test"))
	store = db.NewStore(pool)
	return ctx, store
}

func TestLiteralSearchAndPersistedChapterFollowCounters(t *testing.T) {
	ctx, testStore := setupIntegrationStore(t)
	author := &model.User{Username: "symbols", Email: "symbols@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	follower := &model.User{Username: "follower", Email: "follower@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	require.NoError(t, testStore.CreateUser(ctx, author))
	require.NoError(t, testStore.CreateUser(ctx, follower))
	fiction := &model.Fiction{AuthorID: author.ID, AuthorName: author.Username, Title: `100%_\\`, Genres: []string{}, Tags: []string{}, Status: "ongoing", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, testStore.CreateFiction(ctx, fiction))
	for _, query := range []string{"%", "_", `\\`} {
		results, err := testStore.ListFictions(ctx, query, "", "recent")
		require.NoError(t, err)
		require.Contains(t, results, fiction)
	}
	draft := &model.Chapter{FictionID: fiction.ID, Title: "Draft", Status: "draft", CreatedAt: time.Now()}
	require.NoError(t, testStore.CreateChapter(ctx, draft))
	public, err := testStore.ListChapters(ctx, fiction.ID, false)
	require.NoError(t, err)
	require.Empty(t, public)
	draft.Status = "published"
	require.NoError(t, testStore.UpdateChapter(ctx, draft))
	require.NoError(t, testStore.UpdateChapter(ctx, draft))
	updated, err := testStore.GetFiction(ctx, fiction.ID)
	require.NoError(t, err)
	require.Equal(t, 1, updated.ChapterCount)
	require.NoError(t, testStore.Follow(ctx, follower.ID, fiction.ID))
	require.NoError(t, testStore.Follow(ctx, follower.ID, fiction.ID))
	require.NoError(t, testStore.Unfollow(ctx, follower.ID, fiction.ID))
	require.NoError(t, testStore.Unfollow(ctx, follower.ID, fiction.ID))
	updated, err = testStore.GetFiction(ctx, fiction.ID)
	require.NoError(t, err)
	require.Zero(t, updated.FollowerCount)
}

func TestCreateFictionWithoutGenresAndTags(t *testing.T) {
	ctx, testStore := setupIntegrationStore(t)
	author := &model.User{Username: "arrayless_author", Email: "arrayless@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	require.NoError(t, testStore.CreateUser(ctx, author))
	token, err := makeToken(author)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Group("/api", authRequired()).POST("/fictions", createFictionHandler)

	request := httptest.NewRequest(http.MethodPost, "/api/fictions", bytes.NewBufferString(`{"title":"No Arrays"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusCreated, response.Code)
	var fiction model.Fiction
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &fiction))
	require.Nil(t, fiction.Genres)
	require.Nil(t, fiction.Tags)
	stored, err := testStore.GetFiction(ctx, fiction.ID)
	require.NoError(t, err)
	require.Nil(t, stored.Genres)
	require.Nil(t, stored.Tags)

	config, _, err := integrationTestConfig()
	require.NoError(t, err)
	pool, err := OpenPool(ctx, config)
	require.NoError(t, err)
	defer pool.Close()
	var genres, tags []string
	require.NoError(t, pool.QueryRow(ctx, `SELECT genres, tags FROM fictions WHERE id = $1`, fiction.ID).Scan(&genres, &tags))
	require.NotNil(t, genres)
	require.NotNil(t, tags)
}

func TestFictionArrayOmissionPersistsAcrossHTTPReads(t *testing.T) {
	ctx, testStore := setupIntegrationStore(t)
	author := &model.User{Username: "array_http_author", Email: "array-http@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	require.NoError(t, testStore.CreateUser(ctx, author))
	token, err := makeToken(author)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/fictions", listFictionsHandler)
	router.GET("/api/fictions/:id", getFictionHandler)
	router.GET("/api/users/:username", publicProfileHandler)
	auth := router.Group("/api", authRequired())
	auth.POST("/fictions", createFictionHandler)
	auth.PUT("/fictions/:id", updateFictionHandler)
	auth.GET("/user/me", meHandler)
	auth.GET("/user/fictions", myFictionsHandler)

	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		if method != http.MethodGet || path == "/api/user/me" || path == "/api/user/fictions" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}

	created := request(http.MethodPost, "/api/fictions", `{"title":"Omitted arrays"}`)
	require.Equal(t, http.StatusCreated, created.Code)
	var omitted model.Fiction
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &omitted))
	require.Nil(t, omitted.Genres)
	require.Nil(t, omitted.Tags)

	empty := request(http.MethodPost, "/api/fictions", `{"title":"Empty arrays","genres":[],"tags":[]}`)
	require.Equal(t, http.StatusCreated, empty.Code)
	var explicitEmpty model.Fiction
	require.NoError(t, json.Unmarshal(empty.Body.Bytes(), &explicitEmpty))
	require.NotNil(t, explicitEmpty.Genres)
	require.NotNil(t, explicitEmpty.Tags)

	assertFictionArrays := func(body []byte, expected *model.Fiction) {
		var fiction model.Fiction
		require.NoError(t, json.Unmarshal(body, &fiction))
		require.Equal(t, expected.ID, fiction.ID)
		if expected.ID == omitted.ID {
			require.Nil(t, fiction.Genres)
			require.Nil(t, fiction.Tags)
			return
		}
		require.NotNil(t, fiction.Genres)
		require.NotNil(t, fiction.Tags)
	}

	viewed := request(http.MethodGet, "/api/fictions/"+omitted.ID, "")
	require.Equal(t, http.StatusOK, viewed.Code)
	assertFictionArrays(viewed.Body.Bytes(), &omitted)

	updated := request(http.MethodPut, "/api/fictions/"+omitted.ID, `{"title":"Still omitted"}`)
	require.Equal(t, http.StatusOK, updated.Code)
	assertFictionArrays(updated.Body.Bytes(), &omitted)

	listed := request(http.MethodGet, "/api/fictions?search=omitted", "")
	require.Equal(t, http.StatusOK, listed.Code)
	var fictions []model.Fiction
	require.NoError(t, json.Unmarshal(listed.Body.Bytes(), &fictions))
	require.Len(t, fictions, 1)
	require.Nil(t, fictions[0].Genres)
	require.Nil(t, fictions[0].Tags)

	profile := request(http.MethodGet, "/api/users/"+author.Username, "")
	require.Equal(t, http.StatusOK, profile.Code)
	var profileResponse struct {
		Fictions []model.Fiction `json:"fictions"`
	}
	require.NoError(t, json.Unmarshal(profile.Body.Bytes(), &profileResponse))
	require.Len(t, profileResponse.Fictions, 2)
	require.Nil(t, profileResponse.Fictions[0].Genres)
	require.Nil(t, profileResponse.Fictions[0].Tags)
	require.NotNil(t, profileResponse.Fictions[1].Genres)
	require.NotNil(t, profileResponse.Fictions[1].Tags)

	me := request(http.MethodGet, "/api/user/me", "")
	require.Equal(t, http.StatusOK, me.Code)

	mine := request(http.MethodGet, "/api/user/fictions", "")
	require.Equal(t, http.StatusOK, mine.Code)
	require.NoError(t, json.Unmarshal(mine.Body.Bytes(), &fictions))
	require.Len(t, fictions, 2)
	require.Nil(t, fictions[0].Genres)
	require.Nil(t, fictions[0].Tags)
	require.NotNil(t, fictions[1].Genres)
	require.NotNil(t, fictions[1].Tags)
}

func TestIntegrationTestConfigDoesNotSkipUnsafeDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://writerslife:writerslife@localhost:5432/writerslife_dev?sslmode=disable")
	t.Setenv("APP_ENV", "development")

	_, skipped, err := integrationTestConfig()
	require.False(t, skipped)
	require.Error(t, err)
}

func TestIntegrationDatabaseReachabilityAndMigrations(t *testing.T) {
	config, skipped, err := integrationTestConfig()
	if skipped {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	require.NoError(t, err)

	pool, err := OpenPool(context.Background(), config)
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, db.Migrate(context.Background(), pool))
}

func TestMigrationSchema(t *testing.T) {
	config, skipped, err := integrationTestConfig()
	if skipped {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	require.NoError(t, err)

	ctx := context.Background()
	pool, err := OpenPool(ctx, config)
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, db.Migrate(ctx, pool))

	_, err = pool.Exec(ctx, `DELETE FROM fictions WHERE id = 'f1'`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `DELETE FROM users WHERE id IN ('u98', 'u99')`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO users (id, username, email, password_hash, created_at) VALUES ('u98', 'author', 'author@example.com', 'hash', now())`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO users (id, username, email, password_hash, created_at) VALUES ('u99', 'author', 'new@example.com', 'hash', now())`)
	require.Error(t, err)

	_, err = pool.Exec(ctx, `INSERT INTO fictions (id, author_id, author_name, title, created_at, updated_at) VALUES ('f1', 'u98', 'author', 'Test fiction', now(), now())`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO chapters (id, fiction_id, title, content, chapter_number, status, created_at) VALUES ('c98', 'f1', 'Original', '', 1, 'draft', now())`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO chapters (id, fiction_id, title, content, chapter_number, status, created_at) VALUES ('c99', 'f1', 'Duplicate', '', 1, 'draft', now())`)
	require.Error(t, err)
}

func TestRequireTestDatabaseRejectsUnsafeConfig(t *testing.T) {
	unsafeConfigs := []Config{
		{AppEnv: "development", DatabaseURL: "postgres://writerslife:writerslife@localhost:5432/writerslife_test?sslmode=disable"},
		{AppEnv: "test", DatabaseURL: "postgres://writerslife:writerslife@localhost:5432/writerslife_dev?sslmode=disable"},
		{AppEnv: "test", DatabaseURL: "postgres://writerslife:writerslife@localhost:5432/writerslife_test_extra?sslmode=disable"},
		{AppEnv: "test", DatabaseURL: "mailto:/writerslife_test"},
	}

	for _, config := range unsafeConfigs {
		require.Error(t, RequireTestDatabase(config))
	}

	require.NoError(t, RequireTestDatabase(Config{
		AppEnv:      "test",
		DatabaseURL: "postgres://writerslife:writerslife@localhost:5432/writerslife_test?sslmode=disable",
	}))
}

func TestSeedIsIdempotentAndDisabledInProduction(t *testing.T) {
	config, skipped, err := integrationTestConfig()
	if skipped {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	require.NoError(t, err)

	ctx := context.Background()
	pool, err := OpenPool(ctx, config)
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, db.Migrate(ctx, pool))
	_, err = pool.Exec(ctx, `TRUNCATE users CASCADE`)
	require.NoError(t, err)

	require.NoError(t, db.Seed(ctx, pool, "test"))
	require.NoError(t, db.Seed(ctx, pool, "test"))

	var userCount int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&userCount))
	require.Equal(t, 2, userCount)

	_, err = pool.Exec(ctx, `TRUNCATE users CASCADE`)
	require.NoError(t, err)
	require.NoError(t, db.Seed(ctx, pool, "production"))
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&userCount))
	require.Zero(t, userCount)
}

func TestRegisterAndLoginPersistAcrossFreshApplicationState(t *testing.T) {
	config, skipped, err := integrationTestConfig()
	if skipped {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	require.NoError(t, err)

	ctx := context.Background()
	pool, err := OpenPool(ctx, config)
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, db.Migrate(ctx, pool))
	_, err = pool.Exec(ctx, `TRUNCATE users CASCADE`)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/auth/register", registerHandler)
	router.POST("/api/auth/login", loginHandler)

	store = db.NewStore(pool)

	registerBody := []byte(`{"username":"persisted_author","email":"persisted@example.com","password":"secret"}`)
	registerRequest := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerBody))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerResponse := httptest.NewRecorder()
	router.ServeHTTP(registerResponse, registerRequest)
	require.Equal(t, http.StatusCreated, registerResponse.Code)

	duplicateResponse := httptest.NewRecorder()
	duplicateRequest := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerBody))
	duplicateRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(duplicateResponse, duplicateRequest)
	require.Equal(t, http.StatusConflict, duplicateResponse.Code)

	duplicateEmailRequest := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{"username":"another_author","email":"persisted@example.com","password":"secret"}`))
	duplicateEmailRequest.Header.Set("Content-Type", "application/json")
	duplicateEmailResponse := httptest.NewRecorder()
	router.ServeHTTP(duplicateEmailResponse, duplicateEmailRequest)
	require.Equal(t, http.StatusConflict, duplicateEmailResponse.Code)

	var userCount int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE username = 'persisted_author'`).Scan(&userCount))
	require.Equal(t, 1, userCount)

	pool.Close()
	freshPool, err := OpenPool(ctx, config)
	require.NoError(t, err)
	defer freshPool.Close()
	store = db.NewStore(freshPool)

	loginBody, err := json.Marshal(map[string]string{"username": "persisted_author", "password": "secret"})
	require.NoError(t, err)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, loginRequest)
	require.Equal(t, http.StatusOK, loginResponse.Code)
}

func TestAuthenticationDatabaseFailuresReturnGenericInternalError(t *testing.T) {
	config, skipped, err := integrationTestConfig()
	if skipped {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	require.NoError(t, err)
	pool, err := OpenPool(context.Background(), config)
	require.NoError(t, err)
	store = db.NewStore(pool)
	pool.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/auth/register", registerHandler)
	router.POST("/api/auth/login", loginHandler)
	for _, testCase := range []struct {
		endpoint string
		body     string
	}{
		{"/api/auth/register", `{"username":"unavailable","email":"unavailable@example.com","password":"secret"}`},
		{"/api/auth/login", `{"username":"unavailable","password":"secret"}`},
	} {
		request := httptest.NewRequest(http.MethodPost, testCase.endpoint, bytes.NewBufferString(testCase.body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.JSONEq(t, `{"error":"internal server error"}`, response.Body.String())
	}
}

func TestFictionHTTPLiteralSearchUpdateCountersAndOwnership(t *testing.T) {
	config, skipped, err := integrationTestConfig()
	if skipped {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	require.NoError(t, err)
	ctx := context.Background()
	pool, err := OpenPool(ctx, config)
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, db.Migrate(ctx, pool))
	_, err = pool.Exec(ctx, `TRUNCATE users CASCADE`)
	require.NoError(t, err)
	store = db.NewStore(pool)

	owner := &model.User{Username: "owner", Email: "owner@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	other := &model.User{Username: "other", Email: "other@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	require.NoError(t, store.CreateUser(ctx, owner))
	require.NoError(t, store.CreateUser(ctx, other))
	ownerToken, err := makeToken(owner)
	require.NoError(t, err)
	otherToken, err := makeToken(other)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/fictions", listFictionsHandler)
	router.GET("/api/fictions/:id", getFictionHandler)
	auth := router.Group("/api", authRequired())
	auth.POST("/fictions", createFictionHandler)
	auth.PUT("/fictions/:id", updateFictionHandler)

	create := httptest.NewRequest(http.MethodPost, "/api/fictions", bytes.NewBufferString(`{"title":"The Query 100%_\\","synopsis":"A PostgreSQL mystery","genres":["Mystery"]}`))
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("Authorization", "Bearer "+ownerToken)
	created := httptest.NewRecorder()
	router.ServeHTTP(created, create)
	require.Equal(t, http.StatusCreated, created.Code)
	var fiction model.Fiction
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &fiction))
	require.Regexp(t, `^f[0-9]+$`, fiction.ID)

	for _, search := range []string{"%", "_", `\`} {
		listed := httptest.NewRecorder()
		router.ServeHTTP(listed, httptest.NewRequest(http.MethodGet, "/api/fictions?search="+url.QueryEscape(search)+"&genre=mystery", nil))
		require.Equal(t, http.StatusOK, listed.Code)
		require.Contains(t, listed.Body.String(), fiction.ID)
	}

	viewed := httptest.NewRecorder()
	router.ServeHTTP(viewed, httptest.NewRequest(http.MethodGet, "/api/fictions/"+fiction.ID, nil))
	require.Equal(t, http.StatusOK, viewed.Code)
	require.Contains(t, viewed.Body.String(), `"viewCount":1`)

	published := &model.Chapter{FictionID: fiction.ID, Title: "Published", Status: "published", CreatedAt: time.Now()}
	require.NoError(t, store.CreateChapter(ctx, published))
	beforeUpdate, err := store.GetFiction(ctx, fiction.ID)
	require.NoError(t, err)
	require.Equal(t, 1, beforeUpdate.ViewCount)
	require.Equal(t, 1, beforeUpdate.ChapterCount)

	ownerUpdate := httptest.NewRequest(http.MethodPut, "/api/fictions/"+fiction.ID, bytes.NewBufferString(`{"title":"The Updated Query"}`))
	ownerUpdate.Header.Set("Content-Type", "application/json")
	ownerUpdate.Header.Set("Authorization", "Bearer "+ownerToken)
	ownerUpdated := httptest.NewRecorder()
	router.ServeHTTP(ownerUpdated, ownerUpdate)
	require.Equal(t, http.StatusOK, ownerUpdated.Code)
	var updated model.Fiction
	require.NoError(t, json.Unmarshal(ownerUpdated.Body.Bytes(), &updated))
	require.Equal(t, beforeUpdate.ViewCount, updated.ViewCount)
	require.Equal(t, beforeUpdate.FollowerCount, updated.FollowerCount)
	require.Equal(t, beforeUpdate.ChapterCount, updated.ChapterCount)

	update := httptest.NewRequest(http.MethodPut, "/api/fictions/"+fiction.ID, bytes.NewBufferString(`{"title":"Stolen"}`))
	update.Header.Set("Content-Type", "application/json")
	update.Header.Set("Authorization", "Bearer "+otherToken)
	forbidden := httptest.NewRecorder()
	router.ServeHTTP(forbidden, update)
	require.Equal(t, http.StatusForbidden, forbidden.Code)
}

func TestChapterHTTPVisibilityNumberingOwnershipAndStatusCounters(t *testing.T) {
	ctx, testStore := setupIntegrationStore(t)
	owner := &model.User{Username: "chapter_owner", Email: "chapter-owner@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	other := &model.User{Username: "chapter_other", Email: "chapter-other@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	require.NoError(t, testStore.CreateUser(ctx, owner))
	require.NoError(t, testStore.CreateUser(ctx, other))
	fiction := &model.Fiction{AuthorID: owner.ID, AuthorName: owner.Username, Title: "Chapter tests", Genres: []string{}, Tags: []string{}, Status: "ongoing", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, testStore.CreateFiction(ctx, fiction))
	ownerToken, err := makeToken(owner)
	require.NoError(t, err)
	otherToken, err := makeToken(other)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/fictions/:id/chapters", listChaptersHandler)
	router.GET("/api/fictions/:id/chapters/:num", getChapterHandler)
	auth := router.Group("/api", authRequired())
	auth.POST("/fictions/:id/chapters", createChapterHandler)
	auth.PUT("/fictions/:id/chapters/:num", updateChapterHandler)

	request := func(method, path, body, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}

	draft := request(http.MethodPost, "/api/fictions/"+fiction.ID+"/chapters", `{"title":"Draft"}`, ownerToken)
	require.Equal(t, http.StatusCreated, draft.Code)
	published := request(http.MethodPost, "/api/fictions/"+fiction.ID+"/chapters", `{"title":"Published","status":"published"}`, ownerToken)
	require.Equal(t, http.StatusCreated, published.Code)
	var draftChapter, publishedChapter model.Chapter
	require.NoError(t, json.Unmarshal(draft.Body.Bytes(), &draftChapter))
	require.NoError(t, json.Unmarshal(published.Body.Bytes(), &publishedChapter))
	require.Equal(t, 1, draftChapter.ChapterNumber)
	require.Equal(t, 2, publishedChapter.ChapterNumber)
	denied := request(http.MethodPost, "/api/fictions/"+fiction.ID+"/chapters", `{"title":"Stolen"}`, otherToken)
	require.Equal(t, http.StatusForbidden, denied.Code)

	publicList := request(http.MethodGet, "/api/fictions/"+fiction.ID+"/chapters", "", "")
	require.Equal(t, http.StatusOK, publicList.Code)
	require.NotContains(t, publicList.Body.String(), "Draft")
	require.Contains(t, publicList.Body.String(), "Published")
	authorList := request(http.MethodGet, "/api/fictions/"+fiction.ID+"/chapters", "", ownerToken)
	require.Contains(t, authorList.Body.String(), "Draft")
	publicDraft := request(http.MethodGet, "/api/fictions/"+fiction.ID+"/chapters/1", "", "")
	require.Equal(t, http.StatusNotFound, publicDraft.Code)
	authorDraft := request(http.MethodGet, "/api/fictions/"+fiction.ID+"/chapters/1", "", ownerToken)
	require.Equal(t, http.StatusOK, authorDraft.Code)

	require.Equal(t, http.StatusOK, request(http.MethodPut, "/api/fictions/"+fiction.ID+"/chapters/1", `{"status":"published"}`, ownerToken).Code)
	require.Equal(t, http.StatusOK, request(http.MethodPut, "/api/fictions/"+fiction.ID+"/chapters/1", `{"status":"draft"}`, ownerToken).Code)
	require.Equal(t, http.StatusOK, request(http.MethodPut, "/api/fictions/"+fiction.ID+"/chapters/1", `{"status":"published"}`, ownerToken).Code)
	updated, err := testStore.GetFiction(ctx, fiction.ID)
	require.NoError(t, err)
	require.Equal(t, 3, updated.ChapterCount)
}

func TestChapterNumbersAreSequentialDuringConcurrentCreates(t *testing.T) {
	ctx, testStore := setupIntegrationStore(t)
	author := &model.User{Username: "concurrent_author", Email: "concurrent@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	require.NoError(t, testStore.CreateUser(ctx, author))
	fiction := &model.Fiction{AuthorID: author.ID, AuthorName: author.Username, Title: "Concurrent chapters", Genres: []string{}, Tags: []string{}, Status: "ongoing", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, testStore.CreateFiction(ctx, fiction))

	errs := make(chan error, 2)
	for _, title := range []string{"One", "Two"} {
		go func(title string) {
			errs <- testStore.CreateChapter(context.Background(), &model.Chapter{FictionID: fiction.ID, Title: title, Status: "draft", CreatedAt: time.Now()})
		}(title)
	}
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
	chapters, err := testStore.ListChapters(ctx, fiction.ID, true)
	require.NoError(t, err)
	require.Len(t, chapters, 2)
	require.Equal(t, 1, chapters[0].ChapterNumber)
	require.Equal(t, 2, chapters[1].ChapterNumber)
}

func TestFollowHTTPStatusListAndCounters(t *testing.T) {
	ctx, testStore := setupIntegrationStore(t)
	author := &model.User{Username: "follow_author", Email: "follow-author@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	follower := &model.User{Username: "follow_user", Email: "follow-user@example.com", PasswordHash: "hash", CreatedAt: time.Now()}
	require.NoError(t, testStore.CreateUser(ctx, author))
	require.NoError(t, testStore.CreateUser(ctx, follower))
	fiction := &model.Fiction{AuthorID: author.ID, AuthorName: author.Username, Title: "Follow tests", Genres: []string{}, Tags: []string{}, Status: "ongoing", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, testStore.CreateFiction(ctx, fiction))
	token, err := makeToken(follower)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/fictions/:id/follow/status", followStatusHandler)
	auth := router.Group("/api", authRequired())
	auth.POST("/fictions/:id/follow", followHandler)
	auth.DELETE("/fictions/:id/follow", unfollowHandler)
	auth.GET("/user/follows", myFollowsHandler)
	request := func(method, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}

	require.Contains(t, request(http.MethodGet, "/api/fictions/"+fiction.ID+"/follow/status").Body.String(), `"following":false`)
	require.Equal(t, http.StatusOK, request(http.MethodPost, "/api/fictions/"+fiction.ID+"/follow").Code)
	require.Equal(t, http.StatusOK, request(http.MethodPost, "/api/fictions/"+fiction.ID+"/follow").Code)
	require.Contains(t, request(http.MethodGet, "/api/fictions/"+fiction.ID+"/follow/status").Body.String(), `"following":true`)
	require.Contains(t, request(http.MethodGet, "/api/user/follows").Body.String(), fiction.ID)
	require.Equal(t, http.StatusOK, request(http.MethodDelete, "/api/fictions/"+fiction.ID+"/follow").Code)
	require.Equal(t, http.StatusOK, request(http.MethodDelete, "/api/fictions/"+fiction.ID+"/follow").Code)
	updated, err := testStore.GetFiction(ctx, fiction.ID)
	require.NoError(t, err)
	require.Zero(t, updated.FollowerCount)
}
