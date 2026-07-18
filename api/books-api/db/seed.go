package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Seed(ctx context.Context, pool *pgxpool.Pool, appEnv string) error {
	if appEnv == "production" {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var usersExist bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users)`).Scan(&usersExist); err != nil {
		return fmt.Errorf("check existing users: %w", err)
	}
	if usersExist {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}
	now := time.Now()
	users := []struct {
		id, username, email, bio string
		createdAt                time.Time
	}{
		{"u1", "elara_writes", "elara@example.com", "Fantasy and sci-fi author. Lover of dragons and space.", now.Add(-90 * 24 * time.Hour)},
		{"u2", "nightquill", "night@example.com", "Horror and thriller writer. Writes best at 3am.", now.Add(-60 * 24 * time.Hour)},
	}
	for _, user := range users {
		_, err = tx.Exec(ctx, `INSERT INTO users (id, username, email, password_hash, bio, avatar_url, created_at) VALUES ($1, $2, $3, $4, $5, '', $6)`, user.id, user.username, user.email, string(hash), user.bio, user.createdAt)
		if err != nil {
			return fmt.Errorf("insert demo user: %w", err)
		}
	}

	fictions := []struct {
		id, authorID, authorName, title, synopsis string
		genres, tags                              []string
		createdAt, updatedAt                      time.Time
		followers, views, chapters                int
	}{
		{"f1", "u1", "elara_writes", "The Last Arcanist", "In a world where magic has been outlawed for centuries, Mira discovers she is the last person born with arcane power. Hunted by the Inquisition and sought by rebels, she must choose between hiding in safety or embracing her destiny.", []string{"Fantasy", "Adventure"}, []string{"magic", "chosen one", "political intrigue"}, now.Add(-80 * 24 * time.Hour), now.Add(-2 * 24 * time.Hour), 142, 3871, 3},
		{"f2", "u1", "elara_writes", "Starfall Station", "A derelict space station on the edge of a dying star holds the last archive of human knowledge. Three unlikely strangers must work together to retrieve it before the star goes supernova.", []string{"Sci-Fi", "Adventure"}, []string{"space", "found family", "survival"}, now.Add(-40 * 24 * time.Hour), now.Add(-5 * 24 * time.Hour), 89, 2104, 2},
		{"f3", "u2", "nightquill", "The Hollow Hours", "Every night at 3am, the town of Marrow Falls becomes a different place. Detective Elena Cross is the only one who remembers what happens in the hollow hours — and she's starting to suspect she may be the reason it's happening.", []string{"Horror", "Mystery"}, []string{"psychological", "small town", "supernatural"}, now.Add(-55 * 24 * time.Hour), now.Add(-1 * 24 * time.Hour), 201, 5430, 2},
	}
	for _, fiction := range fictions {
		_, err = tx.Exec(ctx, `INSERT INTO fictions (id, author_id, author_name, title, synopsis, genres, tags, status, created_at, updated_at, follower_count, view_count, chapter_count) VALUES ($1, $2, $3, $4, $5, $6, $7, 'ongoing', $8, $9, $10, $11, $12)`, fiction.id, fiction.authorID, fiction.authorName, fiction.title, fiction.synopsis, fiction.genres, fiction.tags, fiction.createdAt, fiction.updatedAt, fiction.followers, fiction.views, fiction.chapters)
		if err != nil {
			return fmt.Errorf("insert demo fiction: %w", err)
		}
	}

	chapters := []struct {
		id, fictionID, title, content string
		number                        int
		publishedAt                   time.Time
	}{
		{"c1", "f1", "Prologue: The Burning Library", "The night the Grand Library burned, Mira was the only witness. She had snuck in after hours to return a stolen book — a book that, as it turned out, she should never have been able to read.\n\nThe fire started in the restricted section. She saw no torch, no lantern. Only a soft blue glow that rapidly became an inferno. The books screamed — that is the only way she could describe it — as their pages curled and blackened.\n\nShe ran. But not before pocketing the small silver compass that had been lying open on the reading table, spinning wildly though there was no magnetic north for a hundred miles.", 1, now.Add(-79 * 24 * time.Hour)},
		{"c2", "f1", "Chapter 1: The Inquisitor's Visit", "Three months after the fire, the Inquisitor arrived in Millhaven.\n\nMira watched from the bakery window as the black carriage rolled down the cobblestone street. No insignia. No escort. That made it worse, somehow — the ones who needed no announcement were the ones with nothing to prove.\n\n\"Back to work,\" said Aunt Renee, not looking up from the bread she was shaping. \"And stop biting your nails.\"\n\nMira pulled her hand from her mouth. She had been doing it again — that nervous habit that had gotten worse since the library. Since the night she had felt something inside her chest unlock like a door she hadn't known was there.", 2, now.Add(-70 * 24 * time.Hour)},
		{"c3", "f1", "Chapter 2: What the Compass Shows", "The compass did not point north. It pointed at people.\n\nMira had figured this out on the walk home from the library, the night of the fire. It spun uselessly when she held it in an empty room. But the moment someone walked in, the needle snapped toward them like an accusation.\n\nShe thought it pointed at heat, at first. Then at heartbeats. It took her two weeks to understand the truth.\n\nIt pointed at magic.", 3, now.Add(-65 * 24 * time.Hour)},
		{"c4", "f2", "Arrival", "The shuttle docked with a sound like a body hitting the floor.\n\nKael had worked salvage for eleven years and he had never heard a station make that sound before. Stations groaned. They hissed. They occasionally shrieked when something critical failed. They did not thud.\n\n\"Structural settling,\" said the station's automated voice, as if it had heard him think. \"Please proceed to the welcome bay.\"\n\nThe welcome bay, it turned out, was a room containing three things: a broken vending machine, two other salvagers who looked as suspicious of him as he was of them, and a single laminated card taped to the wall that read: DO NOT GO BELOW DECK 7.", 1, now.Add(-38 * 24 * time.Hour)},
		{"c5", "f2", "Deck 7", "They went to deck 7 within the hour.\n\nIt wasn't stubbornness, exactly. Or rather, it was stubbornness, but the rational kind. Kael had learned long ago that signs saying DO NOT were essentially maps to wherever the money was. The other two — a woman named Sable who claimed to be a data archaeologist, and a kid who refused to give a name and was therefore called 'the kid' — had reached the same conclusion independently.\n\n\"We're not a team,\" Sable said, as they all stepped into the elevator together.\n\n\"Obviously,\" said Kael.\n\n\"Obviously,\" said the kid.", 2, now.Add(-30 * 24 * time.Hour)},
		{"c6", "f3", "3:00 AM, Tuesday", "The call came in at 2:58, which meant Elena had two minutes to finish her coffee and get to the scene before it became something else entirely.\n\nMarrow Falls was a small enough town that 'the scene' was never far. She was there in four minutes. She was late.\n\nThe Hendersons' living room had rearranged itself. That was the only way to put it. The furniture was in the same positions — couch against the east wall, television opposite, armchair by the window — but the distances between them had changed. The room was larger. Or the furniture was smaller. Or something in between that Elena didn't have the geometry for.\n\nMr. Henderson stood in the center of it, looking at his hands.\n\n\"It happened again,\" he said.", 1, now.Add(-53 * 24 * time.Hour)},
		{"c7", "f3", "The Only One Who Remembers", "By morning, it was gone.\n\nThe room was back to its normal dimensions. Mr. Henderson remembered nothing. His wife, asleep upstairs the whole time, had a vague feeling she'd had a strange dream but couldn't say what it was.\n\nElena sat in her cruiser outside their house and wrote in her notebook: *Third incident this month. Third time I'm the only one who remembers.*\n\nShe had started keeping the notebook six months ago, after the first hollow hour. She had initially written it off as stress, sleep deprivation, the particular way that small-town detective work could make a person feel untethered from consensus reality.\n\nThen she started noticing that the incidents always centered on her location.", 2, now.Add(-48 * 24 * time.Hour)},
	}
	for _, chapter := range chapters {
		_, err = tx.Exec(ctx, `INSERT INTO chapters (id, fiction_id, title, content, chapter_number, status, published_at, created_at) VALUES ($1, $2, $3, $4, $5, 'published', $6, $6)`, chapter.id, chapter.fictionID, chapter.title, chapter.content, chapter.number, chapter.publishedAt)
		if err != nil {
			return fmt.Errorf("insert demo chapter: %w", err)
		}
	}

	for _, sequence := range []struct {
		name  string
		value int
	}{{"user_id_sequence", 2}, {"fiction_id_sequence", 3}, {"chapter_id_sequence", 7}} {
		if _, err := tx.Exec(ctx, `SELECT setval($1::regclass, $2, true)`, sequence.name, sequence.value); err != nil {
			return fmt.Errorf("position %s: %w", sequence.name, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}
	return nil
}
