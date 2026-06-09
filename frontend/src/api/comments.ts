import api from './index'
import type { Comment, CreateCommentRequest } from '@/types'

export async function fetchComments(postId: number): Promise<Comment[]> {
  const { data } = await api.get<Comment[]>(`/posts/${postId}/comments`)
  return data
}

export async function createComment(postId: number, req: CreateCommentRequest): Promise<Comment> {
  const { data } = await api.post<Comment>(`/posts/${postId}/comments`, req)
  return data
}
