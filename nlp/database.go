package nlp

import (
	"io/ioutil"
	"strings"

	cache "github.com/pmylund/go-cache"
)

const (
	DB_MAP = iota
	DB_PREFTREE
)

type Database struct {
	DBType  int
	dbmap   map[string]string
	dbptree *cache.Cache
}

func NewDatabase(t int) *Database {
	db := Database{
		DBType: t,
	}

	if t == DB_MAP {
		db.dbmap = make(map[string]string)
	} else if t == DB_PREFTREE {
		db.dbptree = cache.New(0, 0)
	}
	return &db
}

func NewDatabaseFromFile(dbFile string) *Database {
	db := Database{
		DBType: DB_MAP,
		dbmap:  make(map[string]string),
	}

	if dbFile != "" {
		filestr, err := ioutil.ReadFile(dbFile)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(filestr), "\n")
		if lines[0] == "DB_PREFTREE" {
			db.DBType = DB_PREFTREE
		}

		for i := 1; i < len(lines); i++ {
			line := lines[i]
			if line != "" {
				pos := strings.Index(line, " ")
				key := line[0:pos]
				data := line[pos+1:]
				db.addDatabase(key, data)
			}
		}
	}

	return &db
}

func (db *Database) addDatabase(key string, data string) {
	if db.DBType == DB_MAP {
		p := db.dbmap[key]
		if p != "" {
			db.dbmap[key] = p + " " + data
		} else {
			db.dbmap[key] = data
		}
	} else {
		_, found := db.dbptree.Get(key)
		if found {

		}
	}
}

func (db *Database) accessDatabase(key string) string {
	switch db.DBType {
	case DB_MAP:
		{
			p := db.dbmap[key]
			if p != "" {
				return p
			}
			break
		}
	case DB_PREFTREE:
		{
			//TODO
			break
		}
	default:
		break
	}

	return ""
}
