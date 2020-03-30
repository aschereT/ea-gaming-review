package db

import (
	"reflect"
	"sort"
	"testing"
)

func Test_Schema(t *testing.T) {
	err := InMemSchema.Validate()
	if err != nil {
		t.Error(err)
	}
}

func Test_CreateDB(t *testing.T) {
	_, err := CreateDB()
	if err != nil {
		t.Error(err)
	}
}

func Test_CreateBlogPost(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	expected := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	for i := range expected {
		id, err := CreateBlogPost(db, expected[i])
		if err != nil {
			t.Error(err)
		}
		expected[i].ID = id
	}

	txn := db.Txn(false)
	defer txn.Abort()
	for i := range expected {
		it, err := txn.Get(BlogPostTable, "id", expected[i].ID)
		if err != nil {
			t.Error(err)
		}

		for obj := it.Next(); obj != nil; obj = it.Next() {
			p := obj.(BlogPost)
			if p != expected[i] {
				t.Errorf("Expected created post to be %#v, got %#v", expected[i], p)
			}
		}
	}
}

func Test_GetBlogIDs(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	sampleBlogPosts := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	expected := []string{}
	for i := range sampleBlogPosts {
		id, err := CreateBlogPost(db, sampleBlogPosts[i])
		if err != nil {
			t.Error(err)
		}
		expected = append(expected, id)
	}

	actual, err := GetBlogIDs(db)
	if err != nil {
		t.Error(err)
	}

	if len(expected) != len(actual) {
		t.Errorf("Expected %d IDs, got %d IDs", len(expected), len(actual))
	}

	sort.Strings(expected)
	sort.Strings(actual)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected IDs %#v, got %#v", expected, actual)
	}
}

func Test_GetBlogIDs_NoPosts(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	actual, err := GetBlogIDs(db)
	if err != nil {
		t.Error(err)
	}

	expected := []string(nil)
	if len(expected) != len(actual) {
		t.Errorf("Expected %d IDs, got %d IDs", len(expected), len(actual))
	}

	sort.Strings(actual)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected IDs %#v, got %#v", expected, actual)
	}
}

func Test_GetBlogPost(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	expected := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	for i := range expected {
		id, err := CreateBlogPost(db, expected[i])
		if err != nil {
			t.Error(err)
		}
		expected[i].ID = id

		post, err := GetBlogPost(db, id)
		if err != nil {
			t.Error(err)
		}
		if *post != expected[i] {
			t.Errorf("Expected created post to be %#v, got %#v", expected[i], *post)
		}
	}
}

func Test_GetBlogPost_Nonexistent(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	post, err := GetBlogPost(db, "this_id_doesnt_exist")
	if err != nil {
		t.Error(err)
	}
	if post != nil {
		t.Errorf("Expected post to not exist, got %#v", *post)
	}
}

func Test_DeleteBlogPost(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	expected := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	for i := range expected {
		id, err := CreateBlogPost(db, expected[i])
		if err != nil {
			t.Error(err)
		}
		expected[i].ID = id
	}

	for i := range expected {
		ex, err := DeleteBlogPost(db, expected[i].ID)
		if err != nil {
			t.Error(err)
		}
		if !ex {
			t.Errorf("Expected post to have existed, got %v", ex)
		}
	}

	for i := range expected {
		txn := db.Txn(false)
		defer txn.Abort()

		post, err := txn.First(BlogPostTable, "id", expected[i].ID)
		if err != nil {
			t.Error(err)
		}
		if post != nil {
			t.Errorf("Expected post to be deleted, but still exists %#v", post)
		}
	}
}

