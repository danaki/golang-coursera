package main

import (
	"bufio"
	json "encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	// "log"
)

// {
// 	"browsers": [
// 		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36",
// 		"LG-LX550 AU-MIC-LX550/2.0 MMP/2.0 Profile/MIDP-2.0 Configuration/CLDC-1.1",
// 		"Mozilla/5.0 (Android; Linux armv7l; rv:10.0.1) Gecko/20100101 Firefox/10.0.1 Fennec/10.0.1",
// 		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; MATBJS; rv:11.0) like Gecko"
// 		],
// 	"company":"Flashpoint",
// 	"country":"Dominican Republic",
// 	"email":"JonathanMorris@Muxo.edu",
// 	"job":"Programmer Analyst #{N}",
// 	"name":"Sharon Crawford",
// 	"phone":"176-88-49"
// }

type User struct {
	Browsers []string
	Company  string
	Country  string
	Email    string
	Job      string
	Name     string
	Phone    string
}

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func (v User) String() string {
	return v.Email + " " + v.Name
}

func easyjson9e1087fdDecodeGithubComDanakiHw3BenchUser(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					out.Browsers = append(out.Browsers, in.UnsafeString())
					in.WantComma()
				}
				in.Delim(']')
			}
		case "company":
			out.Company = in.UnsafeString()
		case "country":
			out.Country = in.UnsafeString()
		case "email":
			out.Email = in.UnsafeString()
		case "job":
			out.Job = in.UnsafeString()
		case "name":
			out.Name = in.UnsafeString()
		case "phone":
			out.Phone = in.UnsafeString()
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeGithubComDanakiHw3BenchUser(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeGithubComDanakiHw3BenchUser(l, v)
}

// SlowSearch ...
func FastSearch(out io.Writer) {
	seenBrowsers := make(map[string]bool)
	uniqueBrowsers := 0

	var user User
	var isAndroid, isMSIE bool
	var line, browser string
	var il, ib int

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fmt.Fprint(out, "found users:")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		il++
		line = scanner.Text()
		err := user.UnmarshalJSON([]byte(line))

		if err != nil {
			panic(err)
		}

		isAndroid = false
		isMSIE = false

		for ib = range user.Browsers {
			browser = user.Browsers[ib]

			if strings.Contains(browser, "Android") {
				isAndroid = true
			} else if strings.Contains(browser, "MSIE") {
				isMSIE = true
			} else {
				continue
			}

			if _, ok := seenBrowsers[browser]; !ok {
				seenBrowsers[browser] = true
				uniqueBrowsers++
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		fmt.Fprint(out, fmt.Sprintf("\n[%d] %s <%s>", il-1, user.Name, strings.Replace(user.Email, "@", " [at] ", -1)))
	}

	fmt.Fprintln(out, "\n\nTotal unique browsers", len(seenBrowsers))
}
