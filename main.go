package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/zono-dev/stplib"
)

// ImgTag is HTML tag values for view of ImgList.
type ImgTag struct {
	Href   string // Attribute of 'a' tag.
	ImgSrc string // Path to img link which is attribute of 'img' tag.
	Alt    string // Attribute of 'img' tag.
}

// PageParam is params for a stph template.
type PageParam struct {
	Items   []ImgTag // HTML tags attributes.
	FlexNum int      // Number of the images in per line in HTML. 'Flex' is came from css flex property.
}

// Conf has params which this application working needs.
// this comes from settings.yaml.
var Conf map[string]string

// CreateImgTag returns ImgTags are created from ImgInfo and bp.
func CreateImgTag(items []stplib.ImgInfo, bp string) []ImgTag {
	its := []ImgTag{}
	for _, v := range items {
		its = append(its, ImgTag{
			Href:   bp + v.OrgPath,
			ImgSrc: bp + v.ResizedFilePath,
			Alt:    v.FileName,
		})
	}
	return its
}

// DeleteImges deletes images in S3 and lines in DynamoDB table.
// If deletation failed, DeleteImages will return non-nil value.
func DeleteImages(bucket string, items []stplib.ImgInfo, files []string, table string) error {
	for _, v := range files {
		tgt := SearchItem(items, v)
		if tgt == nil {
			return nil
		} else {
			err := DelObjS3(bucket, []string{tgt.OrgPath, tgt.ResizedFilePath})
			if err != nil {
				fmt.Printf("Failed to delete on S3 bucket[%s], target file name=%#v\n", bucket, files)
				return err
			}
			err = DeleteItem(table, v)
			if err != nil {
				fmt.Printf("Failed to delete. target file name=%s\n", v)
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

func IsDiv(i int, base int, plus int) bool {
	return (i+plus)%base == 0
}

func CreatePage(tmpl string, pp PageParam) (string, error) {
	var out bytes.Buffer
	funcMap := template.FuncMap{
		"isDiv": IsDiv,
	}
	tpl := template.Must(template.New(filepath.Base(tmpl)).Funcs(funcMap).ParseFiles(tmpl))
	err := tpl.Execute(&out, pp)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return out.String(), nil
}

// InitPage get items from DynamoDB table and create page param.
func InitPage() ([]stplib.ImgInfo, PageParam, error) {
	items, err := GetItems(Conf["table_name"])

	if err != nil {
		fmt.Printf("[ERROR] Failed to get items with GetItems %#v.\n", err)
		return []stplib.ImgInfo{}, PageParam{}, err
	}

	// sort by time of file creation
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	pp := PageParam{}
	pp.Items = CreateImgTag(items, Conf["base_url"])
	pp.FlexNum = 4
	return items, pp, err
}

// IndexPage is a handle function for index page. It gets data from DynamoDB table and outputs index page.
func IndexPage(w http.ResponseWriter, r *http.Request) {

	_, pp, err := InitPage()
	if err != nil {
		fmt.Printf("[ERROR] Failed to InitPage. %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	out, err := CreatePage(Conf["tmpl_path"], pp)
	if err != nil {
		fmt.Printf("[ERROR] Failed to CreatePage. %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)
	return
}

// DeletePage deletes the image files and data in DynamoDB. After that, it redirects to index page.
func DeletePage(w http.ResponseWriter, r *http.Request) {

	// Redirect all methods except 'post'.
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", 301)
		return
	}

	items, _, err := InitPage()
	if err != nil {
		fmt.Printf("[ERROR] Failed to InitPage. %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get value from html form which has 'name="del"' attribute.
	rv := r.FormValue("del")
	fmt.Println(rv)
	err = DeleteImages(Conf["bucket_name"], items, []string{rv}, Conf["table_name"])
	if err != nil {
		fmt.Printf("Failed to Delete. :(. %#v", err)
	}
	// Anyway, redirect to index page...
	http.Redirect(w, r, "/", 301)
	return
}

// RegistHandle sets handle functions
func RegistHandle() {
	http.HandleFunc("/", IndexPage)
	http.HandleFunc("/delete", DeletePage)
}

// main function. Let's get started. :)
func main() {
	Conf = ReadConfig(filepath.Join("configs", "settings.yaml"))
	s3svc = NewS3Sess(Conf["region"])
	fmt.Println(Conf)

	RegistHandle()
	log.Fatal(http.ListenAndServe((":" + Conf["port"]), nil))
}
