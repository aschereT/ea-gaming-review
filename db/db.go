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
					Indexer: &memdb.StringFieldIndex{Field: "ArticleID"},
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

//parent should already grab a transaction handler already
func getBlogPostWithTxn(txn *memdb.Txn, articleID string) (post *BlogPost, err error) {
	foundObj, err := txn.First(BlogPostTable, "id", articleID)
	if err != nil {
		return nil, err
	}
	if foundObj == nil {
		return nil, nil
	}

	foundPost := foundObj.(BlogPost)
	return &foundPost, nil
}

//parent should already grab a transaction handler already
func getBlogCommentWithTxn(txn *memdb.Txn, commentID string) (comment *BlogComment, err error) {
	foundObj, err := txn.First(CommentsTable, "id", commentID)
	if err != nil {
		return nil, err
	}
	if foundObj == nil {
		return nil, nil
	}

	foundComment := foundObj.(BlogComment)
	return &foundComment, nil
}

//parent should already grab a transaction handler already
func getBlogCommentIDsWithTxn(txn *memdb.Txn, articleID string) (ids []string, err error) {
	post, err := getBlogPostWithTxn(txn, articleID)
	if err != nil {
		return nil, err
	}
	if post == nil {
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

//parent should already grab a transaction handler already
func deleteBlogCommentIDsWithTxn(txn *memdb.Txn, articleID, commentID string) (exists bool, err error) {
	blogPost, err := getBlogPostWithTxn(txn, articleID)
	if err != nil {
		return false, err
	}
	if blogPost == nil {
		return false, fmt.Errorf("No such blog post")
	}

	toDeleteObject, err := getBlogCommentWithTxn(txn, commentID)
	if err != nil {
		return false, err
	}
	if toDeleteObject == nil {
		return false, fmt.Errorf("No such comment")
	}

	err = txn.Delete(CommentsTable, toDeleteObject)
	if err != nil {
		return true, err
	}

	return true, nil
}

//Inserts a new post, generating a unique ID for it and returning that
func CreateBlogPost(inMemDB *memdb.MemDB, post BlogPost) (id string, err error) {
	txn := inMemDB.Txn(true)
	defer txn.Abort()

	id = ksuid.New().String()
	post.ID = id
	err = txn.Insert(BlogPostTable, post)

	if err != nil {
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

	return getBlogPostWithTxn(txn, id)
}

//Deletes a single post and its attendant comments. exists indicates if err is 404 or something else
func DeleteBlogPost(inMemDB *memdb.MemDB, articleID string) (exists bool, err error) {
	txn := inMemDB.Txn(true)
	defer txn.Abort()

	toDeleteObject, err := getBlogPostWithTxn(txn, articleID)
	if err != nil {
		return false, err
	}
	if toDeleteObject == nil {
		return false, nil
	}

	err = txn.Delete(BlogPostTable, toDeleteObject)
	if err != nil {
		return true, err
	}

	commentsToDelete, err := getBlogCommentIDsWithTxn(txn, articleID)
	if err != nil {
		return true, err
	}

	for _, commentID := range commentsToDelete {
		exists, err := deleteBlogCommentIDsWithTxn(txn, articleID, commentID)
		if err != nil {
			return true, err
		}
		if !exists {
			return true, fmt.Errorf("Error deleting comment %s for blog post %s", commentID, articleID)
		}
	}
	txn.Commit()
	return true, nil
}

//Returns a list of all comment IDs on the given articleID
//TODO: pagination?
func GetCommentIDs(inMemDB *memdb.MemDB, articleID string) (ids []string, err error) {
	txn := inMemDB.Txn(false)
	defer txn.Abort()

	return getBlogCommentIDsWithTxn(txn, articleID)
}

//Gets a single comment. comment is nil if such comment is not found
func GetBlogComment(inMemDB *memdb.MemDB, articleID, id string) (comment *BlogComment, err error) {
	txn := inMemDB.Txn(false)
	defer txn.Abort()

	post, err := getBlogPostWithTxn(txn, articleID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, fmt.Errorf("No such blog post")
	}

	return getBlogCommentWithTxn(txn, id)
}

//Inserts a new comment, generating a unique ID for it and returning that
func CreateBlogComment(inMemDB *memdb.MemDB, comment BlogComment) (id string, err error) {
	txn := inMemDB.Txn(true)
	defer txn.Abort()

	post, err := getBlogPostWithTxn(txn, comment.ArticleID)
	if err != nil {
		return "", err
	}
	if post == nil {
		return "", fmt.Errorf("No such blog post")
	}

	id = ksuid.New().String()
	comment.ID = id
	err = txn.Insert(CommentsTable, comment)

	if err != nil {
		return "", err
	}

	txn.Commit()
	return id, nil
}

//Deletes a single comment. exists indicates if err is 404 or something else
func DeleteBlogComment(inMemDB *memdb.MemDB, articleID, commentID string) (exists bool, err error) {
	txn := inMemDB.Txn(true)
	defer txn.Abort()

	exists, err = deleteBlogCommentIDsWithTxn(txn, articleID, commentID)
	if err != nil {
		return exists, err
	}
	if !exists {
		return exists, err
	}
	txn.Commit()
	return exists, err
}