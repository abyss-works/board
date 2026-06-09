import api from './index'
import type { Comment, CreateCommentRequest, UpdateCommentRequest } from '@/types'

export async function fetchComments(postId: number): Promise<Comment[]> {
  const { data } = await api.get<Comment[]>(`/posts/${postId}/comments`)
  return data
}

export async function createComment(postId: number, req: CreateCommentRequest): Promise<Comment> {
  const { data } = await api.post<Comment>(`/posts/${postId}/comments`, req)
  return data
}

export async function updateComment(commentId: number, req: UpdateCommentRequest): Promise<Comment> {
  const { data } = await api.put<Comment>(`/comments/${commentId}`, req)
  return data
}

export async function deleteComment(commentId: number, password: string): Promise<void> {
  await api.delete(`/comments/${commentId}`, { data: { password } })
}
