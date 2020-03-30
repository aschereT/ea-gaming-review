package db

import (
	"fmt"
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
			fmt.Println(p)
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
