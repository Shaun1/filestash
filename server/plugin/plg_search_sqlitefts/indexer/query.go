package indexer

import (
	"database/sql"
	"path/filepath"
	"regexp"
	"time"

	. "github.com/mickael-kerjean/filestash/server/common"
)

func (this sqliteIndex) Search(path string, q string, timeRange SearchTimeRange) ([]IFile, error) {
	files := []IFile{}
	var (
		rows *sql.Rows
		err  error
	)

	if q == "" {
		rows, err = this.searchByTimeRange(path, timeRange)
	} else {
		rows, err = this.searchByKeyword(path, q, timeRange)
	}
	if err != nil {
		Log.Warning("search::query DBQuery (%s)", err.Error())
		return files, ErrNotReachable
	}
	defer rows.Close()

	for rows.Next() {
		f := File{}
		var t string
		if err = rows.Scan(&f.FType, &f.FPath, &f.FSize, &t); err != nil {
			Log.Warning("search::query scan (%s)", err.Error())
			return files, ErrNotReachable
		}
		if tm, err := time.Parse(time.RFC3339, t); err == nil {
			f.FTime = tm.Unix() * 1000
		}
		f.FName = filepath.Base(f.FPath)
		files = append(files, f)
	}
	return files, nil
}

func (this sqliteIndex) searchByKeyword(path string, q string, timeRange SearchTimeRange) (*sql.Rows, error) {
	query := "SELECT type, path, size, modTime FROM file WHERE path IN (" +
		"   SELECT path FROM file_index WHERE file_index MATCH ? AND path > ? AND path < ?" +
		"   ORDER BY rank LIMIT 50000" +
		")"
	args := []interface{}{
		regexp.MustCompile(`(\.|\-)`).ReplaceAllString(q, "\"$1\""),
		path, path + "~",
	}
	query, args = appendModTimeBounds(query, args, timeRange)
	return this.db.Query(query, args...)
}

func (this sqliteIndex) searchByTimeRange(path string, timeRange SearchTimeRange) (*sql.Rows, error) {
	query := "SELECT type, path, size, modTime FROM file WHERE path > ? AND path < ?"
	args := []interface{}{path, path + "~"}
	query, args = appendModTimeBounds(query, args, timeRange)
	query += " LIMIT 50000"
	return this.db.Query(query, args...)
}

func appendModTimeBounds(query string, args []interface{}, timeRange SearchTimeRange) (string, []interface{}) {
	if timeRange.From != nil {
		query += " AND modTime >= ?"
		args = append(args, time.UnixMilli(*timeRange.From))
	}
	if timeRange.To != nil {
		query += " AND modTime <= ?"
		args = append(args, time.UnixMilli(*timeRange.To))
	}
	return query, args
}
