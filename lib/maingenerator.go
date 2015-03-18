package lib

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"io"
	"os"

	"golang.org/x/tools/imports"
)

type mainGenerator struct {
	rootDbPackageName       string
	rootHandlersPackageName string
}

func (g *mainGenerator) Generate(tables map[string]tableInfo) error {
	os.Mkdir("bin", 0755)
	importstmpl := template.Must(template.New("mainImports").Parse(`
package main
import({{$rootHandlersPackageName := .rootHandlersPackageName}}{{$rootdbpackagename := .rootdbpackagename}}
{{range .Tables}}"{{$rootHandlersPackageName}}/{{.TableName}}"
{{.TableName}}db "{{$rootdbpackagename}}/{{.TableName}}"
{{end}}
"net/http"
	"github.com/gorilla/mux"
"os"
"database/sql"
	_ "github.com/ziutek/mymysql/godrv"
)
`))
	importstmpl = importstmpl
	routestmpl := template.Must(template.New("mainRoutes").Parse(`
func main(){
	dbUrl := os.Args[1]
    	dbconn, err := sql.Open("mymysql", dbUrl)
	if err != nil {
		panic(err)
	}
    db.MustValidateChecksum(dbconn, os.Args[2])
    {{range .Tables}}
    {{.TableName}}db.DB = dbconn
    {{end}}
    r := mux.NewRouter()
    g := r.Methods("GET").Subrouter()
    po := r.Methods("POST").Subrouter()
    pu := r.Methods("PUT").Subrouter()
    d := r.Methods("DELETE").Subrouter()
{{range .Tables}}
g.HandleFunc("/{{.TableName}}/", {{.TableName}}.List)
g.HandleFunc("/{{.TableName}}/{id}/", {{.TableName}}.Get)
po.HandleFunc("/{{.TableName}}/", {{.TableName}}.Post)
pu.HandleFunc("/{{.TableName}}/{id}/", {{.TableName}}.Put)
d.HandleFunc("/{{.TableName}}/{id}/", {{.TableName}}.Delete)
{{end}}

http.ListenAndServe(":8080",r)
}
`))
	routestmpl = routestmpl
	var b bytes.Buffer
	path, err := GetRootPath()
	if err != nil {
		return err
	}
	err = importstmpl.Execute(&b, map[string]interface{}{"rootHandlersPackageName": path + "/http", "Tables": tables, "rootdbpackagename": path + "/db"})
	if err != nil {
		fmt.Println(err)
	}
	var final bytes.Buffer
	io.Copy(&final, &b)
	b = bytes.Buffer{}
	routestmpl.Execute(&b, map[string]interface{}{"Verbs": []string{"List", "Get", "Post", "Put", "Delete"}, "Tables": tables})
	io.Copy(&final, &b)
	f, err := os.Create("bin/main.go")
	if err != nil {
		return err
	}

	formatted, err := format.Source(final.Bytes())
	if err != nil {
		return err
	}
	formatted, err = imports.Process(f.Name(), formatted, nil)
	if err != nil {
		return err
	}
	io.Copy(f, bytes.NewBuffer(formatted))
	return nil
}