export interface Post {
  id: number
  title: string
  content: string
  author: string
  createdAt: string
}

export interface Comment {
  id: number
  postId: number
  content: string
  author: string
  createdAt: string
  updatedAt?: string
}

export interface CreatePostRequest {
  title: string
  content: string
  password?: string
}

export interface UpdatePostRequest {
  title: string
  content: string
  password: string
}

export interface CreateCommentRequest {
  content: string
  password?: string
}

export interface UpdateCommentRequest {
  content: string
  password: string
}
