# ea-gaming-review

![Docker Image CI](https://github.com/aschereT/ea-gaming-review/workflows/Docker%20Image%20CI/badge.svg)

## User Story

As an avid video game reviewer 

I want a way to create blog posts for my video game reviews 

So that I can share my reviews in a way that my readers can respond to

## Endpoints

GET /blog -> returns list of blog posts (pagination?)
POST /blog -> add a new posts

GET /blog/?id -> returns a blog post (without comments)
DELETE /blog/?id -> delete post (and its comments)

GET /blog/?id/comment -> get comments
POST /blog/?id/comment -> add a comment