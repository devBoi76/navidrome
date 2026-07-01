package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddWorkToFts, downAddWorkToFts)
}

func upAddWorkToFts(ctx context.Context, tx *sql.Tx) error {
	// Drop existing triggers to avoid conflicts during schema migration
	triggers := []string{
		"media_file_fts_ai", "media_file_fts_ad", "media_file_fts_au",
	}
	for _, trigger := range triggers {
		_, err := tx.ExecContext(ctx, "DROP TRIGGER IF EXISTS "+trigger)
		if err != nil {
			return fmt.Errorf("dropping trigger %s: %w", trigger, err)
		}
	}

	// Drop and recreate media_file_fts with work column added
	_, err := tx.ExecContext(ctx, "DROP TABLE IF EXISTS media_file_fts")
	if err != nil {
		return fmt.Errorf("dropping media_file_fts: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE VIRTUAL TABLE IF NOT EXISTS media_file_fts USING fts5(
			title, album, artist, album_artist,
			sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
			disc_subtitle, work, search_participants, search_normalized,
			content='', content_rowid='rowid',
			tokenize='unicode61 remove_diacritics 2'
		)
	`)
	if err != nil {
		return fmt.Errorf("creating media_file_fts with work column: %w", err)
	}

	// Repopulate FTS table with work column
	_, err = tx.ExecContext(ctx, `
		INSERT INTO media_file_fts(rowid, title, album, artist, album_artist,
			sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
			disc_subtitle, work, search_participants, search_normalized)
		SELECT rowid, title, album, artist, album_artist,
			sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
			COALESCE(disc_subtitle, ''), COALESCE(work, ''),
			COALESCE(search_participants, ''), COALESCE(search_normalized, '')
		FROM media_file
	`)
	if err != nil {
		return fmt.Errorf("repopulating media_file_fts: %w", err)
	}

	// Recreate INSERT trigger
	_, err = tx.ExecContext(ctx, `
		CREATE TRIGGER media_file_fts_ai AFTER INSERT ON media_file BEGIN
			INSERT INTO media_file_fts(rowid, title, album, artist, album_artist,
				sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
				disc_subtitle, work, search_participants, search_normalized)
			VALUES (NEW.rowid, NEW.title, NEW.album, NEW.artist, NEW.album_artist,
				NEW.sort_title, NEW.sort_album_name, NEW.sort_artist_name, NEW.sort_album_artist_name,
				COALESCE(NEW.disc_subtitle, ''), COALESCE(NEW.work, ''),
				COALESCE(NEW.search_participants, ''), COALESCE(NEW.search_normalized, ''));
		END
	`)
	if err != nil {
		return fmt.Errorf("creating media_file_fts insert trigger: %w", err)
	}

	// Recreate DELETE trigger
	_, err = tx.ExecContext(ctx, `
		CREATE TRIGGER media_file_fts_ad AFTER DELETE ON media_file BEGIN
			INSERT INTO media_file_fts(media_file_fts, rowid, title, album, artist, album_artist,
				sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
				disc_subtitle, work, search_participants, search_normalized)
			VALUES ('delete', OLD.rowid, OLD.title, OLD.album, OLD.artist, OLD.album_artist,
				OLD.sort_title, OLD.sort_album_name, OLD.sort_artist_name, OLD.sort_album_artist_name,
				COALESCE(OLD.disc_subtitle, ''), COALESCE(OLD.work, ''),
				COALESCE(OLD.search_participants, ''), COALESCE(OLD.search_normalized, ''));
		END
	`)
	if err != nil {
		return fmt.Errorf("creating media_file_fts delete trigger: %w", err)
	}

	// Recreate UPDATE trigger
	_, err = tx.ExecContext(ctx, `
		CREATE TRIGGER media_file_fts_au AFTER UPDATE ON media_file
		WHEN
			OLD.title IS NOT NEW.title OR
			OLD.album IS NOT NEW.album OR
			OLD.artist IS NOT NEW.artist OR
			OLD.album_artist IS NOT NEW.album_artist OR
			OLD.sort_title IS NOT NEW.sort_title OR
			OLD.sort_album_name IS NOT NEW.sort_album_name OR
			OLD.sort_artist_name IS NOT NEW.sort_artist_name OR
			OLD.sort_album_artist_name IS NOT NEW.sort_album_artist_name OR
			OLD.disc_subtitle IS NOT NEW.disc_subtitle OR
			OLD.work IS NOT NEW.work OR
			OLD.search_participants IS NOT NEW.search_participants OR
			OLD.search_normalized IS NOT NEW.search_normalized
		BEGIN
			INSERT INTO media_file_fts(media_file_fts, rowid, title, album, artist, album_artist,
				sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
				disc_subtitle, work, search_participants, search_normalized)
			VALUES ('delete', OLD.rowid, OLD.title, OLD.album, OLD.artist, OLD.album_artist,
				OLD.sort_title, OLD.sort_album_name, OLD.sort_artist_name, OLD.sort_album_artist_name,
				COALESCE(OLD.disc_subtitle, ''), COALESCE(OLD.work, ''),
				COALESCE(OLD.search_participants, ''), COALESCE(OLD.search_normalized, ''));
			INSERT INTO media_file_fts(rowid, title, album, artist, album_artist,
				sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
				disc_subtitle, work, search_participants, search_normalized)
			VALUES (NEW.rowid, NEW.title, NEW.album, NEW.artist, NEW.album_artist,
				NEW.sort_title, NEW.sort_album_name, NEW.sort_artist_name, NEW.sort_album_artist_name,
				COALESCE(NEW.disc_subtitle, ''), COALESCE(NEW.work, ''),
				COALESCE(NEW.search_participants, ''), COALESCE(NEW.search_normalized, ''));
		END
	`)
	if err != nil {
		return fmt.Errorf("creating media_file_fts update trigger: %w", err)
	}

	return nil
}

func downAddWorkToFts(ctx context.Context, tx *sql.Tx) error {
	// Drop existing triggers
	triggers := []string{
		"media_file_fts_ai", "media_file_fts_ad", "media_file_fts_au",
	}
	for _, trigger := range triggers {
		_, err := tx.ExecContext(ctx, "DROP TRIGGER IF EXISTS "+trigger)
		if err != nil {
			return fmt.Errorf("dropping trigger %s: %w", trigger, err)
		}
	}

	// Drop media_file_fts
	_, err := tx.ExecContext(ctx, "DROP TABLE IF EXISTS media_file_fts")
	if err != nil {
		return fmt.Errorf("dropping media_file_fts: %w", err)
	}

	// Recreate original media_file_fts without work column
	_, err = tx.ExecContext(ctx, `
		CREATE VIRTUAL TABLE IF NOT EXISTS media_file_fts USING fts5(
			title, album, artist, album_artist,
			sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
			disc_subtitle, search_participants, search_normalized,
			content='', content_rowid='rowid',
			tokenize='unicode61 remove_diacritics 2'
		)
	`)
	if err != nil {
		return fmt.Errorf("recreating media_file_fts: %w", err)
	}

	// Repopulate without work column
	_, err = tx.ExecContext(ctx, `
		INSERT INTO media_file_fts(rowid, title, album, artist, album_artist,
			sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
			disc_subtitle, search_participants, search_normalized)
		SELECT rowid, title, album, artist, album_artist,
			sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
			COALESCE(disc_subtitle, ''), COALESCE(search_participants, ''),
			COALESCE(search_normalized, '')
		FROM media_file
	`)
	if err != nil {
		return fmt.Errorf("repopulating media_file_fts: %w", err)
	}

	// Recreate triggers
	_, err = tx.ExecContext(ctx, `
		CREATE TRIGGER media_file_fts_ai AFTER INSERT ON media_file BEGIN
			INSERT INTO media_file_fts(rowid, title, album, artist, album_artist,
				sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
				disc_subtitle, search_participants, search_normalized)
			VALUES (NEW.rowid, NEW.title, NEW.album, NEW.artist, NEW.album_artist,
				NEW.sort_title, NEW.sort_album_name, NEW.sort_artist_name, NEW.sort_album_artist_name,
				COALESCE(NEW.disc_subtitle, ''), COALESCE(NEW.search_participants, ''),
				COALESCE(NEW.search_normalized, ''));
		END
	`)
	if err != nil {
		return fmt.Errorf("creating media_file_fts insert trigger: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE TRIGGER media_file_fts_ad AFTER DELETE ON media_file BEGIN
			INSERT INTO media_file_fts(media_file_fts, rowid, title, album, artist, album_artist,
				sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
				disc_subtitle, search_participants, search_normalized)
			VALUES ('delete', OLD.rowid, OLD.title, OLD.album, OLD.artist, OLD.album_artist,
				OLD.sort_title, OLD.sort_album_name, OLD.sort_artist_name, OLD.sort_album_artist_name,
				COALESCE(OLD.disc_subtitle, ''), COALESCE(OLD.search_participants, ''),
				COALESCE(OLD.search_normalized, ''));
		END
	`)
	if err != nil {
		return fmt.Errorf("creating media_file_fts delete trigger: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE TRIGGER media_file_fts_au AFTER UPDATE ON media_file
		WHEN
			OLD.title IS NOT NEW.title OR
			OLD.album IS NOT NEW.album OR
			OLD.artist IS NOT NEW.artist OR
			OLD.album_artist IS NOT NEW.album_artist OR
			OLD.sort_title IS NOT NEW.sort_title OR
			OLD.sort_album_name IS NOT NEW.sort_album_name OR
			OLD.sort_artist_name IS NOT NEW.sort_artist_name OR
			OLD.sort_album_artist_name IS NOT NEW.sort_album_artist_name OR
			OLD.disc_subtitle IS NOT NEW.disc_subtitle OR
			OLD.search_participants IS NOT NEW.search_participants OR
			OLD.search_normalized IS NOT NEW.search_normalized
		BEGIN
			INSERT INTO media_file_fts(media_file_fts, rowid, title, album, artist, album_artist,
				sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
				disc_subtitle, search_participants, search_normalized)
			VALUES ('delete', OLD.rowid, OLD.title, OLD.album, OLD.artist, OLD.album_artist,
				OLD.sort_title, OLD.sort_album_name, OLD.sort_artist_name, OLD.sort_album_artist_name,
				COALESCE(OLD.disc_subtitle, ''), COALESCE(OLD.search_participants, ''),
				COALESCE(OLD.search_normalized, ''));
			INSERT INTO media_file_fts(rowid, title, album, artist, album_artist,
				sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
				disc_subtitle, search_participants, search_normalized)
			VALUES (NEW.rowid, NEW.title, NEW.album, NEW.artist, NEW.album_artist,
				NEW.sort_title, NEW.sort_album_name, NEW.sort_artist_name, NEW.sort_album_artist_name,
				COALESCE(NEW.disc_subtitle, ''), COALESCE(NEW.search_participants, ''),
				COALESCE(NEW.search_normalized, ''));
		END
	`)
	if err != nil {
		return fmt.Errorf("creating media_file_fts update trigger: %w", err)
	}

	return nil
}
