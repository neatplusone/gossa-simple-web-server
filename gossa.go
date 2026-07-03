package main

import (
	"archive/zip"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash"
	"html"
	"html/template"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	// Register WASM MIME type - required for WebAssembly to work correctly
	// Without this, browsers reject .wasm files with "expected magic word" errors
	mime.AddExtensionType(".wasm", "application/wasm")
}

type rowTemplate struct {
	Name    string
	Href    template.URL
	Size    string
	Ext     string
	ModTime string // Added for modification time display and sorting
}

type pageTemplate struct {
	Title       template.HTML
	ExtraPath   template.HTML
	Ro          bool
	RowsFiles   []rowTemplate
	RowsFolders []rowTemplate
}

var host = flag.String("h", "0.0.0.0", "host to listen to")
var port = flag.String("p", "8000", "port to listen to")
var extraPath = flag.String("prefix", "/", "url prefix at which gossa can be reached, e.g. /gossa/ (slashes of importance)")
var symlinks = flag.Bool("symlinks", false, "follow symlinks \033[4mWARNING\033[0m: symlinks will by nature allow to escape the defined path (default: false)")
var verb = flag.Bool("verb", false, "verbosity")
var skipHidden = flag.Bool("k", true, "\nskip hidden files")
var ro = flag.Bool("ro", false, "read only mode (no upload, rename, move, etc...)")
var calcFolderSize = flag.Bool("calcfoldersize", false, "calculate and display folder sizes (may slow down browsing in large directories)")
var maxUpload = flag.Int("maxupload", 0, "max upload size in megabytes per file (0 = unlimited)")
var tlsCert = flag.String("tls-cert", "", "path to TLS certificate file to serve HTTPS (requires -tls-key)")
var tlsKey = flag.String("tls-key", "", "path to TLS key file to serve HTTPS (requires -tls-cert)")
var auth = flag.String("auth", "", "enable HTTP basic auth, format user:pass (empty = disabled)")

type rpcCall struct {
	Call string   `json:"call"`
	Args []string `json:"args"`
}

var rootPath = ""
var handler http.Handler

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// needArgs panics (recovered by exitPath) when an rpc call is missing arguments.
func needArgs(rpc rpcCall, n int) {
	if len(rpc.Args) < n {
		panic(errors.New("missing rpc arguments"))
	}
}

// within reports whether p is rootPath itself or strictly inside it, guarding
// against sibling-directory prefix matches (e.g. /srv/data vs /srv/data-secret).
func within(p string) bool {
	return p == rootPath || strings.HasPrefix(p, rootPath+string(os.PathSeparator))
}

func exitPath(w http.ResponseWriter, s ...interface{}) {
	if r := recover(); r != nil {
		log.Println("error", s, r)
		w.WriteHeader(500)
		w.Write([]byte("error"))
	} else if *verb {
		log.Println(s...)
	}
}

func humanize(bytes int64) string {
	b := float64(bytes)
	u := 0
	for {
		if b < 1024 {
			return strconv.FormatFloat(b, 'f', 1, 64) + [9]string{"B", "k", "M", "G", "T", "P", "E", "Z", "Y"}[u]
		}
		b = b / 1024
		u++
	}
}

