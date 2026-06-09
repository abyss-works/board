import api from './index'
import type { Post, CreatePostRequest, UpdatePostRequest } from '@/types'

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

export async function updatePost(id: number, req: UpdatePostRequest): Promise<Post> {
  const { data } = await api.put<Post>(`/posts/${id}`, req)
  return data
}

export async function deletePost(id: number): Promise<void> {
  await api.delete(`/posts/${id}`)
}
