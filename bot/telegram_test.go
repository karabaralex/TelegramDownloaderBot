package bot

import (
	"fmt"
	"os"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	// clear pairs
	fmt.Println("setup", len(pair))
	for m := range pair {
		delete(pair, m)
	}
}

// test command matcher match works
func TestCommandMatcher(t *testing.T) {
	matcher := NewCommandMatcher("/test")
	update := &tgbotapi.Update{Message: &tgbotapi.Message{Text: "/test", Entities: []tgbotapi.MessageEntity{{Offset: 0, Length: 4, Type: "bot_command"}}}}
	if !matcher.match(update) {
		t.Error("command matcher match failed")
	}
}

func TestCommandMatcherShouldNotMatchIfNoEntity(t *testing.T) {
	matcher := NewCommandMatcher("/test")
	update := &tgbotapi.Update{Message: &tgbotapi.Message{Text: "/test"}}
	if matcher.match(update) {
		t.Error("command matcher match failed")
	}
}

func TestCommandMatcherRegex(t *testing.T) {
	// regex matches any number
	matcher := NewCommandMatcher("/(\\d+)")
	if !matcher.match(&tgbotapi.Update{Message: &tgbotapi.Message{Text: "/123", Entities: []tgbotapi.MessageEntity{{Offset: 0, Length: 4, Type: "bot_command"}}}}) {
		t.Error("command matcher match failed")
	}
}

func TestCommandMatcherNotMatch(t *testing.T) {
	// regex matches any number
	matcher := NewCommandMatcher("/(\\d+)")
	if matcher.match(&tgbotapi.Update{Message: &tgbotapi.Message{Text: "/abc"}}) {
		t.Error("command matcher match failed")
	}
}

func TestTextMatcherRegex(t *testing.T) {
	// regex matches any number
	matcher := NewTextMatcher("(\\d+)")
	if !matcher.match(&tgbotapi.Update{Message: &tgbotapi.Message{Text: "123"}}) {
		t.Error("command matcher match failed")
	}
}

// test find hanlder for update
func TestFindHandlerForUpdate(t *testing.T) {
	// add handler
	AddHandler(NewCommandMatcher("/test"), func(message *Info) {
		message.Text = "test"
	})
	update := &tgbotapi.Update{Message: &tgbotapi.Message{Text: "/test", Entities: []tgbotapi.MessageEntity{{Offset: 0, Length: 4, Type: "bot_command"}}}}
	if _, ok := findHandlerForUpdate(update); !ok {
		t.Error("find handler for update failed")
	}
}

func TestFindHandlerForUpdateInMultiple(t *testing.T) {
	AddHandler(NewCommandMatcher("/test1"), func(message *Info) {
		message.Text = "test1"
	})
	AddHandler(NewCommandMatcher("/test2"), func(message *Info) {
		message.Text = "test2"
	})
	update := &tgbotapi.Update{Message: &tgbotapi.Message{Text: "/test66", Entities: []tgbotapi.MessageEntity{{Offset: 0, Length: 4, Type: "bot_command"}}}}
	if _, ok := findHandlerForUpdate(update); ok {
		t.Error("found handler but should not")
	}
	update = &tgbotapi.Update{Message: &tgbotapi.Message{Text: "/test2", Entities: []tgbotapi.MessageEntity{{Offset: 0, Length: 4, Type: "bot_command"}}}}
	if _, ok := findHandlerForUpdate(update); !ok {
		t.Error("find handler for update failed")
	}
}

func TestFileMatcher(t *testing.T) {
	matcher := NewFileNameMatcher()
	update := &tgbotapi.Update{Message: &tgbotapi.Message{Document: &tgbotapi.Document{FileName: "test"}}}
	if !matcher.match(update) {
		t.Error("file matcher match failed")
	}
}

func TestFinalMatches(t *testing.T) {
	postIdMatcher := NewCommandMatcher("/[0-9]+")
	watchMatcher := NewCommandMatcher("/watch")
	anyTextMatcher := NewTextMatcher(".*")
	fileMatcher := NewFileNameMatcher()
	update := &tgbotapi.Update{Message: &tgbotapi.Message{Text: "/test", Entities: []tgbotapi.MessageEntity{{Offset: 0, Length: 4, Type: "bot_command"}}}}
	matchers := []Matcher{postIdMatcher, watchMatcher, anyTextMatcher, fileMatcher}
	for _, matcher := range matchers {
		if matcher.match(update) {
			t.Error("nothing should match, but it did", matcher)
		}
	}

	update = &tgbotapi.Update{Message: &tgbotapi.Message{Text: "some text"}}
	for _, matcher := range matchers {
		if matcher.match(update) && matcher != anyTextMatcher {
			t.Error("nothing should match, but it did", matcher)
		}
	}

	update = &tgbotapi.Update{Message: &tgbotapi.Message{Text: "/1234", Entities: []tgbotapi.MessageEntity{{Offset: 0, Length: 4, Type: "bot_command"}}}}
	for _, matcher := range matchers {
		if matcher.match(update) && matcher != postIdMatcher {
			t.Error("nothing should match, but it did", matcher)
		}
	}

	update = &tgbotapi.Update{Message: &tgbotapi.Message{Document: &tgbotapi.Document{FileName: "test"}}}
	for _, matcher := range matchers {
		if matcher.match(update) && matcher != fileMatcher {
			t.Error("nothing should match, but it did", matcher)
		}
	}
}
