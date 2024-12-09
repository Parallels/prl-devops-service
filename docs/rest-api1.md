---
layout: api
title: RestAPi
endpoints: 
  - title: dddd
    requires_authorization: true
    default_required_roles: 
      - admin
      - user
    path: /api/posts
    method: GET
    description: Get all posts
    parameters: 
      - name: id
        required: true
        type: string
    content_markdown: |
        # ffldfskdlf 
        dsdfds
    left_code_blocks: 
      - code_block: |
          {
            "id": "1",
            "title": "Post 1",
            "content": "Content 1",
            "category_id": "1",
            "user_id": "1",
            "created_at": "2021-01-01T00:00:00Z",
            "updated_at": "2021-01-01T00:00:00Z"
          }
        title: Json Response
        language: json
    right_code_blocks: 
      - code_block: |
          {
            "id": "1",
            "title": "Post 1",
            "content": "Content 1",
            "category_id": "1",
            "user_id": "1",
            "created_at": "2021-01-01T00:00:00Z",
            "updated_at": "2021-01-01T00:00:00Z"
          }
        title: Json Response
        language: json
  - title: Generates a token
    description: This endpoint generates a token
    requires_authorization: true
    path: /v1/auth/token [post]
    method: post
    left_code_blocks:
        - code_block: '{ "success": true }'
          title: Success
          language: json
        - code_block: '{ "failure": 401 }'
          title: Unauthorized
          language: json
        - code_block: '{ "bad": 400 }'
          title: BadRequest
          language: json
    right_code_blocks:
        - code_block: curl HTTP://localhost/api/v1/auth/token
          title: Success
          language: bash
  - title: dddd
    requires_authorization: true
    default_required_roles: 
      - admin
      - user
    path: /api/posts:id
    method: GET
    description: Get all posts
    parameters: 
      - name: id
        required: true
        type: string
    content_markdown: |
        # ffldfskdlf 
        dsdfds
    left_code_blocks: 
      - code_block: |
          {
            "id": "1",
            "title": "Post 1",
            "content": "Content 1",
            "category_id": "1",
            "user_id": "1",
            "created_at": "2021-01-01T00:00:00Z",
            "updated_at": "2021-01-01T00:00:00Z"
          }
        title: Json Response
    right_code_blocks: 
      - code_block: |
          {
            "id": "1",
            "title": "Post 1",
            "content": "Content 1",
            "category_id": "1",
            "user_id": "1",
            "created_at": "2021-01-01T00:00:00Z",
            "updated_at": "2021-01-01T00:00:00Z"
          }
        title: Json Response
  - name: Create a post
    url: /api/posts
    method: POST
    description: Create a post
  - name: Update a post
    url: /api/posts/:id
    method: PUT
    description: Update a post by id
  - name: Delete a post
    url: /api/posts/:id
    method: DELETE
    description: Delete a post by id
  - name: Get all categories
    url: /api/categories
    method: GET
    description: Get all categories
  - name: Get a category
    url: /api/categories/:id
    method: GET
    description: Get a category by id
  - name: Create a category
    url: /api/categories
    method: POST
    description: Create a category
  - name: Update a category
    url: /api/categories/:id
    method: PUT
    description: Update a category by id
  - name: Delete a category
    url: /api/categories/:id
    method: DELETE
    description: Delete a category by id
  - name: Get all tags
    url: /api/tags
    method: GET
    description: Get all tags
  - name: Get a tag
    url: /api/tags/:id
    method: GET
    description: Get a tag by id
  - name: Create a tag
    url: /api/tags
    method: POST
    description: Create a tag
  - name: Update a tag
    url: /api/tags/:id
    method: PUT
    description: Update a tag by id
  - name: Delete a tag
    url: /api/tags/:id
    method: DELETE
    description: Delete a tag by id
  - name: Get all users
    url: /api/users
    method: GET
    description: Get all users
  - name: Get a user
    url: /api/users/:id
    method: GET
    description: Get a user by id
---

# someteds

s
sadaslkdasldkas
askdals
