package main
import (
	"database/sql"
	"net/http"
	"errors"
	"fmt"
	"strconv"
	"strings"
	_ "github.com/mattn/go-sqlite3"
	"net/url"
	"path/filepath"
)

var databases map[string]string = make(map[string]string)

// Add tile database
func RegisterTileDatabase(path string, name string) (err error) {
	if len(path) == 0 {
		err = errors.New("Path must not be empty")
		return
	}

	if len(name) == 0 {
		_, file := filepath.Split(path)
		name = file
	}

	databases[name] = path
	return
}

func GetTile(w http.ResponseWriter, r *http.Request) {
	pathComponents := strings.Split(r.URL.Path[1:], "/")

	// expected: db, z, x, y
	if len(pathComponents) != 4 {
		http.NotFound(w, r)
		return
	}

	dbPath, ok := databases[pathComponents[0]]
	if !ok {
		http.NotFound(w, r)
		return
	}

	dbUri, err := url.Parse("file:" + dbPath)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	dsn := dbUri.String() + "?cache=shared&immutable=1"
	db, err := sql.Open("sqlite3", dsn)

	indexOfPeriod := strings.LastIndex(pathComponents[3], ".")
	if indexOfPeriod != -1 {
		pathComponents[3] = pathComponents[3][:indexOfPeriod]
	}

	z, errz := strconv.Atoi(pathComponents[1])
	x, errx := strconv.Atoi(pathComponents[2])
	y, erry := strconv.Atoi(pathComponents[3])

	if errz != nil || errx != nil || erry != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(
		w, "Would fetch tile (z, x, y) %v, %v, %v in database %s\n", z, x, y, pathComponents[0])
	fmt.Fprintf(w, "Database URI: %q", dbUri)
}