func Test_CreateBlogComment(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	blogPosts := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	for i := range blogPosts {
		id, err := CreateBlogPost(db, blogPosts[i])
		if err != nil {
			t.Error(err)
		}
		blogPosts[i].ID = id
	}

	expectedComments := []BlogComment{}
	for i := range blogPosts {
		expectedComments = append(expectedComments, BlogComment{ArticleID: blogPosts[i].ID, AuthorName: "firstposter", CommentText: "First!"},
			BlogComment{ArticleID: blogPosts[i].ID, AuthorName: "spammer", CommentText: "Learn how to get hired with this one weird trick!"})
	}

	for i, comment := range expectedComments {
		id, err := CreateBlogComment(db, comment)
		if err != nil {
			t.Error(err)
		}
		expectedComments[i].ID = id
	}

	txn := db.Txn(false)
	defer txn.Abort()
	for i := range expectedComments {
		it, err := txn.Get(CommentsTable, "id", expectedComments[i].ID)
		if err != nil {
			t.Error(err)
		}

		for obj := it.Next(); obj != nil; obj = it.Next() {
			p := obj.(BlogComment)
			if p != expectedComments[i] {
				t.Errorf("Expected created comment to be %#v, got %#v", blogPosts[i], p)
			}
		}
	}
}

func Test_GetBlogComment(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	blogPosts := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	for i := range blogPosts {
		id, err := CreateBlogPost(db, blogPosts[i])
		if err != nil {
			t.Error(err)
		}
		blogPosts[i].ID = id
	}

	expectedComments := []BlogComment{}
	for i := range blogPosts {
		expectedComments = append(expectedComments, BlogComment{ArticleID: blogPosts[i].ID, AuthorName: "firstposter", CommentText: "First!"},
			BlogComment{ArticleID: blogPosts[i].ID, AuthorName: "spammer", CommentText: "Learn how to get hired with this one weird trick!"})
	}

	for i, comment := range expectedComments {
		id, err := CreateBlogComment(db, comment)
		if err != nil {
			t.Error(err)
		}
		expectedComments[i].ID = id
	}

	for _, comment := range expectedComments {
		actualComment, err := GetBlogComment(db, comment.ArticleID, comment.ID)
		if err != nil {
			t.Error(err)
		}
		if comment != *actualComment {
			t.Errorf("Expected returned comment to be %#v, got %#v", comment, actualComment)
		}
	}
}

func Test_GetBlogCommentIDs(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	blogPosts := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	for i := range blogPosts {
		id, err := CreateBlogPost(db, blogPosts[i])
		if err != nil {
			t.Error(err)
		}
		blogPosts[i].ID = id
	}

	expectedComments := []BlogComment{}
	for i := range blogPosts {
		expectedComments = append(expectedComments, BlogComment{ArticleID: blogPosts[i].ID, AuthorName: "firstposter", CommentText: "First!"},
			BlogComment{ArticleID: blogPosts[i].ID, AuthorName: "spammer", CommentText: "Learn how to get hired with this one weird trick!"})
	}

	for i, comment := range expectedComments {
		id, err := CreateBlogComment(db, comment)
		if err != nil {
			t.Error(err)
		}
		expectedComments[i].ID = id
	}

	for _, post := range blogPosts {
		articleID := post.ID
		actualCommentIDs, err := GetCommentIDs(db, articleID)
		if err != nil {
			t.Error(err)
		}

		expectedCommentIDs := func(comments []BlogComment, articleID string) (result []string) {
			for _, comment := range comments {
				if comment.ArticleID == articleID {
					result = append(result, comment.ID)
				}
			}
			return result
		}(expectedComments, articleID)

		if len(expectedCommentIDs) != len(actualCommentIDs) {
			t.Errorf("Expected %d IDs, got %d IDs", len(expectedCommentIDs), len(actualCommentIDs))
		}

		sort.Strings(expectedCommentIDs)
		sort.Strings(actualCommentIDs)
		if !reflect.DeepEqual(expectedCommentIDs, actualCommentIDs) {
			t.Errorf("Expected IDs %#v, got %#v", expectedCommentIDs, actualCommentIDs)
		}
	}

}
func Test_DeleteBlogComment(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	blogPosts := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	for i := range blogPosts {
		id, err := CreateBlogPost(db, blogPosts[i])
		if err != nil {
			t.Error(err)
		}
		blogPosts[i].ID = id
	}

	expectedComments := []BlogComment{}
	for i := range blogPosts {
		expectedComments = append(expectedComments, BlogComment{ArticleID: blogPosts[i].ID, AuthorName: "firstposter", CommentText: "First!"},
			BlogComment{ArticleID: blogPosts[i].ID, AuthorName: "spammer", CommentText: "Learn how to get hired with this one weird trick!"})
	}

	for i, comment := range expectedComments {
		id, err := CreateBlogComment(db, comment)
		if err != nil {
			t.Error(err)
		}
		expectedComments[i].ID = id
	}

	for _, comment := range expectedComments {
		exists, err := DeleteBlogComment(db, comment.ArticleID, comment.ID)
		if err != nil {
			t.Error(err)
		}
		if !exists {
			t.Errorf("Expected comment to be deleted to have existed, got %v", exists)
		}
	}

	for _, comment := range expectedComments {
		actualComment, err := GetBlogComment(db, comment.ArticleID, comment.ID)
		if err != nil {
			t.Error(err)
		}
		if actualComment != nil {
			t.Errorf("Expected comment to have been deleted, got %#v", *actualComment)
		}
	}
}

