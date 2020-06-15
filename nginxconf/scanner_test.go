package nginxconf_test

import (
	"testing"

	"github.com/bingoohuang/gonginx/nginxconf"
)

func TestScanner(t *testing.T) {
	content := []byte(`
#COMMENT
# DOUBLE #COMMENT
WORD1 WORD2;
WORD3 {
    WORD4 'SQ\t\r\n\'\"\\1' "DQ\t\r\n\'\"\\1";
}`)
	scanner := nginxconf.NewScanner(content)
	expectedTokens := []nginxconf.Token{
		{Typ: nginxconf.Comment, Lit: "COMMENT"},
		{Typ: nginxconf.Comment, Lit: " DOUBLE #COMMENT"},
		{Typ: nginxconf.Word, Lit: "WORD1"},
		{Typ: nginxconf.Word, Lit: "WORD2"},
		nginxconf.SemicolonToken,
		{Typ: nginxconf.Word, Lit: "WORD3"},
		nginxconf.BraceOpenToken,
		{Typ: nginxconf.Word, Lit: "WORD4"},
		{Typ: nginxconf.Word, Lit: "SQ\t\r\n'\"\\1"},
		{Typ: nginxconf.Word, Lit: "DQ\t\r\n'\"\\1"},
		nginxconf.SemicolonToken,
		nginxconf.BraceCloseToken,
	}

	for i, expectedToken := range expectedTokens {
		token := scanner.Scan()
		if token != expectedToken {
			t.Errorf("unexpected nginxconf.Token: i=%d, expected=%s, actual=%q\n", i, expectedToken, token)
			t.FailNow()
		}
	}

	if token := scanner.Scan(); token.Typ != nginxconf.EOF {
		t.Errorf("unexpected nginxconf.Token: expected=%s, actual=%q\n", nginxconf.EOFToken, token)
	}
}

func TestScanUnterminatedSingleQuotedString1(t *testing.T) {
	content := []byte(`'WORD2`)
	scanner := nginxconf.NewScanner(content)

	var tok nginxconf.Token

	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()

	tok = scanner.Scan()
}

func TestScanUnterminatedSingleQuotedString2(t *testing.T) {
	content := []byte("'WORD2\n")
	scanner := nginxconf.NewScanner(content)

	var tok nginxconf.Token

	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()

	tok = scanner.Scan()
}

func TestScanInvalidQuotedCharInSingleQuotedString(t *testing.T) {
	content := []byte(`'WORD2\/'`)
	scanner := nginxconf.NewScanner(content)

	var tok nginxconf.Token

	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()

	tok = scanner.Scan()
}

func TestScanUnterminatedDoubleQuotedString1(t *testing.T) {
	content := []byte(`"WORD2`)
	scanner := nginxconf.NewScanner(content)

	var tok nginxconf.Token

	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()

	tok = scanner.Scan()
}

func TestScanUnterminatedDoubleQuotedString2(t *testing.T) {
	content := []byte("\"WORD2\n")
	scanner := nginxconf.NewScanner(content)

	var tok nginxconf.Token

	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()

	tok = scanner.Scan()
}

func TestScanInvalidQuotedCharInDoubleQuotedString(t *testing.T) {
	content := []byte(`"WORD2\/"`)
	scanner := nginxconf.NewScanner(content)

	var tok nginxconf.Token

	defer func() {
		if err := recover(); err == nil {
			t.Error("unexpected scan result:", tok)
		}
	}()

	tok = scanner.Scan()
}

func TestScanLastWord(t *testing.T) {
	content := []byte("WORD")
	scanner := nginxconf.NewScanner(content)

	defer func() {
		if err := recover(); err != nil {
			t.Error("scan fail:", err.(error).Error())
		}
	}()

	tok := scanner.Scan()

	if tok.Lit != "WORD" {
		t.Error("unexpected result:", tok)
	}
}
