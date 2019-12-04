package feed

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "delete_source.html",
		FileModTime: time.Unix(1574533501, 0),

		Content: string("<!DOCTYPE html>\n<html>\n<body>\n    <h1>feed source</h1>\n    <form method=\"post\" action=\"/feeds/{{ .source.ID }}/delete\">\n        <div>ต้องการจะลบฟีดนี้? </div>\n\n        <div>\n            <p>ชื่อ: {{ .source.Name }}</p>\n            <p>url: {{ .source.URL }}</p>\n        </div>\n        <button>ยืนยัน</button>\n        <a href=\"/feeds\">ยกเลิก</a>\n\n        <div>* เนื้อหาของฟีดทั้งหมดจะถูกลบตามไปด้วย</div>\n    </form>\n</body>\n</html>"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "edit_source.html",
		FileModTime: time.Unix(1574534010, 0),

		Content: string("<!DOCTYPE html>\n<html>\n<body>\n    <h1>edit feed source (id={{.source.ID}})</h1>\n    <form method=\"post\" action=\"/feeds/{{.source.ID}}/edit\">\n        <div>\n            <label>URL</label>\n            <input name=\"url\" value=\"{{ .source.URL }}\">\n        </div>   \n        <div>\n            <label>Slug</label>\n            <input name=\"slug\" value=\"{{ .source.Slug }}\" >\n        </div>   \n        <div>\n            <label>Name</label>\n            <input name=\"name\" value=\"{{ .source.Name }}\">\n        </div>   \n\n        <input type=\"hidden\" value=\"{{.ID}}\" name=\"ID\" />\n        <button>Submit</button>\n    </form>\n</body>\n</html>"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "form.html",
		FileModTime: time.Unix(1574380935, 0),

		Content: string(""),
	}
	file5 := &embedded.EmbeddedFile{
		Filename:    "list.html",
		FileModTime: time.Unix(1574533963, 0),

		Content: string("<!DOCTYPE html>\n<html>\n\n<body>\n    <h1>feed source</h1>\n    <form method=\"post\" action=\"/feeds\">\n        <div>\n            <label>URL</label>\n            <input name=\"url\">\n        </div>\n        <div>\n            <label>Slug</label>\n            <input name=\"slug\">\n        </div>\n        <div>\n            <label>Name</label>\n            <input name=\"name\">\n        </div>\n\n        <button>Submit</button>\n    </form>\n    {{ if .message }}<div>{{.message}}</div>{{ end }}\n    <ol>\n        {{range .sources}}\n        <li>{{.Name}} - {{.URL}} [ <a href=\"/feeds/{{.ID}}/items\">items</a> | <a href=\"/feeds/{{.ID}}/edit\">edit</a> |\n            <a href=\"/feeds/{{.ID}}/delete\">delete</a>]</li>\n        {{end}}\n    </ol>\n</body>\n\n</html>"),
	}
	file6 := &embedded.EmbeddedFile{
		Filename:    "rss_raw.xml",
		FileModTime: time.Unix(1575476882, 0),

		Content: string("<?xml version=\"1.0\" encoding=\"utf-8\" standalone=\"yes\" ?>\n<rss version=\"2.0\"  xmlns:content=\"http://purl.org/rss/1.0/modules/content/\" xmlns:wfw=\"http://wellformedweb.org/CommentAPI/\" xmlns:dc=\"http://purl.org/dc/elements/1.1/\" xmlns:atom=\"http://www.w3.org/2005/Atom\" xmlns:sy=\"http://purl.org/rss/1.0/modules/syndication/\" xmlns:slash=\"http://purl.org/rss/1.0/modules/slash/\" xmlns:itunes=\"http://www.itunes.com/dtds/podcast-1.0.dtd\" xmlns:googleplay=\"http://www.google.com/schemas/play-podcasts/1.0\" xmlns:spotify=\"http://www.spotify.com/ns/rss\">\n  <channel>\n    <title>{{ .Config.Title | html }}</title>\n    <link>{{ .Config.PermaLink | html }}</link>\n    <description>{{ .Config.Description | html}}</description>\n    <image>\n        <url>{{ .Config.CoverURL | html  }}</url>\n        <title>{{ .Config.Title | html }}</title>\n        <link>{{ .Config.PermaLink | html }}</link>\n    </image>\n    <generator>rjio</generator>\n    <author>{{ .Config.Author }}</author>\n    <itunes:author>{{ .Config.Author }}</itunes:author>\n    <itunes:subtitle>{{ .Config.Description | html }}</itunes:subtitle>\n    <itunes:summary>{{ .Config.Description | html }}</itunes:summary>  \n    <itunes:owner>\n      <itunes:name>{{ .Config.Author }}</itunes:name>\n      <itunes:email>{{ .Config.Email }}</itunes:email>\n    </itunes:owner>\n    <itunes:image href=\"{{ .Config.CoverURL | html  }}\" />\n    <itunes:category text=\"{{ .Config.Category }}\" />\n    {{ with .Config.Language }}<language>{{.}}</language>{{end}}\n    <atom:link href=\"{{ .Config.PermaLink | html }}\" rel=\"self\" type=\"application/rss+xml\" />\n    {{ range .Entries }}{{ .Entry }}\n    {{ end }}\n  </channel>\n</rss>"),
	}
	file7 := &embedded.EmbeddedFile{
		Filename:    "view_feed_items.html",
		FileModTime: time.Unix(1575129266, 0),

		Content: string("<!DOCTYPE html>\n<html>\n<body>\n    <h1><a href=\"/feeds\">feeds</a> > items > {{ .source.Name}}</h1>\n\n    {{ if .message }}<div>{{.message}}</div>{{ end }}\n\n    <form method=\"POST\"  action=\"/feeds/{{ .source.ID}}/refresh\">\n        <button>Refresh</button>\n    </form>\n    <table>\n        <thead>\n            <tr>\n                <th>id</th>\n                <th>feed_id</th>\n                <th>guid</th>\n                <th>title</th>\n            </tr>\n        </thead>\n        <tbody>\n            {{ range .items }}\n            <tr>\n                <td>{{ .ID }}</td>\n                <td>{{ .FeedID }}</td>\n                <td>{{ .GUID }}</td> \n                <td>{{ .Title }}</td> \n                <td>{{ .PubDate }}</td> \n            </tr>\n            {{ end }}\n        </tbody>\n    </table>\n</body>\n</html>\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1574605252, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "delete_source.html"
			file3, // "edit_source.html"
			file4, // "form.html"
			file5, // "list.html"
			file6, // "rss_raw.xml"
			file7, // "view_feed_items.html"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`../templates`, &embedded.EmbeddedBox{
		Name: `../templates`,
		Time: time.Unix(1574605252, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"delete_source.html":   file2,
			"edit_source.html":     file3,
			"form.html":            file4,
			"list.html":            file5,
			"rss_raw.xml":          file6,
			"view_feed_items.html": file7,
		},
	})
}
