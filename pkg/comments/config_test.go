package comments

import "testing"

func TestConfig(t *testing.T) {
	if len(languages) == 0 {
		t.Fail()
	}
}

func TestLanguages(t *testing.T) {
	tests := []struct {
		lang      string
		extension string
	}{
		{lang: "Java", extension: ".java"},
		{lang: "Python", extension: ".py"},
		{lang: "JavaScript", extension: ".js"},
	}
	for _, test := range tests {
		t.Run(test.lang, func(t *testing.T) {
			for _, extension := range languages[test.lang].Extensions {
				if extension == test.extension {
					return
				}
			}
			t.Fail()
		})
	}
}

func TestCommentStyle(t *testing.T) {
	tests := []struct {
		filename       string
		commentStyleID string
	}{
		{filename: "Test.java", commentStyleID: "SlashAsterisk"},
		{filename: "Test.py", commentStyleID: "PythonStyle"},
	}
	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			if style := FileCommentStyle(test.filename); test.commentStyleID != style.ID {
				t.Logf("Extension = %v, Comment style = %v", test.filename, style.ID)
				t.Fail()
			}
		})
	}
}
