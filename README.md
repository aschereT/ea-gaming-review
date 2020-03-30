# ea-gaming-review

![Docker Image CI](https://github.com/aschereT/ea-gaming-review/workflows/Docker%20Image%20CI/badge.svg)
![Go](https://github.com/aschereT/ea-gaming-review/workflows/Go/badge.svg)

## User Story

As an avid video game reviewer 

I want a way to create blog posts for my video game reviews 

So that I can share my reviews in a way that my readers can respond to

## Acceptance Criteria

A blog post will show a title, article text (plain text) and an author name 

Comments are made on blog posts and show comment text (plain text) and an author name 

## Endpoints

`GET /blog` -> returns list of blog posts IDs (pagination?)

`POST /blog` -> add a new posts

`GET /blog/{id}` -> returns a blog post (without comments)

`DELETE /blog/{id}` -> delete post (and its comments)

`GET /blog/{id}/comment` -> get list of comment IDs

`GET /blog/{id}/comment/{commentid}` -> get a comment

`DELETE /blog/{id}/comment/{commentid}` -> delete a comment

`POST /blog/{id}/comment` -> add a comment

## Running

### From prebuilt image
Run `docker run -a STDOUT -a STDERR --rm -p 8080:8080 ascheret/easerver:latest`

See below for building locally

### Requirements

- `make`
- `docker`
- `docker-compose`
- `Insomnia`

### Steps

1. Execute `make`
2. Import `insomnia.json` into Insomnia
3. Interact with the API