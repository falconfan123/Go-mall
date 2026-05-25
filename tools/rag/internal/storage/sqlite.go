package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
)

type Store struct {
	db *sql.DB
}

type Session struct {
	ID              string
	Prompt          string
	Status          string
	RepoRoot        string
	WorktreePath    string
	Branch          string
	AllowedCommands []string
	Iteration       int
	MaxIterations   int
	DryRun          bool
	Summary         string
	CommitMessage   string
	LastError       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SessionEvent struct {
	Seq       int
	SessionID string
	Role      string
	Content   string
	CreatedAt time.Time
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	store := &Store{db: db}
	if err := store.init(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) init(ctx context.Context) error {
	stmts := []string{
		`PRAGMA busy_timeout=5000;`,
		`PRAGMA journal_mode=WAL;`,
		`CREATE TABLE IF NOT EXISTS meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			line_start INTEGER NOT NULL,
			line_end INTEGER NOT NULL,
			content TEXT NOT NULL,
			priority REAL NOT NULL,
			hash TEXT NOT NULL,
			source_kind TEXT NOT NULL DEFAULT '',
			source_name TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_path ON chunks(path);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			prompt TEXT NOT NULL,
			status TEXT NOT NULL,
			repo_root TEXT NOT NULL,
			worktree_path TEXT NOT NULL,
			branch TEXT NOT NULL,
			allowed_commands TEXT NOT NULL,
			iteration INTEGER NOT NULL,
			max_iterations INTEGER NOT NULL,
			dry_run INTEGER NOT NULL,
			summary TEXT NOT NULL,
			commit_message TEXT NOT NULL,
			last_error TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS session_events (
			seq INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_session_events_session ON session_events(session_id, seq);`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("init sqlite schema: %w", err)
		}
	}
	migrations := []string{
		`ALTER TABLE chunks ADD COLUMN source_kind TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE chunks ADD COLUMN source_name TEXT NOT NULL DEFAULT ''`,
	}
	for _, stmt := range migrations {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("migrate sqlite schema: %w", err)
		}
	}
	return nil
}

func (s *Store) ReplaceChunks(ctx context.Context, chunks []indexer.Chunk) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM chunks`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO chunks(path, line_start, line_end, content, priority, hash, source_kind, source_name) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		if _, err := stmt.ExecContext(ctx, chunk.Path, chunk.LineStart, chunk.LineEnd, chunk.Content, chunk.Priority, chunk.Hash, chunk.SourceKind, chunk.SourceName); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO meta(key, value) VALUES(?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, "last_indexed_at", time.Now().UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) LoadChunks(ctx context.Context) ([]indexer.Chunk, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, path, line_start, line_end, content, priority, hash, source_kind, source_name FROM chunks ORDER BY path, line_start`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chunks []indexer.Chunk
	for rows.Next() {
		var chunk indexer.Chunk
		if err := rows.Scan(&chunk.ID, &chunk.Path, &chunk.LineStart, &chunk.LineEnd, &chunk.Content, &chunk.Priority, &chunk.Hash, &chunk.SourceKind, &chunk.SourceName); err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	return chunks, rows.Err()
}

func (s *Store) ChunkCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM chunks`).Scan(&count)
	return count, err
}

func (s *Store) CountChunksBySource(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			CASE
				WHEN source_name != '' THEN source_name
				WHEN source_kind != '' THEN source_kind
				ELSE 'repo'
			END AS source_label,
			COUNT(1)
		FROM chunks
		GROUP BY source_label
		ORDER BY source_label
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			return nil, err
		}
		out[name] = count
	}
	return out, rows.Err()
}

func (s *Store) SetMeta(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO meta(key, value) VALUES(?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

func (s *Store) Meta(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (s *Store) SaveSession(ctx context.Context, session Session) error {
	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now
	allowed, err := json.Marshal(session.AllowedCommands)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO sessions(
			id, prompt, status, repo_root, worktree_path, branch, allowed_commands,
			iteration, max_iterations, dry_run, summary, commit_message, last_error, created_at, updated_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			prompt=excluded.prompt,
			status=excluded.status,
			repo_root=excluded.repo_root,
			worktree_path=excluded.worktree_path,
			branch=excluded.branch,
			allowed_commands=excluded.allowed_commands,
			iteration=excluded.iteration,
			max_iterations=excluded.max_iterations,
			dry_run=excluded.dry_run,
			summary=excluded.summary,
			commit_message=excluded.commit_message,
			last_error=excluded.last_error,
			updated_at=excluded.updated_at
	`, session.ID, session.Prompt, session.Status, session.RepoRoot, session.WorktreePath, session.Branch, string(allowed), session.Iteration, session.MaxIterations, boolToInt(session.DryRun), session.Summary, session.CommitMessage, session.LastError, session.CreatedAt.Format(time.RFC3339), session.UpdatedAt.Format(time.RFC3339))
	return err
}

func (s *Store) GetSession(ctx context.Context, id string) (Session, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, prompt, status, repo_root, worktree_path, branch, allowed_commands,
			iteration, max_iterations, dry_run, summary, commit_message, last_error, created_at, updated_at
		FROM sessions WHERE id = ?
	`, id)
	var session Session
	var allowed string
	var dryRun int
	var createdAt string
	var updatedAt string
	if err := row.Scan(&session.ID, &session.Prompt, &session.Status, &session.RepoRoot, &session.WorktreePath, &session.Branch, &allowed, &session.Iteration, &session.MaxIterations, &dryRun, &session.Summary, &session.CommitMessage, &session.LastError, &createdAt, &updatedAt); err != nil {
		return Session{}, err
	}
	_ = json.Unmarshal([]byte(allowed), &session.AllowedCommands)
	session.DryRun = dryRun == 1
	session.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	session.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return session, nil
}

func (s *Store) AppendEvent(ctx context.Context, sessionID, role, content string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO session_events(session_id, role, content, created_at) VALUES(?, ?, ?, ?)`, sessionID, role, content, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (s *Store) Events(ctx context.Context, sessionID string) ([]SessionEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT seq, session_id, role, content, created_at FROM session_events WHERE session_id = ? ORDER BY seq`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []SessionEvent
	for rows.Next() {
		var event SessionEvent
		var createdAt string
		if err := rows.Scan(&event.Seq, &event.SessionID, &event.Role, &event.Content, &createdAt); err != nil {
			return nil, err
		}
		event.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) ListSessions(ctx context.Context) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, prompt, status, repo_root, worktree_path, branch, allowed_commands,
			iteration, max_iterations, dry_run, summary, commit_message, last_error, created_at, updated_at
		FROM sessions ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		var session Session
		var allowed string
		var dryRun int
		var createdAt string
		var updatedAt string
		if err := rows.Scan(&session.ID, &session.Prompt, &session.Status, &session.RepoRoot, &session.WorktreePath, &session.Branch, &allowed, &session.Iteration, &session.MaxIterations, &dryRun, &session.Summary, &session.CommitMessage, &session.LastError, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(allowed), &session.AllowedCommands)
		session.DryRun = dryRun == 1
		session.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		session.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func NormalizeCommands(commands []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		if _, ok := seen[command]; ok {
			continue
		}
		seen[command] = struct{}{}
		out = append(out, command)
	}
	return out
}
