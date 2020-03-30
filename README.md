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

`GET /blog/{id}/comment` -> get comments

`POST /blog/{id}/comment` -> add a comment

## Running

### Requirements

- `make`
- `docker`
- `docker-compose`
- `Insomnia`

### Steps

1. Execute `make`
2. Import `insomnia.json` into Insomnia
3. Interact with the API