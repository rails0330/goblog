package main

import (
    "github.com/hoisie/mustache"
    "github.com/hoisie/web"
    "github.com/kless/goconfig/config"
    "database/sql"
    "strings"
    "strconv"
     _ "github.com/jbarham/gopgsqldriver"
)

// Connect to the database
var db, err = sql.Open("postgres", "user=Steve dbname=goblog host=localhost port=5432")

type Entry struct {
    Id int
    Title, Content string
}

/*
* Handles rendering templates in a normalized context
*/
func RenderTemplate(template string, context map[string]interface{})string {
    c, _ := config.ReadDefault("goblog.conf")

    title, _ := c.String("general", "title")
    motto, _ := c.String("general", "motto")

    var send = map[string]interface{} {
        "title": title,
        "motto": motto,
    }
    // Append all values of context to the global context
    for key, val := range context {
        send[key] = val
    }

    return mustache.RenderFileInLayout("templates/" + template, "templates/base.mustache", send)
}

/*
* Main page
*/
func Index() string {
    rows, _ := db.Query("SELECT id, title, content FROM entries ORDER BY id DESC")

    entries := []*Entry {}

    // Get the entries
    for i := 0; rows.Next(); i++ {
        var entry = new(Entry)

        rows.Scan(&entry.Id, &entry.Title, &entry.Content)

        // Parse newlines
        entry.Content = strings.Replace(entry.Content, "\r\n\r\n", "</p><p>", -1);

        entries = append(entries, entry)
    }

    send := make(map[string]interface{})

    if len(entries) == 0 {
        send["entries"] = false
    } else {
        send["entries"] = entries
    }

    return RenderTemplate("index.mustache", send)
}

func Manage() string {
    return RenderTemplate("manage.mustache", nil)
}

func Existing() string {
    rows, _ := db.Query("SELECT id, title, content FROM entries ORDER BY id DESC")

    entries := []*Entry {}
    for i := 0; rows.Next(); i++ {
        var entry = new(Entry)
        rows.Scan(&entry.Id, &entry.Title, nil)

        entries = append(entries, entry)
    }

    send := make(map[string]interface{})
    if len(entries) == 0 {
        send["entries"] = false
    } else {
        send["entries"] = entries
    }

    return RenderTemplate("existing.mustache", send)
}

func ExistingEdit(val string) string {
    id, err := strconv.Atoi(val)
    if err != nil {
        return "Invalid or malformed id"
    }

    row := db.QueryRow("SELECT id, title, content FROM entries WHERE id=$1 LIMIT 1", id)
    entry := new(Entry)
    err = row.Scan(&entry.Id, &entry.Title, &entry.Content)
    if err != nil {
        return err.Error()
    }

    send := map[string]interface{} {
        "Id": entry.Id,
        "Title": entry.Title,
        "Content": entry.Content,
    }

    return RenderTemplate("existing_edit.mustache", send);
}

func Create(ctx *web.Context) string {
    // Check to see if we're actually publishing
    title, exists_title := ctx.Params["title"]
    content, exists_content := ctx.Params["content"]

    var send = map[string]interface{} {
        "show_success": false,
    }

    // We are! So let's save it
    if exists_title && exists_content {
        // Insert the row
        _, err := db.Exec("INSERT INTO entries (title, content) VALUES($1, $2)", title, content)

        if err != nil {
            return RenderTemplate("error.mustache", nil)
        }

        send["show_success"] = true
    }

    return RenderTemplate("create.mustache", send)
}

func main() {
    web.Get("/", Index)
    web.Get("/manage", Manage)
    web.Get("/manage/create", Create)
    web.Post("/manage/create", Create)
    web.Get("/manage/existing", Existing)
    web.Get("/manage/existing/(.*)", ExistingEdit)
    web.Run("0.0.0.0:9999")
}
