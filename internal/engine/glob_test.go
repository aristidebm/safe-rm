package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesAnySimpleStar(t *testing.T) {
	matched, err := MatchesAny("/home/user/docs/report.txt", []string{"/home/user/docs/*.txt"})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected /home/user/docs/*.txt to match")
	}
}

func TestMatchesAnyQuestionMark(t *testing.T) {
	matched, err := MatchesAny("/home/user/docs/report.txt", []string{"/home/user/docs/report.???"})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected report.??? to match report.txt")
	}
}

func TestMatchesAnyNoMatch(t *testing.T) {
	matched, err := MatchesAny("/home/user/docs/report.txt", []string{"*.pdf", "*.doc"})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if matched {
		t.Fatal("expected no match for .pdf/.doc patterns against .txt file")
	}
}

func TestMatchesAnyWithGlobstar(t *testing.T) {
	matched, err := MatchesAny("/home/user/projects/go/src/main.go", []string{"/home/user/**/main.go"})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected **/main.go to match deep path")
	}
}

func TestMatchesAnyGlobstarAnyDir(t *testing.T) {
	matched, err := MatchesAny("/home/user/any/deep/path/file.txt", []string{"/home/user/**/file.txt"})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected ** to match any directory depth")
	}
}

func TestMatchesAnyTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	matched, err := MatchesAny(filepath.Join(home, "docs", "notes.txt"), []string{"~/docs/*"})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected ~/docs/* to match $HOME/docs/notes.txt")
	}
}

func TestMatchesAnyRelativePath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	matched, err := MatchesAny("relative/path/file.txt", []string{filepath.Join(wd, "relative/path/*")})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected relative path to be resolved and match")
	}
}

func TestMatchesAnyEmptyPatterns(t *testing.T) {
	matched, err := MatchesAny("/some/file.txt", []string{})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if matched {
		t.Fatal("expected no match for empty patterns")
	}
}

func TestMatchGlobstarPrefixOnly(t *testing.T) {
	matched, err := MatchesAny("/home/user/foo/bar/baz.txt", []string{"**/baz.txt"})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected **/baz.txt to match")
	}
}

func TestMatchGlobstarExact(t *testing.T) {
	matched, err := MatchesAny("/home/user/file.txt", []string{"**"})
	if err != nil {
		t.Fatalf("MatchesAny() returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected ** to match anything")
	}
}
