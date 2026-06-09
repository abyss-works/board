import api from './index'
import type { Post, CreatePostRequest } from '@/types'

export async function fetchPosts(): Promise<Post[]> {
  const { data } = await api.get<Post[]>('/posts')
  return data
}

export async function fetchPost(id: number): Promise<Post> {
  const { data } = await api.get<Post>(`/posts/${id}`)
  return data
}

export async function createPost(req: CreatePostRequest): Promise<Post> {
  const { data } = await api.post<Post>('/posts', req)
  return data
}
