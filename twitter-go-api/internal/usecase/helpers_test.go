package usecase

import (
	"reflect"
	"testing"
)

func TestExtractHashtags(t *testing.T) {
	t.Parallel()

	content := "hello #Go #go #backend and #Go_Lang"
	got := extractHashtags(content)
	want := []string{"go", "backend", "go_lang"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected hashtags: got %v want %v", got, want)
	}
}

func TestBuildTSQuery(t *testing.T) {
	t.Parallel()

	got := buildTSQuery("  hello, world!! go-lang ")
	want := "hello & world & go & lang"
	if got != want {
		t.Fatalf("unexpected tsquery: got %q want %q", got, want)
	}
}
