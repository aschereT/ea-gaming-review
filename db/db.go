package db

import (
	"fmt"

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

//BlogComment represents a comment on a BlogPost
type BlogComment struct {
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

//Initialise the in-memory database
func CreateDB() (*memdb.MemDB, error) {
	// Create a new data base
	db, err := memdb.NewMemDB(InMemSchema)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func checkBlogPostExists(txn *memdb.Txn, articleID string) (exists bool, err error) {
	//check if blog post exists
	foundObj, err := txn.First(BlogPostTable, "id", articleID)
	if err != nil {
		return false, err
	}
	if foundObj == nil {
		return false, nil
	}
	return true, nil
}

//Inserts a new post, generating a unique ID for it and returning that
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

	return id, nil
}

//Returns a list of all blog IDs
//TODO: pagination?
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

//Gets a single post. post is nil if such post is not found
func GetBlogPost(inMemDB *memdb.MemDB, id string) (post *BlogPost, err error) {
	txn := inMemDB.Txn(false)
	defer txn.Abort()

	foundObj, err := txn.First(BlogPostTable, "id", id)
	if err != nil {
		return nil, err
	}
	if foundObj == nil {
		return nil, nil
	}

	foundPost := foundObj.(BlogPost)
	return &foundPost, nil
}

//Deletes a single post. exists indicates if err is 404 or something else
func DeleteBlogPost(inMemDB *memdb.MemDB, id string) (exists bool, err error) {
	toDeleteObject, err := GetBlogPost(inMemDB, id)
	if err != nil {
		return false, err
	}
	if toDeleteObject == nil {
		return false, nil
	}

	txn := inMemDB.Txn(true)
	err = txn.Delete(BlogPostTable, toDeleteObject)
	if err != nil {
		txn.Abort()
		return true, err
	}

	txn.Commit()
	return true, nil
}

//Inserts a new comment, generating a unique ID for it and returning that
func CreateBlogComment(inMemDB *memdb.MemDB, comment BlogComment) (id string, err error) {
	txn := inMemDB.Txn(true)

	exists, err := checkBlogPostExists(txn, comment.ArticleID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("No such blog post")
	}

	id = ksuid.New().String()
	comment.ID = id
	err = txn.Insert(CommentsTable, comment)

	if err != nil {
		txn.Abort()
		return "", err
	}

	txn.Commit()
	return id, nil
}

//Returns a list of all blog IDs
//TODO: pagination?
func GetCommentIDs(inMemDB *memdb.MemDB, articleID string) (ids []string, err error) {
	txn := inMemDB.Txn(false)
	defer txn.Abort()

	exists, err := checkBlogPostExists(txn, articleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("No such blog post")
	}

	it, err := txn.Get(CommentsTable, "articleid", articleID)
	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(BlogComment)
		ids = append(ids, p.ID)
	}

	return ids, nil
}