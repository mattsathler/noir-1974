package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var envCandidates = []string{".env.local", ".env", "../../.env.local", "../../.env"}

var gameTables = []string{
	"player_case_progress", "case_states", "case_activity_records", "case_mail_messages",
	"case_evidence", "case_relationships", "cases", "player_locations", "player_characters",
	"player_sessions", "player_accounts", "world_mutations", "company_locations",
	"personal_relationships", "employment_relationships", "world_entities", "world_regions",
	"worlds", "game_servers",
}

const truncateSQL = `TRUNCATE TABLE
	player_case_progress,
	case_states,
	case_activity_records,
	case_mail_messages,
	case_evidence,
	case_relationships,
	cases,
	player_locations,
	player_characters,
	player_sessions,
	player_accounts,
	world_mutations,
	company_locations,
	personal_relationships,
	employment_relationships,
	world_entities,
	world_regions,
	worlds,
	game_servers
	RESTART IDENTITY CASCADE`

func main() {
	confirm := flag.Bool("yes", false, "wipe the configured database and rotate the world seed")
	flag.Parse()

	_, seedOverride := os.LookupEnv("NOIR_WORLD_SEED")
	root, err := findProjectRoot()
	check(err)
	envPath, err := loadAPIEnvironment(root)
	check(err)

	if seedOverride {
		fatal("NOIR_WORLD_SEED is set in the process environment; remove or rotate that override before resetting")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fatal("DATABASE_URL is required")
	}
	target, err := url.Parse(databaseURL)
	check(err)
	if target.Hostname() == "" || target.Path == "" {
		fatal("DATABASE_URL does not identify a database endpoint")
	}

	fmt.Printf("Target: host=%s database=%s\n", target.Hostname(), strings.TrimPrefix(target.Path, "/"))
	fmt.Printf("Rows to clear: all Noir game tables (%d), including player accounts and characters\n", len(gameTables))
	fmt.Printf("Seed file: %s\n", envPath)
	if !*confirm {
		fmt.Println("Dry run only. Re-run with --yes after confirming this is the intended database and the API is stopped.")
		return
	}
	if err := ensureLocalAPIStopped(); err != nil {
		fatal(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, databaseURL)
	check(err)
	defer conn.Close(context.Background())

	tx, err := conn.Begin(ctx)
	check(err)
	if _, err := tx.Exec(ctx, truncateSQL); err != nil {
		_ = tx.Rollback(ctx)
		fatal(err.Error())
	}

	seedBytes := make([]byte, 20)
	if _, err := rand.Read(seedBytes); err != nil {
		_ = tx.Rollback(ctx)
		fatal(err.Error())
	}
	newSeed := "noir-1974-world-" + hex.EncodeToString(seedBytes)
	oldFile, err := os.ReadFile(envPath)
	if err != nil {
		_ = tx.Rollback(ctx)
		fatal(err.Error())
	}
	updatedFile := replaceSeed(string(oldFile), newSeed)
	if err := atomicWrite(envPath, []byte(updatedFile)); err != nil {
		_ = tx.Rollback(ctx)
		fatal(err.Error())
	}
	if err := tx.Commit(ctx); err != nil {
		_ = atomicWrite(envPath, oldFile)
		fatal(err.Error())
	}

	fmt.Printf("Reset complete. New NOIR_WORLD_SEED: %s\n", newSeed)
	fmt.Println("All game table rows were cleared. Starting the API will recreate the shared world and opening cases.")
}

func findProjectRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := current; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "apps", "api", "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not locate Noir project root (apps/api/go.mod)")
		}
	}
}

func loadAPIEnvironment(root string) (string, error) {
	apiDir := filepath.Join(root, "apps", "api")
	for _, candidate := range envCandidates {
		path := filepath.Join(apiDir, candidate)
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				return "", err
			}
			return path, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return "", errors.New("no API .env.local or .env file found")
}

func replaceSeed(contents, seed string) string {
	lines := strings.Split(contents, "\n")
	seedFound := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "NOIR_WORLD_SEED=") {
			lines[i] = "NOIR_WORLD_SEED=" + seed
			seedFound = true
		}
	}
	if !seedFound {
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = append(lines[:len(lines)-1], "NOIR_WORLD_SEED="+seed, "")
		} else {
			lines = append(lines, "NOIR_WORLD_SEED="+seed)
		}
	}
	return strings.Join(lines, "\n")
}

func atomicWrite(path string, contents []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".noir-env-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(contents); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func ensureLocalAPIStopped() error {
	connection, err := net.DialTimeout("tcp", "127.0.0.1:8080", 300*time.Millisecond)
	if err != nil {
		return nil
	}
	_ = connection.Close()
	return errors.New("the API appears to be running on localhost:8080; stop it before resetting the database")
}

func check(err error) {
	if err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
