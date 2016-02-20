package main
import (
	"database/sql"
	"net/http"
	"errors"
	"strconv"
	"strings"
	_ "github.com/mattn/go-sqlite3"
	"net/url"
	"path/filepath"
)

var databases map[string]string = make(map[string]string)

// Add tile database
func RegisterTileDatabase(path string, name ...string) (err error) {
	if len(path) == 0 {
		err = errors.New("Path must not be empty")
		return
	}

	_, file := filepath.Split(path)
	databaseName := file
	if len(name) > 0 {
		databaseName = name[1]
	}

	databases[databaseName] = path
	return
}

func GetTile(w http.ResponseWriter, r *http.Request) {
	pathComponents, err := prepareRequest(r.URL)

	// TODO: Cleanup of error processing
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	x, y, z, err := getCoordinates(pathComponents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dsn, err, http_code := getDsn(pathComponents)
	if err != nil {
		http.Error(w, err.Error(), http_code)
		return
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		http.Error(w, "Database error (1)", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var tile_data *sql.RawBytes

	// TODO: Use .Query to properly handle "tile not found"
	err = db.QueryRow("SELECT tile_data FROM tiles WHERE zoom_level = ? AND tile_row = ? AND tile_column = ?", z, y, x).Scan(&tile_data)
	if err != nil {
		http.Error(w, "Database error (2)", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "image/png")
	w.Write(*tile_data)
}

func prepareRequest(requestUrl *url.URL) (pathComponents []string, err error) {
	pathComponents = strings.Split(requestUrl.Path[1:], "/")

	// expected: db, z, x, y
	if len(pathComponents) != 4 {
		err = errors.New("Unexpected number of path components")
	}

	return
}

func getDsn(pathComponents []string) (dsn string, err error, http_code int) {
	dbPath, ok := databases[pathComponents[0]]
	if !ok {
		err = errors.New("Unknown database")
		http_code = http.StatusNotFound
		return
	}

	dbUri, err := url.Parse("file:" + dbPath)
	if err != nil {
		err = errors.New("Bad database")
		http_code = http.StatusInternalServerError
		return
	}

	dsn = dbUri.String() + "?cache=shared&immutable=1"
	return
}

func getCoordinates(pathComponents []string) (x uint, y uint, z uint, err error) {
	indexOfPeriod := strings.LastIndex(pathComponents[3], ".")
	if indexOfPeriod != -1 {
		pathComponents[3] = pathComponents[3][:indexOfPeriod]
	}

	z64, err_z := strconv.ParseUint(pathComponents[1], 10, 0)
	x64, err_x := strconv.ParseUint(pathComponents[2], 10, 0)
	y64, err_y := strconv.ParseUint(pathComponents[3], 10, 0)

	if err_z != nil {
		err = err_z
	}

	if err_x != nil {
		err = err_x
	}

	if err_y != nil {
		err = err_y
	}

	if err_z != nil || err_x != nil || err_y != nil {
		return
	}

	// mbtiles format stores y axis reversed
	y64 = (1 << z64) - y64 - 1

	x = uint(x64)
	y = uint(y64)
	z = uint(z64)
	return
}