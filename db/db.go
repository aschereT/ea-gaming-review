package db

import (
	"github.com/hashicorp/go-memdb"
	"github.com/segmentio/ksuid"
)

//BlogPost represents a single blog post
type BlogPost struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	ArticleText string `json:"ArticleText"`
	AuthorName  string `json:"AuthorName"`
}

//Comment represents a comment on a BlogPost
type Comment struct {
	ID          string `json:"ID"`
	ArticleID   string `json:"ArticleID"`
	CommentText string `json:"CommentText"`
	AuthorName  string `json:"AuthorName"`
}

const BlogPostTable = "BlogPost"
const CommentsTable = "Comments"

//InMemSchema is the schema for the in-memory database
var InMemSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		"BlogPost": &memdb.TableSchema{
			Name: BlogPostTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": &memdb.IndexSchema{
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "ID"},
				},
				"title": &memdb.IndexSchema{
					Name:    "title",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "Title"},
				},
				"articletext": &memdb.IndexSchema{
					Name:    "articletext",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "ArticleText"},
				},
				"authorname": &memdb.IndexSchema{
					Name:    "authorname",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "AuthorName"},
				},
			},
		},
		"Comments": &memdb.TableSchema{
			Name: CommentsTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": &memdb.IndexSchema{
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "ID"},
				},
				"articleid": &memdb.IndexSchema{
					Name:    "articleid",
					Unique:  false,
					Indexer: &memdb.UUIDFieldIndex{Field: "ArticleID"},
				},
				"commenttext": &memdb.IndexSchema{
					Name:    "commenttext",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "CommentText"},
				},
				"authorname": &memdb.IndexSchema{
					Name:    "authorname",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "AuthorName"},
				},
			},
		},
	},
}

func CreateDB() (*memdb.MemDB, error) {
	// Create a new data base
	db, err := memdb.NewMemDB(InMemSchema)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateBlogPost(inMemDB *memdb.MemDB, post BlogPost) (id string, err error) {
	txn := inMemDB.Txn(true)

	id = ksuid.New().String()
	post.ID = id
	err = txn.Insert(BlogPostTable, post)

	if err != nil {
		txn.Abort()
		return
	}

	txn.Commit()
	return
}

func GetBlogIDs(inMemDB *memdb.MemDB) (ids []string, err error) {
	txn := inMemDB.Txn(false)
	defer txn.Abort()

	it, err := txn.Get(BlogPostTable, "id")
	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(BlogPost)
		ids = append(ids, p.ID)
	}

	return ids, nil
}

func GetBlogPost(inMemDB *memdb.MemDB, id string) (post BlogPost, err error) {
	txn := inMemDB.Txn(false)
	defer txn.Abort()

	it, err := txn.Get(BlogPostTable, "id", id)
	if err != nil {
		return
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		post = obj.(BlogPost)
	}
	return
}