// Format time with full date and time
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// Calculate the total size of a directory recursively
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// Modify the replyList function to populate the ModTime field
func replyList(w http.ResponseWriter, r *http.Request, fullPath string, path string) {
	files, err := os.ReadDir(fullPath)
	check(err)
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name()) })

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	title := "/" + strings.TrimPrefix(path, *extraPath)
	p := pageTemplate{}
	p.ExtraPath = template.HTML(html.EscapeString(*extraPath))
	p.Ro = *ro
	p.Title = template.HTML(html.EscapeString(title))

	// Initialize RowsFolders and RowsFiles as empty slices
	p.RowsFolders = []rowTemplate{}
	p.RowsFiles = []rowTemplate{}

	// Add the parent directory entry first if not at root
	if path != *extraPath {
		// Special parent directory entry - use "up" indicator and add it first
		p.RowsFolders = append(p.RowsFolders, rowTemplate{"[Up]", "../", "", "up", ""})
	}

	for _, el := range files {
		info, errInfo := el.Info()
		el, err := os.Stat(fullPath + "/" + el.Name())
		if err != nil || errInfo != nil {
			log.Println("error - cant stat a file", err)
			continue
		}

		if *skipHidden && strings.HasPrefix(el.Name(), ".") {
			continue // dont print hidden files if we're not allowed
		}
		if !*symlinks && info.Mode()&os.ModeSymlink != 0 {
			continue // dont follow symlinks if we're not allowed
		}

		href := url.PathEscape(el.Name())
		name := el.Name()

		if el.IsDir() && strings.HasPrefix(href, "/") {
			href = strings.Replace(href, "/", "", 1)
		}

		// Format the modification time
		modTime := formatTime(el.ModTime())

		if el.IsDir() {
			size := ""
			if *calcFolderSize {
				if dirSize, err := getDirSize(filepath.Join(fullPath, el.Name())); err == nil {
					size = humanize(dirSize)
				}
			}
			row := rowTemplate{name + "/", template.URL(href), size, "folder", modTime}
			p.RowsFolders = append(p.RowsFolders, row)
		} else {
			sl := strings.Split(name, ".")
			ext := ""
			if len(sl) > 1 {
				ext = strings.ToLower(sl[len(sl)-1])
			}
			row := rowTemplate{name, template.URL(href), humanize(el.Size()), ext, modTime}
			p.RowsFiles = append(p.RowsFiles, row)
		}
	}

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Add("Content-Encoding", "gzip")
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		check(err)
		defer gz.Close()
		tmpl.Execute(gz, p)
	} else {
		tmpl.Execute(w, p)
	}
}

func doContent(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, *extraPath) { // redir when were not hitting the supplementary path if one is set
		http.Redirect(w, r, *extraPath, http.StatusFound)
		return
	}

	path := html.UnescapeString(r.URL.Path)
	defer exitPath(w, "get content", path)
	fullPath := enforcePath(path)
	stat, errStat := os.Stat(fullPath)
	check(errStat)

	// Serve .html files directly when explicitly requested (bypass FileServer's index.html redirect)
	if !stat.IsDir() && strings.HasSuffix(strings.ToLower(path), ".html") {
		http.ServeFile(w, r, fullPath)
		return
	}

	if stat.IsDir() {
		replyList(w, r, fullPath, path)
	} else {
		handler.ServeHTTP(w, r)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	path := r.Header.Get("gossa-path")
	defer exitPath(w, "upload", path)

	if *maxUpload > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, int64(*maxUpload)*1024*1024)
	}

	path, err := url.PathUnescape(path)
	check(err)
	reader, err := r.MultipartReader()
	check(err)
	part, err := reader.NextPart()
	if err != nil && err != io.EOF { // errs EOF when no more parts to process
		check(err)
	}
	dst, err := os.Create(enforcePath(path))
	check(err)
	defer dst.Close()
	_, err = io.Copy(dst, part)
	check(err)
	w.Write([]byte("ok"))
}

func zipRPC(w http.ResponseWriter, r *http.Request) {
	zipPath := r.URL.Query().Get("zipPath")
	zipName := r.URL.Query().Get("zipName")
	defer exitPath(w, "zip", zipPath)
	zipFullPath := enforcePath(zipPath)
	_, err := os.Lstat(zipFullPath)
	check(err)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", zipName+".zip"))
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	err = filepath.Walk(zipFullPath, func(path string, f fs.FileInfo, err error) error {
		check(err)
		if f.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(zipFullPath, path)
		check(err)
		if *skipHidden && (strings.HasPrefix(rel, ".") || strings.HasPrefix(f.Name(), ".")) {
			return nil // hidden files not allowed
		}
		if f.Mode()&os.ModeSymlink != 0 {
			panic(errors.New("symlink not allowed in zip downloads")) // filepath.Walk doesnt support symlinks
		}

		header, err := zip.FileInfoHeader(f)
		check(err)
		header.Name = filepath.ToSlash(rel) // make the paths consistent between OSes
		header.Method = zip.Store
		headerWriter, err := zipWriter.CreateHeader(header)
		check(err)
		file, err := os.Open(path)
		check(err)
		defer file.Close()
		_, err = io.Copy(headerWriter, file)
		check(err)
		return nil
	})

	check(err)
}