func Test_DeleteBlogPost_WithComments(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		t.Error(err)
	}

	expectedBlogPosts := []BlogPost{BlogPost{Title: "Test Title 1", ArticleText: "Test Body 1", AuthorName: "Test Author Name 1"}, BlogPost{Title: "Test Title 2", ArticleText: "Test Body 2", AuthorName: "Test Author Name 2"}}
	for i := range expectedBlogPosts {
		id, err := CreateBlogPost(db, expectedBlogPosts[i])
		if err != nil {
			t.Error(err)
		}
		expectedBlogPosts[i].ID = id
	}

	expectedComments := []BlogComment{}
	for _, post := range expectedBlogPosts {
		expectedComments = append(expectedComments, BlogComment{ArticleID: post.ID, AuthorName: "firstposter", CommentText: "First!"},
			BlogComment{ArticleID: post.ID, AuthorName: "spammer", CommentText: "Learn how to get hired with this one weird trick!"})
	}

	for i, comment := range expectedComments {
		id, err := CreateBlogComment(db, comment)
		if err != nil {
			t.Error(err)
		}
		expectedComments[i].ID = id
	}

	for i := range expectedBlogPosts {
		ex, err := DeleteBlogPost(db, expectedBlogPosts[i].ID)
		if err != nil {
			t.Error(err)
		}
		if !ex {
			t.Errorf("Expected post to have existed, got %v", ex)
		}
	}

	for i := range expectedBlogPosts {
		txn := db.Txn(false)
		defer txn.Abort()

		post, err := txn.First(BlogPostTable, "id", expectedBlogPosts[i].ID)
		if err != nil {
			t.Error(err)
		}
		if post != nil {
			t.Errorf("Expected post to be deleted, but still exists %#v", post)
		}
	}

	for _, comment := range expectedComments {
		actualComment, err := GetBlogComment(db, comment.ArticleID, comment.ID)
		if err == nil {
			t.Errorf("Expected error when getting comments whose parent blog post have been deleted, %#v", comment)
		}
		if actualComment != nil {
			t.Errorf("Expected comment to have been deleted, got %#v", *actualComment)
		}
	}

	txn := db.Txn(false)
	defer txn.Abort()
	for _, comment := range expectedComments {
		actualComment, err := txn.First(CommentsTable, "id", comment.ID)
		if err != nil {
			t.Error(err)
		}

		if actualComment != nil {
			ac := actualComment.(BlogComment)
			t.Errorf("Expected comments to have been deleted, got %#v", ac)
		}
	}
}