func rpc(w http.ResponseWriter, r *http.Request) {
	var err error
	var rpc rpcCall
	defer exitPath(w, "rpc", &rpc)
	bodyBytes, err := io.ReadAll(r.Body)
	check(err)
	check(json.Unmarshal(bodyBytes, &rpc))
	ret := []byte("ok")

	switch rpc.Call {
	case "mkdirp":
		needArgs(rpc, 1)
		err = os.MkdirAll(enforcePath(rpc.Args[0]), os.ModePerm)
	case "mv":
		needArgs(rpc, 2)
		err = os.Rename(enforcePath(rpc.Args[0]), enforcePath(rpc.Args[1]))
	case "rm":
		needArgs(rpc, 1)
		err = os.RemoveAll(enforcePath(rpc.Args[0]))
	case "sum":
		needArgs(rpc, 2)
		file, err := os.Open(enforcePath(rpc.Args[0]))
		check(err)
		defer file.Close()
		var hash hash.Hash
		switch rpc.Args[1] {
		case "md5":
			hash = md5.New()
		case "sha1":
			hash = sha1.New()
		case "sha256":
			hash = sha256.New()
		case "sha512":
			hash = sha512.New()
		default:
			check(errors.New("unknown hash algorithm"))
		}
		_, err = io.Copy(hash, file)
		check(err)
		checksum := hash.Sum(nil)
		ret = make([]byte, hex.EncodedLen(len(checksum)))
		hex.Encode(ret, checksum)
	default:
		check(errors.New("unknown rpc call"))
	}

	check(err)
	w.Write(ret)
}

func enforcePath(p string) string {
	joined := filepath.Join(rootPath, strings.TrimPrefix(p, *extraPath))
	fp, err := filepath.Abs(joined)
	sl, _ := filepath.EvalSymlinks(fp) // err skipped as it would error for unexistent files (RPC check). The actual behaviour is tested below

	// panic if we had a error getting absolute path,
	// ... or if path doesnt contain the prefix path we expect,
	// ... or if we're skipping hidden folders, and one is requested,
	// ... or if we're skipping symlinks, path exists, and a symlink out of bound requested
	if err != nil || !within(fp) || *skipHidden && strings.Contains(p, "/.") || !*symlinks && len(sl) > 0 && !within(sl) {
		panic(errors.New("invalid path"))
	}

	return fp
}

func main() {
	if flag.Parse(); len(flag.Args()) == 1 {
		rootPath = flag.Args()[0]
	} else {
		fmt.Printf("\nusage: ./gossa [OPTIONS] ~/directory-to-share\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if (*tlsCert == "") != (*tlsKey == "") {
		fmt.Println("error: -tls-cert and -tls-key must be provided together")
		os.Exit(1)
	}

	var err error
	rootPath, err = filepath.Abs(rootPath)
	check(err)

	if !*ro {
		http.HandleFunc(*extraPath+"rpc", rpc)
		http.HandleFunc(*extraPath+"post", upload)
	}
	http.HandleFunc(*extraPath+"zip", zipRPC)
	http.HandleFunc("/", doContent)
	handler = http.StripPrefix(*extraPath, http.FileServer(http.Dir(rootPath)))

	var root http.Handler = http.DefaultServeMux
	if *auth != "" {
		root = basicAuth(root, *auth)
	}
	server := &http.Server{Addr: *host + ":" + *port, Handler: root}

	scheme := "http"
	if *tlsCert != "" {
		scheme = "https"
	}
	fmt.Printf("Gossa-SWS starting on directory %s\n", rootPath)
	fmt.Printf("Verbose: %t, Symlinks: %t, Read-Only: %t, Hidden-Files Skipped: %t, Calculate Folder Sizes: %t, Auth: %t\n",
		*verb, *symlinks, *ro, *skipHidden, *calcFolderSize, *auth != "")
	fmt.Printf("Listening on %s://%s:%s%s\n", scheme, *host, *port, *extraPath)

	if *tlsCert != "" {
		err = server.ListenAndServeTLS(*tlsCert, *tlsKey)
	} else {
		err = server.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		check(err)
	}
}

// basicAuth wraps next with HTTP basic authentication using constant-time
// comparison. creds is in "user:pass" form.
func basicAuth(next http.Handler, creds string) http.Handler {
	parts := strings.SplitN(creds, ":", 2)
	wantUser := parts[0]
	wantPass := ""
	if len(parts) == 2 {
		wantPass = parts[1]
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		userOk := subtle.ConstantTimeCompare([]byte(user), []byte(wantUser)) == 1
		passOk := subtle.ConstantTimeCompare([]byte(pass), []byte(wantPass)) == 1
		if !ok || !userOk || !passOk {
			w.Header().Set("WWW-Authenticate", `Basic realm="gossa"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
